package dto

import "encoding/xml"

type RSS struct {
    XMLName xml.Name  `xml:"rss"`
    Channel RSSChannel `xml:"channel"`
}

type RSSChannel struct {
    Title string     `xml:"title"`
    Link  string     `xml:"link"`
    Items []RSSItem  `xml:"item"`
}

type RSSItem struct {
    Title       string `xml:"title"`
    Link        string `xml:"link"`
    Description string `xml:"description"`
    PubDate     string `xml:"pubDate"`
}


