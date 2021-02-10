package main

//RssFeed ...
type RssFeed struct {
	Channel Channel `xml:"channel"`
}

//Channel ...
type Channel struct {
	//The name of the channel. It's how people refer to your service. If you have an HTML website that contains the same information as your RSS file, the title of your channel should be the same as the title of your website.
	Title string `xml:"title"`
	//The URL to the HTML website corresponding to the channel.
	Link string `xml:"link"`
	//Phrase or sentence describing the channel.
	Description string `xml:"description"`
	//The language the channel is written in. This allows aggregators to group all Italian language sites, for example, on a single page. A list of allowable values for this element, as provided by Netscape, is here. You may also use values defined by the W3C. 	en-us
	Language string `xml:"language"`
	//Copyright notice for content in the channel. 	Copyright 2002, Spartanburg Herald-Journal
	Copyright string `xml:"copyright"`
	//Email address for person responsible for editorial content. 	geo@herald.com (George Matesky)
	ManagingEditor string `xml:"managingEditor"`
	//Email address for person responsible for technical issues relating to channel. 	betty@herald.com (Betty Guernsey)
	WebMaster string `xml:"webMaster"`
	//The publication date for the content in the channel. For example, the New York Times publishes on a daily basis, the publication date flips once every 24 hours. That's when the pubDate of the channel changes. All date-times in RSS conform to the Date and Time Specification of RFC 822, with the exception that the year may be expressed with two characters or four characters (four preferred). 	Sat, 07 Sep 2002 0:00:01 GMT
	PubDate string `xml:"pubDate"`
	//The last time the content of the channel changed. 	Sat, 07 Sep 2002 9:42:31 GMT
	LastBuildDate string `xml:"lastBuildDate"`
	//Specify one or more categories that the channel belongs to. Follows the same rules as the <item>-level category element. More info. 	<category>Newspapers</category>
	Category string `xml:"category"`
	//A string indicating the program used to generate the channel. 	MightyInHouse Content System v2.3
	Generator string `xml:"generator"`
	//A URL that points to the documentation for the format used in the RSS file. It's probably a pointer to this page. It's for people who might stumble across an RSS file on a Web server 25 years from now and wonder what it is. 	http:	//backend.userland.com/rss
	Docs string `xml:"docs"`
	//Allows processes to register with a cloud to be notified of updates to the channel, implementing a lightweight publish-subscribe protocol for RSS feeds. More info here. 	<cloud domain="rpc.sys.com" port="80" path="/RPC2" registerProcedure="pingMe" protocol="soap"/>
	Cloud string `xml:"cloud"`
	//ttl stands for time to live. It's a number of minutes that indicates how long a channel can be cached before refreshing from the source. More info here. 	<ttl>60</ttl>
	TTL int `xml:"ttl"`
	//Specifies a GIF, JPEG or PNG image that can be displayed with the channel. More info here.
	Image Image `xml:"image"`
	//Specifies a text input box that can be displayed with the channel. More info here.
	TextInput `xml:"textInput"`
	//A hint for aggregators telling them which hours they can skip. More info here.
	SkipHours SkipHours `xml:"skipHours"`
	//A hint for aggregators telling them which days they can skip. More info here.
	SkipDays SkipDays `xml:"skipDays"`
	Item     []Item   `xml:"item"`
}

//SkipHours ...
type SkipHours struct {
	Hour []string `xml:"hour"`
}

//SkipDays ...
type SkipDays struct {
	Day []string `xml:"day"`
}

//TextInput ...
type TextInput struct {
	Name        string `xml:"name"`
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
}

//Image ...
type Image struct {
	URL         string `xml:"url"`
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Width       int    `xml:"width"`
	Height      int    `xml:"height"`
}

//Item ...
type Item struct {
	Title       string     `xml:"title"`
	Description string     `xml:"description"`
	Link        string     `xml:"link"`
	Author      string     `xml:"author"`
	Category    string     `xml:"category"`
	Comments    string     `xml:"comments"`
	Enclosures  Enclosures `xml:"enclosures"`
	GUID        string     `xml:"guid"`
	PubDate     string     `xml:"pubDate"`
	Source      string     `xml:"source"`
}

//Enclosures ...
type Enclosures struct {
	Type   string `xml:"type"`
	Length int    `xml:"length"`
	URL    string `xml:"url"`
}
