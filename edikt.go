package main

import (
	"ediktscraper/openstreetmap"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Edikt map[string]*goquery.Selection

func ParseEdikt(doc *goquery.Document) Edikt {
	edikt := make(Edikt)

	rows := doc.Find("div.row")
	rows.Each(func(i int, r *goquery.Selection) {

		// Get key from <span.col-sm-3>
		keySect := r.Find("span.col-sm-3")
		key := strings.TrimSpace(keySect.Text())
		key = strings.TrimSuffix(key, ":")

		// Get value from <p.col-sm-9>
		valueSect := r.Find("p.col-sm-9")

		// add edikt field
		edikt[key] = valueSect
	})

	return edikt
}

//------------------------------------------------------------------------------------------------------------

func (e Edikt) Get(key string) *goquery.Selection {
	valueSect, ok := e[key]
	if !ok {
		return new(goquery.Selection)
	}
	return valueSect
}

func (e Edikt) GetTxt(key string) string {
	valueSect := e.Get(key)
	value := strings.TrimSpace(valueSect.Text())
	value = strings.ReplaceAll(value, "\u00a0", " ")
	return value
}

func (e Edikt) GetInt(key string) int {
	value := e.GetTxt(key)

	// remove EUR, m², ...
	spl := strings.Split(value, " ")
	value = spl[0]
	// remove ,00
	spl = strings.Split(value, ",")
	value = spl[0]
	// remove all .
	value = strings.ReplaceAll(value, ".", "")
	// trim space
	value = strings.TrimSpace(value)

	// special case zero
	if len(value) == 0 {
		return 0
	}

	// parse int
	i, err := strconv.Atoi(value)
	if err != nil {
		return -1
	}
	return i
}

func (e Edikt) GetLinks(key string, baseURL *url.URL) []string {
	links := make([]string, 0, 8)
	valueSect := e.Get(key)

	// Find all anchors with an href attribute inside the section
	valueSect.Find("a[href]").Each(func(_ int, a *goquery.Selection) {
		if href, ok := a.Attr("href"); ok {
			href = strings.TrimSpace(href)
			if href == "" {
				return // skip empty
			}
			// build abs links
			rel, _ := url.Parse(href)
			abs := baseURL.ResolveReference(rel).String()
			links = append(links, abs)
		}
	})
	return links
}

//------------------------------------------------------------------------------------------------------------

func (e Edikt) Schaetzwert() int {
	return e.GetInt("Schätzwert")
}

func (e Edikt) Objektgroesse() int {
	return e.GetInt("Objektgröße")
}

func (e Edikt) Grundstuecksgroesse() int {
	return e.GetInt("Grundstücksgröße")
}

func (e Edikt) PlzOrt() string {
	return e.GetTxt("PLZ/Ort")
}

func (e Edikt) Entfernung() int {
	return openstreetmap.Distance(e.PlzOrt())
}

func (e Edikt) Liegenschaftsadresse() string {
	return e.GetTxt("Liegenschaftsadresse")
}

func (e Edikt) KurzgutachtenLink(baseURL *url.URL) string {
	links := e.GetLinks("Kurzgutachten", baseURL)
	if len(links) == 0 {
		return ""
	}
	if len(links) != 1 {
		panic("more then one link")
	}
	return links[0]
}

func (e Edikt) Kurzgutachten(baseURL *url.URL) string {
	link := e.KurzgutachtenLink(baseURL)
	if len(link) == 0 {
		return "" // err: no link
	}

	doc, _, _ := RequestPage(link)
	txt := doc.Find("body").Text()
	return cleanText(txt)
}

func (e Edikt) LanggutachtenLinks(baseURL *url.URL) []string {
	return e.GetLinks("Langgutachten", baseURL)
}

func (e Edikt) Langgutachten(baseURL *url.URL) [][]byte {
	files := make([][]byte, 0)
	for _, link := range e.LanggutachtenLinks(baseURL) {
		b := Request(link)
		files = append(files, b)
	}
	return files
}
