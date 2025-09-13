package main

import (
	"ediktscraper/email"
	"fmt"
)

const (
	// maxCost caps the acceptable "Schätzwert" (appraised value) in EUR.
	maxCost = 30000
	// buildableLotUrl lists "UL" category items (buildable lots) in the edikte portal.
	buildableLotUrl = "https://edikte.justiz.gv.at/edikte/ex/exedi3.nsf/suchedi?SearchView&subf=eex&SearchOrder=4&SearchMax=4999&retfields=~VKat=UL&ftquery=&query=%28%5BVKat%5D%3D%28UL%29%29"
	// agriForestLandUrl lists "LF" category items (agricultural/forest land) in the edikte portal.
	agriForestLandUrl = "https://edikte.justiz.gv.at/edikte/ex/exedi3.nsf/suchedi?SearchView&subf=eex&SearchOrder=4&SearchMax=4999&retfields=~VKat=LF&ftquery=&query=%28%5BVKat%5D%3D%28LF%29%29"
)

func main() {

	// Load persistent database of already-seen edikt "alldoc" URLs.
	db := LoadDB()

	// Collect all edikt "alldoc" URLs from both category search pages.
	ediktAlldocURLs := CollectEdiktAlldocURLs([]string{buildableLotUrl, agriForestLandUrl})

	// Process each edikt page independently.
	var body string
	for _, ediktAlldocURL := range ediktAlldocURLs {
		// Fetch the edikt page and parse it into a document.
		// base is the resolved base URL used for converting relative links to absolute.
		doc, _, base := RequestPage(ediktAlldocURL)

		// Extract structured fields from the document.
		edikt := ParseEdikt(doc)

		// -------------------------------------------------------------------------------------

		// Validate entry: require positive appraised value and at least one long-appraisal link.
		sw := edikt.Schaetzwert()
		if sw <= 0 || len(edikt.LanggutachtenLinks(base)) == 0 {
			println("Canceled", sw, "eur")
			continue
		}

		// Enforce budget cap: skip items priced above maxCost.
		if sw > maxCost {
			println("Expensive", sw, "eur")
			continue
		}

		// De-duplicate: skip entries that were already processed earlier.
		if isKnown := db.AddEdikt(ediktAlldocURL); isKnown {
			println("Known", sw, "eur")
			continue
		}

		// -------------------------------------------------------------------------------------

		// Preview
		var m string
		m += fmt.Sprintf("╔═══════════════════════════════════════════════════════════════════════════════════\n")
		m += fmt.Sprintf("║  Schätzwert:    %d EUR\n", sw)
		m += fmt.Sprintf("║  Objektgröße:   %d m²\n", edikt.Objektgroesse())
		m += fmt.Sprintf("║  Grundgröße:    %d m²\n", edikt.Grundstuecksgroesse())
		m += fmt.Sprintf("║  PlzOrt:        %s\n", edikt.PlzOrt())
		m += fmt.Sprintf("║  Entfernung:    %d km\n", edikt.Entfernung())
		m += fmt.Sprintf("║  AllDocLink:    %s\n", ediktAlldocURL)
		m += fmt.Sprintf("║  Kurzgutachten: %s\n", edikt.KurzgutachtenLink(base))
		for _, l := range edikt.LanggutachtenLinks(base) {
			m += fmt.Sprintf("║  Langgutachten: %v\n", l)
		}
		m += fmt.Sprintf("╚═══════════════════════════════════════════════════════════════════════════════════\n")
		fmt.Println(m)
		body += m
	}

	// send email
	if len(body) > 0 {
		email.SendEmail("edikte@post4me.at", "Edikte: Neuigkeiten des Tages!", body)
	}
}
