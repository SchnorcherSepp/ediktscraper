package main

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func request(link string) (doc *goquery.Document, base *url.URL) {

	// Context with timeout to avoid hanging requests.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// HTTP client with sane timeout.
	client := &http.Client{
		Timeout: 25 * time.Second,
	}

	// Build request with reasonable headers.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:142.0) Gecko/20100101 Firefox/142.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "de,en;q=0.8")

	// Execute request.
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Fatalf("unexpected HTTP status: %s", resp.Status)
	}

	// Parse HTML.
	doc, err = goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// base url
	base, err = url.Parse(buildableLotUrl)
	if err != nil {
		log.Fatal(err)
	}

	// return
	return doc, base
}
