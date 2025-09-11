package main

import (
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func findAllDocLinks(doc *goquery.Document, base *url.URL) []string {
	ret := make([]string, 0, 50)
	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		if href, ok := s.Attr("href"); ok && strings.HasPrefix(href, "alldoc") {
			rel, _ := url.Parse(href)
			abs := base.ResolveReference(rel).String()
			ret = append(ret, abs)
		}
	})
	return ret
}
