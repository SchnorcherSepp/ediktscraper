package main

import "fmt"

const (
	maxCost           = 30000
	buildableLotUrl   = "https://edikte.justiz.gv.at/edikte/ex/exedi3.nsf/suchedi?SearchView&subf=eex&SearchOrder=4&SearchMax=4999&retfields=~VKat=UL&ftquery=&query=%28%5BVKat%5D%3D%28UL%29%29"
	agriForestLandUrl = "https://edikte.justiz.gv.at/edikte/ex/exedi3.nsf/suchedi?SearchView&subf=eex&SearchOrder=4&SearchMax=4999&retfields=~VKat=LF&ftquery=&query=%28%5BVKat%5D%3D%28LF%29%29"
)

func main() {

	// collect all edicts
	edikts := make([]EdiktElements, 0)
	for _, link := range []string{buildableLotUrl, agriForestLandUrl} {
		// get all doc links
		doc, baseUrl := request(link)
		allDocLinks := findAllDocLinks(doc, baseUrl)

		// get all edikt elements
		for _, allDocLink := range allDocLinks {
			doc, baseUrl = request(allDocLink)
			ees := findEdiktElements(doc, allDocLink)
			edikts = append(edikts, ees)
		}
	}

	// check edicts
	var countInvalid, countExpensive int

	for _, ees := range edikts {
		sw := ees.Schaetzwert()
		if sw < 0 {
			countInvalid++
			continue
		}
		if sw > maxCost {
			countExpensive++
			continue
		}

		fmt.Printf("Schätzwert:\t%d EUR\n", sw)
		fmt.Printf("Objektgröße:\t%d m²\n", ees.Objektgroesse())
		fmt.Printf("Grundstücksgröße:\t%d m²\n", ees.Grundstuecksgroesse())
		fmt.Printf("PlzOrt:\t%s\n", ees.PlzOrt())
		fmt.Printf("Entfernung:\t%d km\n", ees.Entfernung())
		fmt.Printf("AllDocLink:\t%s\n", ees.AllDocLink())
		fmt.Printf("\n")

	}
}
