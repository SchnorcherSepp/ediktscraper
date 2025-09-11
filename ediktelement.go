package main

import (
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

/*
	Keys:
		x Bekannt gemacht am
		x Beschreibung des mitzuversteigernden Zubehörs
		x Termin
		- Objektgröße
		x Versteigerung Grundstücke EZ 87 (16.10.2025 09:00)
		x Wert des mitzuversteigernden Zubehörs
		x Foto(s)
		- Grundstücksgröße
		- Langgutachten
		x Sonstiges
		x Aktenzeichen
		x Dienststelle
		x Kategorie(n)
		x Sonstige Hinweise
		- Kurzgutachten
		- Schätzwert
		x Besichtigungszeit
		x Versteigerung ehemalige Hofstelle (24.09.2025 10:15)
		x Versteigerung Wohnhaus mit Wirtschaftstrakt (25.09.2025 09:30)
		x Versteigerung GST - Nr. 4661 (07.10.2025 09:00)
		x Zuschlag ohne Überbot LN in Stoob
		x Zuschlag ohne Überbot LN in Neutal
		x Geringstes Gebot
		x EZ
		x Ort
		x Versteigerung Einheit B (24.09.2025 10:15)
		x Versteigerung Einheit A (24.09.2025 10:15)
		x Versteigerung Wohnhaus GST - Nr. 214/2, 215 (07.10.2025 09:00)
		x Lageplan
		x Versteigerung EZ 118 und EZ 1563 (16.10.2025 09:00)
		x Grundbuch
		x Vadium
		x Versteigerungsort
		x Versteigerung landw. Grundstück (25.09.2025 09:30)
		x Versteigerung Grundstücke EZ 205 (16.10.2025 09:00)
		x Beschreibung (WE)
		x BLNr
		x Grundriss(e)
		x Ort und Zeit der Einsichtnahme
		x Letzte Änderung am
		x Grundstücksnr.
		- PLZ/Ort
		x wegen
		x Telefonkontakt
		- Liegenschaftsadresse
		x Versteigerungstermin
*/

type EdiktElement struct {
	Key   string
	Value string
	Obj   *goquery.Selection
}

type EdiktElements map[string]*EdiktElement

func findEdiktElements(doc *goquery.Document, allDocLink string) EdiktElements {
	ret := make(EdiktElements)
	ret["allDocLink"] = &EdiktElement{
		Key:   "allDocLink",
		Value: allDocLink,
		Obj:   new(goquery.Selection),
	}

	rows := doc.Find("div.row")
	rows.Each(func(i int, r *goquery.Selection) {

		// Get key from <span.col-sm-3>
		keySect := r.Find("span.col-sm-3")
		key := strings.TrimSpace(keySect.Text())
		key = strings.TrimSuffix(key, ":")

		// Get value from <p.col-sm-9>
		valueSect := r.Find("p.col-sm-9")
		value := strings.TrimSpace(valueSect.Text())
		value = strings.ReplaceAll(value, "\u00a0", " ")

		// add to list
		if len(key) > 0 {
			ee := new(EdiktElement)
			ee.Key = key
			ee.Value = value
			ee.Obj = valueSect
			ret[key] = ee
		}
	})

	return ret
}

//----------------------------------------------------------------------------

func (ees EdiktElements) AllDocLink() string {
	ee := ees["allDocLink"]
	return ee.Value
}

func (ees EdiktElements) Schaetzwert() int {
	ee, ok := ees["Schätzwert"]
	if !ok {
		return -1
	}

	// remove EUR
	spl := strings.Split(ee.Value, " ")
	value := spl[0]
	// remove ,00
	spl = strings.Split(value, ",")
	value = spl[0]
	// remove .
	value = strings.ReplaceAll(value, ".", "")
	// trim space
	value = strings.TrimSpace(value)

	if len(value) == 0 {
		return -2
	}

	i, err := strconv.Atoi(value)
	if err != nil {
		return -3
	}

	return i
}

func (ees EdiktElements) Objektgroesse() int {
	ee, ok := ees["Objektgröße"]
	if !ok {
		return 0 // no value is a valid option
	}

	// remove m²
	spl := strings.Split(ee.Value, " ")
	value := spl[0]
	// remove ,00
	spl = strings.Split(value, ",")
	value = spl[0]
	// remove .
	value = strings.ReplaceAll(value, ".", "")
	// trim space
	value = strings.TrimSpace(value)

	if len(value) == 0 {
		return -2
	}

	i, err := strconv.Atoi(value)
	if err != nil {
		return -3
	}

	return i
}

func (ees EdiktElements) Grundstuecksgroesse() int {
	ee, ok := ees["Grundstücksgröße"]
	if !ok {
		return -1
	}

	// remove m²
	spl := strings.Split(ee.Value, " ")
	value := spl[0]
	// remove ,00
	spl = strings.Split(value, ",")
	value = spl[0]
	// remove .
	value = strings.ReplaceAll(value, ".", "")
	// trim space
	value = strings.TrimSpace(value)

	if len(value) == 0 {
		return -2
	}

	i, err := strconv.Atoi(value)
	if err != nil {
		panic(err)
		return -3
	}

	return i
}

func (ees EdiktElements) PlzOrt() string {
	ee, ok := ees["PLZ/Ort"]
	if !ok {
		return "---"
	}
	return ee.Value
}

func (ees EdiktElements) Entfernung() int {
	return distance(ees.PlzOrt())
}

func (ees EdiktElements) Liegenschaftsadresse() string {
	ee, ok := ees["Liegenschaftsadresse"]
	if !ok {
		return "---"
	}
	return ee.Value
}

func (ees EdiktElements) Langgutachten() []string {
	ee, ok := ees["Langgutachten"]
	if !ok {
		return make([]string, 0)
	}

	base, err := url.Parse(ees.AllDocLink())
	if err != nil {
		log.Fatal(err)
	}

	ret := make([]string, 0)
	ee.Obj.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		if href, ok := s.Attr("href"); ok {
			rel, _ := url.Parse(href)
			abs := base.ResolveReference(rel).String()
			ret = append(ret, abs)
		}
	})

	return ret
}

func (ees EdiktElements) Kurzgutachten() []string {
	ee, ok := ees["Kurzgutachten"]
	if !ok {
		return make([]string, 0)
	}

	base, err := url.Parse(ees.AllDocLink())
	if err != nil {
		log.Fatal(err)
	}

	ret := make([]string, 0)
	ee.Obj.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		if href, ok := s.Attr("href"); ok {
			rel, _ := url.Parse(href)
			abs := base.ResolveReference(rel).String()
			ret = append(ret, abs)
		}
	})

	return ret
}

func (ees EdiktElements) KurzgutachtenText() string {
	links := ees.Kurzgutachten()
	if len(links) == 0 {
		return "---"
	}
	if len(links) > 1 {
		log.Fatal("too many Kurzgutachten links")
	}

	doc, _ := request(links[0])
	text := doc.Text()
	text = strings.ReplaceAll(text, "\n", "|")
	re := regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)
	return text
}
