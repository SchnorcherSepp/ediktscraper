package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Request downloads the raw response body for the given URL.
// It performs an HTTP GET with a context deadline and browser-like headers.
// Returns the body bytes as-is (HTML, PDF, etc.). On any error, it logs and exits.
// Only 2xx responses are accepted.
// No size limit is enforced; large responses will be fully buffered in memory.
func Request(link string) []byte {

	// Create a context with a hard deadline so the request cannot hang forever.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// HTTP client with a sane timeout for connection + response.
	client := &http.Client{
		Timeout: 25 * time.Second,
	}

	// Build a GET request bound to the context and set pragmatic headers.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
	if err != nil {
		panic(err) // Abort immediately on request construction failure.
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:142.0) Gecko/20100101 Firefox/142.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "de,en;q=0.8")

	// Execute the HTTP call.
	resp, err := client.Do(req)
	if err != nil {
		println("link:", link)
		panic(err) // Abort on transport-level error.
	}
	defer resp.Body.Close()

	// Enforce a successful 2xx status.
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		panic(resp.Status)
	}

	// Read the full body once so we can both parse and return it as text.
	// If very large pages are expected, consider a size cap.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	// return (html or pdf)
	return body
}

// RequestPage fetches a URL, parses the HTML, and returns:
//   - doc: goquery document built from the response body
//   - source: the raw HTML as a UTF-8 string
//   - base: the base URL derived from the input link
//
// On any error, it logs and exits.
// If the response is not HTML, parsing will fail.
// Base URL is parsed from the provided link and does not reflect redirects.
func RequestPage(link string) (doc *goquery.Document, source string, base *url.URL) {

	// download source code
	body := Request(link)
	source = string(body)

	// Build a goquery document from the in-memory bytes.
	var err error
	doc, err = goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		panic(err) // Abort if HTML cannot be parsed.
	}

	// Derive the base URL. Currently reads from an external variable.
	// Consider using resp.Request.URL or the input link instead.
	base, err = url.Parse(link)
	if err != nil {
		panic(err)
	}

	// Return the parsed document and the base URL.
	return doc, source, base
}
