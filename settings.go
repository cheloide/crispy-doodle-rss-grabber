package main

//Settings ...
type Settings struct {
	DBPath string `json:"dbPath"`
	Feeds  []Feed `json:"feeds"`
}

//Feed ...
type Feed struct {
	Name       string        `json:"name"`
	RssURL     string        `json:"rssUrl"`
	Command    Command       `json:"command"`
	Rules      []RssItemRule `json:"rules"`
	BucketName string        `json:"bucketName"`
	Key        string        `json:"key"`
}

//Command ..
type Command struct {
	Executable        string            `json:"executable"`
	ArgumentTemplates []string          `json:"argumentTemplates"`
	Variables         map[string]string `json:"variables"`
}

//RssItemRule ...
type RssItemRule struct {
	Operator             string   `json:"operator"`
	RssItemField         string   `json:"rssItemField"`
	Negate               bool     `json:"negate"`
	Requirement          string   `json:"requirement"`
	Equals               []string `json:"equals"`
	Contains             []string `json:"contains"`
	StartsWith           []string `json:"startsWith"`
	EndsWith             []string `json:"endsWith"`
	EqualsIgnoreCase     []string `json:"equalsIgnoreCase"`
	ContainsIgnoreCase   []string `json:"containsIgnoreCase"`
	StartsWithIgnoreCase []string `json:"startsWithIgnoreCase"`
	EndsWithIgnoreCase   []string `json:"endsWithIgnoreCase"`
}
