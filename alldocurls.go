package main

import (
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// CollectEdiktAlldocURLs iterates over Edikt search result pages, downloads each page,
// extracts all "alldoc..." links via ExtractEdiktAlldocURLs, and returns a flat slice of absolute URLs.
func CollectEdiktAlldocURLs(ediktPageLinks []string) []string {
	// Preallocate with zero length. Capacity is unknown; keep default to avoid guesswork.
	edikte := make([]string, 0, 50)

	// Iterate over each search result page URL.
	for _, link := range ediktPageLinks {
		// Fetch and parse the page. The second return value is intentionally ignored per the caller's API.
		doc, _, baseURL := RequestPage(link)
		// Extract all absolute "alldoc" URLs from the page.
		ediktAlldocURLs := extractEdiktAlldocURLs(doc, baseURL)
		// Append the page's results to the aggregated slice.
		edikte = append(edikte, ediktAlldocURLs...)
	}

	// Return the aggregated list of absolute "alldoc" URLs.
	return edikte
}

// ExtractEdiktAlldocURLsextractEdiktAlldocURLs extracts absolute URLs from <a> elements on the Austrian "Edikte"
// search page whose href starts with the literal prefix "alldoc". It resolves relative hrefs against 'base'.
// Site context: https://edikte.justiz.gv.at/edikte/
//
// Output a slice of absolute URL strings, in encounter order. Duplicates are preserved.
func extractEdiktAlldocURLs(doc *goquery.Document, base *url.URL) []string {
	// Preallocate to reduce reallocations for typical page sizes
	ret := make([]string, 0, 50)

	// Iterate all anchors that have an href attribute
	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		// Only consider links starting with "alldoc"
		if href, ok := s.Attr("href"); ok && strings.HasPrefix(href, "alldoc") {
			// Parse and resolve to absolute URL; ignore parse errors per original behavior
			rel, _ := url.Parse(href)
			abs := base.ResolveReference(rel).String()
			ret = append(ret, abs)
		}
	})

	return ret
}
