package main

import (
	"fmt"
)

const (
	maxCost           = 30000
	buildableLotUrl   = "https://edikte.justiz.gv.at/edikte/ex/exedi3.nsf/suchedi?SearchView&subf=eex&SearchOrder=4&SearchMax=4999&retfields=~VKat=UL&ftquery=&query=%28%5BVKat%5D%3D%28UL%29%29"
	agriForestLandUrl = "https://edikte.justiz.gv.at/edikte/ex/exedi3.nsf/suchedi?SearchView&subf=eex&SearchOrder=4&SearchMax=4999&retfields=~VKat=LF&ftquery=&query=%28%5BVKat%5D%3D%28LF%29%29"
)

func main() {

	db := LoadDB()

	ediktAlldocURLs := CollectEdiktAlldocURLs([]string{buildableLotUrl, agriForestLandUrl})

	for _, ediktAlldocURL := range ediktAlldocURLs {
		doc, _, base := RequestPage(ediktAlldocURL)
		edikt := ParseEdikt(doc)

		sw := edikt.Schaetzwert()
		if sw <= 0 || sw > maxCost {
			println("skip", sw, "eur")
			continue
		}
		if isKnown := db.AddEdikt(ediktAlldocURL); isKnown {
			println("known", sw, "eur")
			continue
		}

		fmt.Printf("Schätzwert: %d EUR\n", sw)
		fmt.Printf("Objektgröße: %d m²\n", edikt.Objektgroesse())
		fmt.Printf("Grundstücksgröße: %d m²\n", edikt.Grundstuecksgroesse())
		fmt.Printf("PlzOrt: %s\n", edikt.PlzOrt())
		fmt.Printf("Entfernung: %d km\n", edikt.Entfernung())
		fmt.Printf("AllDocLink: %s\n", ediktAlldocURL)
		fmt.Printf("Kurzgutachten: %s\n", edikt.KurzgutachtenLink(base))
		fmt.Printf("Langgutachten: %v\n", edikt.LanggutachtenLinks(base))
		fmt.Printf("Langgutachten: %d\n", len(edikt.Langgutachten(base)))
		fmt.Printf("KurzgutachtenText: %s\n", edikt.Kurzgutachten(base))
		fmt.Printf("------------------------------\n")
	}

}
