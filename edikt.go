package main

import (
	"ediktscraper/openstreetmap"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Edikt maps a field label to its DOM selection containing the value.
// Keys are the visible labels (e.g., "PLZ/Ort"), values are <p.col-sm-9> nodes.
type Edikt map[string]*goquery.Selection

// ParseEdikt builds an Edikt from the given document.
// It scans rows shaped like:
//
//	<div class="row">
//	  <span class="col-sm-3">Label:</span>
//	  <p class="col-sm-9">Value...</p>
//	</div>
//
// Returns a map from cleaned label (colon removed) to the value <p>.
func ParseEdikt(doc *goquery.Document) Edikt {
	edikt := make(Edikt)

	rows := doc.Find("div.row")
	rows.Each(func(i int, r *goquery.Selection) {

		// Extract the key from <span.col-sm-3>, remove trailing colon, trim whitespace.
		keySect := r.Find("span.col-sm-3")
		key := strings.TrimSpace(keySect.Text())
		key = strings.TrimSuffix(key, ":")

		// Extract the value container from <p.col-sm-9>.
		valueSect := r.Find("p.col-sm-9")

		// Store field selection under its key. Overwrites on duplicate keys.
		edikt[key] = valueSect
	})

	return edikt
}

//------------------------------------------------------------------------------------------------------------

// Get returns the value selection for the given key.
// If the key is missing, it returns an empty *goquery.Selection to avoid nil checks.
func (e Edikt) Get(key string) *goquery.Selection {
	valueSect, ok := e[key]
	if !ok {
		return new(goquery.Selection)
	}
	return valueSect
}

// GetTxt returns the text content for the given key.
// It trims surrounding whitespace and normalizes NBSP to a normal space.
func (e Edikt) GetTxt(key string) string {
	valueSect := e.Get(key)
	value := strings.TrimSpace(valueSect.Text())
	value = strings.ReplaceAll(value, "\u00a0", " ")
	return value
}

// GetInt parses a positive integer from the field's text.
// It tolerates formats like "1.234,00 EUR" by:
//   - splitting on space and taking the first token,
//   - dropping the decimal part after a comma,
//   - removing dots as thousands separators,
//   - trimming spaces.
//
// Returns 0 for empty strings, -1 if parsing fails.
func (e Edikt) GetInt(key string) int {
	value := e.GetTxt(key)

	// Keep the first token before any space (drops units like EUR, m², etc.).
	spl := strings.Split(value, " ")
	value = spl[0]

	// Drop decimal part after a comma.
	spl = strings.Split(value, ",")
	value = spl[0]

	// Remove thousands separators.
	value = strings.ReplaceAll(value, ".", "")

	// Final trim.
	value = strings.TrimSpace(value)

	// Special case: empty means zero.
	if len(value) == 0 {
		return 0
	}

	// Parse as base-10 integer.
	i, err := strconv.Atoi(value)
	if err != nil {
		return -1
	}
	return i
}

// GetLinks collects absolute URLs from all <a href> elements within the value of key.
// baseURL is used to resolve relative links. Empty or whitespace-only hrefs are skipped.
func (e Edikt) GetLinks(key string, baseURL *url.URL) []string {
	links := make([]string, 0, 8)
	valueSect := e.Get(key)

	// Find all anchors with an href attribute inside the section.
	valueSect.Find("a[href]").Each(func(_ int, a *goquery.Selection) {
		if href, ok := a.Attr("href"); ok {
			href = strings.TrimSpace(href)
			if href == "" {
				return // skip empty
			}
			// Resolve relative link against base URL to get an absolute URL.
			rel, _ := url.Parse(href) // ignore parse error; ResolveReference will handle nil poorly but href came from DOM
			abs := baseURL.ResolveReference(rel).String()
			links = append(links, abs)
		}
	})
	return links
}

//------------------------------------------------------------------------------------------------------------

// Schaetzwert returns the integer value of the "Schätzwert" field.
// See GetInt for parsing behavior.
func (e Edikt) Schaetzwert() int {
	return e.GetInt("Schätzwert")
}

// Objektgroesse returns the integer value of the "Objektgröße" field.
func (e Edikt) Objektgroesse() int {
	return e.GetInt("Objektgröße")
}

// Grundstuecksgroesse returns the integer value of the "Grundstücksgröße" field.
func (e Edikt) Grundstuecksgroesse() int {
	return e.GetInt("Grundstücksgröße")
}

// PlzOrt returns the "PLZ/Ort" field as plain text.
func (e Edikt) PlzOrt() string {
	return e.GetTxt("PLZ/Ort")
}

// Entfernung returns a distance metric derived from the "PLZ/Ort" using openstreetmap.Distance.
// The unit and semantics depend on the openstreetmap implementation.
func (e Edikt) Entfernung() int {
	return openstreetmap.Distance(e.PlzOrt())
}

// Liegenschaftsadresse returns the "Liegenschaftsadresse" field as plain text.
func (e Edikt) Liegenschaftsadresse() string {
	return e.GetTxt("Liegenschaftsadresse")
}

// KurzgutachtenLink returns the single absolute URL from the "Kurzgutachten" field.
// Returns an empty string if no link exists. Panics if more than one link is present.
func (e Edikt) KurzgutachtenLink(baseURL *url.URL) string {
	links := e.GetLinks("Kurzgutachten", baseURL)
	if len(links) == 0 {
		return ""
	}
	if len(links) != 1 {
		panic("more than one link")
	}
	return links[0]
}

// Kurzgutachten fetches and returns the cleaned text of the short appraisal page.
// Returns an empty string if no link is available. Errors from RequestPage are ignored.
func (e Edikt) Kurzgutachten(baseURL *url.URL) string {
	link := e.KurzgutachtenLink(baseURL)
	if len(link) == 0 {
		return "" // no link available
	}

	doc, _, _ := RequestPage(link) // ignoring status and error
	txt := doc.Find("body").Text()
	return CleanText(txt)
}

// LanggutachtenLinks returns all absolute URLs from the "Langgutachten" field.
func (e Edikt) LanggutachtenLinks(baseURL *url.URL) []string {
	return e.GetLinks("Langgutachten", baseURL)
}

// Langgutachten downloads all long appraisal files referenced by LanggutachtenLinks.
// Returns a slice of file contents as byte slices. Errors from Request are ignored.
func (e Edikt) Langgutachten(baseURL *url.URL) [][]byte {
	files := make([][]byte, 0)
	for _, link := range e.LanggutachtenLinks(baseURL) {
		b := Request(link) // ignoring errors and HTTP status
		files = append(files, b)
	}
	return files
}
