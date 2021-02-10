package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

const defaultSettingsPath = "settings.json"

var database *bolt.DB
var regexArgument = regexp.MustCompile(`\${ARG\.(.*?)}`)
var regexKey = regexp.MustCompile(`\${(.*?)\.(.*?)}`)

func main() {

	fmt.Printf("Process Started (%s)\n", time.Now().Format(time.RFC3339))

	var settingsHash string
	var settings Settings
	var settingsPath string
	if len(os.Args) > 1 {
		settingsPath = os.Args[1]
		fmt.Println("Custom configuration path argument found ", settingsPath)
	} else {
		settingsPath = defaultSettingsPath
	}

	if err := getSettings(settingsPath, &settingsHash, &settings); err != nil {
		log.Fatal(err)
	}

	db, err := bolt.Open(settings.DBPath, 0644, nil)
	if err != nil {
		log.Fatal(err)
	}

	database = db
	// processRSS(&settingsHash, &settings)
	fmt.Printf("Finished Successfully (%s)\n", time.Now().Format(time.RFC3339))

}

func processRSS(settingsHash *string, settings *Settings) {

	for i := 0; i < len(settings.Feeds); i++ {
		var feed = settings.Feeds[i]
		fmt.Printf("Processing RSS Feed from %s\n", feed.RssURL)
		rssFeed, err := getRssFeedFromURL(feed.RssURL)
		if err != nil {
			log.Printf("Fail to get RSS Feed from %s\n", feed.RssURL)
			continue
		}
		processRSSFeed(&rssFeed, settings, i)
	}
}
func processRSSFeed(rssFeed *RssFeed, settings *Settings, feedIndex int) {
	var arguments [][]string
	var feedSettings = settings.Feeds[feedIndex]
	var bucketKeyPairs [][]string
	var itemTitles []string
	var argumentTemplates = replaceCommandArguments(feedSettings.Command.ArgumentTemplates, feedSettings.Command.Variables)
	for i := 0; i < len(rssFeed.Channel.Item); i++ {
		var item = rssFeed.Channel.Item[i]
		if validateRules(item, rssFeed, feedSettings.Rules) {
			newArg := replaceCommandVariables(item, rssFeed, argumentTemplates)
			arguments = append(arguments, newArg)
			itemTitles = append(itemTitles, item.Title)
			bucketKeyPairs = append(bucketKeyPairs, []string{templateReplace(item, rssFeed, feedSettings.BucketName), templateReplace(item, rssFeed, feedSettings.Key)})
		}
	}
	runCommands(feedSettings.Command.Executable, arguments, bucketKeyPairs, itemTitles)
}
func replaceCommandArguments(argumentTemplates []string, arguments map[string]string) []string {

	var newArgumentTemplates []string = argumentTemplates

	for i := 0; i < len(newArgumentTemplates); i++ {
		a := newArgumentTemplates[i]
		for placeholders := regexArgument.FindStringSubmatch(a); placeholders != nil; placeholders = regexArgument.FindStringSubmatch(a) {
			newArgumentTemplates[i] = strings.ReplaceAll(a, placeholders[0], arguments[placeholders[1]])
			a = newArgumentTemplates[i]
		}
	}
	return newArgumentTemplates
}

func replaceCommandVariables(item Item, rssFeed *RssFeed, argumentTemplates []string) []string {

	var arguments []string = make([]string, len(argumentTemplates))
	copy(arguments, argumentTemplates)

	for i := 0; i < len(arguments); i++ {
		arguments[i] = templateReplace(item, rssFeed, arguments[i])

	}
	return arguments
}

func replaceItemKey(item Item, rssFeed *RssFeed, keyString string) string {

	return templateReplace(item, rssFeed, keyString)
}

func templateReplace(item Item, rssFeed *RssFeed, template string) string {

	var replaced string = template
	for placeholders := regexKey.FindStringSubmatch(replaced); placeholders != nil; placeholders = regexKey.FindStringSubmatch(replaced) {

		var value string
		var fieldName = strings.Title(placeholders[2])

		if placeholders[1] == "RSSROOT" {
			value = fmt.Sprint(reflect.ValueOf(&rssFeed).Elem().FieldByName(fieldName).Interface())
		} else if placeholders[1] == "RSSITEM" {
			value = fmt.Sprint(reflect.ValueOf(&item).Elem().FieldByName(fieldName).Interface())
		}
		replaced = strings.ReplaceAll(replaced, placeholders[0], value)
	}
	return replaced
}

func runCommands(executable string, arguments [][]string, bucketKeyPair [][]string, itemTitle []string) {
	for i := 0; i < len(arguments); i++ {
		if !keyExists(bucketKeyPair[i][0], bucketKeyPair[i][1]) {

			fmt.Println("Item:", itemTitle[i])
			fmt.Println("Running", executable, arguments[i])

			cmd := exec.Command(executable, arguments[i]...)

			var out bytes.Buffer
			var stderr bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &stderr
			err := cmd.Run()
			if err != nil {
				log.Println(fmt.Sprint(err) + ": " + stderr.String())
				continue
			}
			keyWrite(bucketKeyPair[i][0], bucketKeyPair[i][1])
			fmt.Println("Result: " + out.String())

		} else {
			log.Printf("Bucket/Key %s %s exists for item \"%s\"", bucketKeyPair[i][0], bucketKeyPair[i][1], itemTitle[i])
		}
	}
}

// func runCommand(item Item, rssFeed *RssFeed, command string, arguments map[string]string) {})
// func dispatchCommands()

func validateRules(item Item, rssFeed *RssFeed, rules []RssItemRule) bool {
	var conditionals []bool

	for i := 0; i < len(rules); i++ {
		rule := rules[i]
		if conditionals == nil {
			conditionals = make([]bool, 1)
			conditionals[0] = validateRule(item, rule)
		} else if rule.Operator == "OR" {
			if conditionals[len(conditionals)-1] {
				return true
			}
			conditionals = append(conditionals, validateRule(item, rule))
		} else {
			var currentValue = conditionals[len(conditionals)-1]
			conditionals[len(conditionals)-1] = currentValue && validateRule(item, rule)
		}
	}
	for i := 0; i < len(conditionals); i++ {
		if conditionals[i] {
			return true
		}
	}
	return false
}
func validateRule(item Item, rule RssItemRule) bool {
	var any = strings.EqualFold(rule.Requirement, "ANY")
	var negate = rule.Negate
	var fieldName = rule.RssItemField

	e := reflect.ValueOf(&item).Elem()
	fieldValue := fmt.Sprint(e.FieldByName(strings.Title(fieldName)).Interface())

	var equals = validateRuleFragment(fieldValue, rule.Equals, stringEquals, false, any)
	var contains = validateRuleFragment(fieldValue, rule.Contains, strings.Contains, false, any)
	var startsWith = validateRuleFragment(fieldValue, rule.StartsWith, strings.HasPrefix, false, any)
	var endsWith = validateRuleFragment(fieldValue, rule.EndsWith, strings.HasSuffix, false, any)
	var equalsIgnoreCase = validateRuleFragment(fieldValue, rule.EqualsIgnoreCase, strings.EqualFold, false, any)
	var containsIgnoreCase = validateRuleFragment(fieldValue, rule.ContainsIgnoreCase, strings.Contains, true, any)
	var startsWithIgnoreCase = validateRuleFragment(fieldValue, rule.StartsWithIgnoreCase, strings.HasPrefix, true, any)
	var endsWithIgnoreCase = validateRuleFragment(fieldValue, rule.EndsWithIgnoreCase, strings.HasSuffix, true, any)
	if any {
		return negateOrDefault(negate, equals || contains || startsWith || endsWith || equalsIgnoreCase || containsIgnoreCase || startsWithIgnoreCase || endsWithIgnoreCase)
	}
	return negateOrDefault(negate, (equals && contains && startsWith && endsWith && equalsIgnoreCase && containsIgnoreCase && startsWithIgnoreCase && endsWithIgnoreCase))
}
func negateOrDefault(negate bool, value bool) bool {
	if negate {
		return !value
	}
	return value
}
func stringEquals(a string, b string) bool { return a == b }

func validateRuleFragment(valueItem string, rules []string, comparator func(a, b string) bool, ignoreCase bool, any bool) bool {

	if !any && len(rules) == 0 {
		return true
	}

	for i := 0; i < len(rules); i++ {
		if comparator(upperCaseOrDefault(ignoreCase, valueItem), upperCaseOrDefault(ignoreCase, rules[i])) {
			return true
		}
	}
	return false
}

func upperCaseOrDefault(uppercase bool, value string) string {
	if uppercase {
		return strings.ToUpper(value)
	}
	return value

}

func getRssFeedFromURL(url string) (RssFeed, error) {
	var rssFeed RssFeed
	resp, err := http.Get(url)
	if err != nil {
		log.Print(err)
		return rssFeed, err
	}

	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
	}
	err = xml.Unmarshal(responseBody, &rssFeed)
	if err != nil {
		log.Print(err)
		return rssFeed, err
	}
	return rssFeed, nil
}

func getXML(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, fmt.Errorf("GET error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf("Status error: %v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("Read body: %v", err)
	}

	return data, nil
}

func getSettings(settingsPath string, settingsHash *string, settings *Settings) error {

	file, err := readFile(settingsPath)
	if err != nil {
		return err
	}
	hash, err := getShaHashFromFile(file)
	if err != nil {
		return err
	}

	defer file.Close()

	if &settingsHash == nil || *settingsHash != hash {
		*settingsHash = hash
		var err error
		s, err := readJSONSettings(settingsPath)
		if err != nil {
			return err
		}
		*settings = s
	}

	return nil
}

func readFile(path string) (*os.File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func getShaHashFromFile(f *os.File) (string, error) {
	var hashString string

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Print(err)
		return hashString, err
	}
	hashString = fmt.Sprintf("%x", h.Sum(nil))
	return hashString, nil

}
func readJSONSettings(path string) (Settings, error) {
	var settings Settings

	rawJSONData, ioErr := ioutil.ReadFile(path)

	if ioErr != nil {
		log.Print(ioErr)
		return settings, ioErr
	}

	var unMarshalErr = json.Unmarshal([]byte(rawJSONData), &settings)

	if unMarshalErr != nil {
		log.Print(unMarshalErr)
		return settings, unMarshalErr
	}
	return settings, nil
}

func keyExists(bucketName string, key string) bool {
	var exists bool = false
	database.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			log.Printf("Bucket %s not found!\n", bucketName)
			return nil
		}
		v := b.Get([]byte(key))
		exists = fmt.Sprintf("%s", v) == "1"
		return nil
	})
	return exists
}

func keyWrite(bucketName string, key string) {
	err := database.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			log.Println("Error creating Bucket", err)
			return err
		}
		err = b.Put([]byte(key), []byte("1"))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Println("Error accessing DB", err)
	}
}
