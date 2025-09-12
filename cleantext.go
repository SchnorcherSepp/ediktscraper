package main

import (
	"regexp"
	"strings"
)

// Labels we want to start on a new line if they appear mid-line
var fieldLabels = []string{
	"Dienststelle", "Aktenzeichen", "wegen", "Grundbuch", "EZ", "Grundstücksnr\\.", "BLNr",
	"Adresse", "PLZ/Ort", "Kategorie\\(n\\)", "Beschreibung \\(WE\\)", "Grundstücksgröße",
	"Stichtag", "Schätzwert", "Wert des mitzuversteigernden Zubehörs", "erstellt von", "Ausdruck vom",
}

func cleanText(s string) string {

	// 1) Normalize newlines and spaces
	s = normalizeSpacesAndNewlines(s)

	// 2) Drop known navigation/footer noise lines
	s = dropNoiseLines(s)

	// 3) Ensure a space after colon like "Key: Value"
	reAfterColon := regexp.MustCompile(`:([^\s])`)
	s = reAfterColon.ReplaceAllString(s, `: $1`)

	// 4) Insert line breaks before known field labels if not at line start
	s = breakBeforeLabels(s, fieldLabels)

	// 5) Trim leading whitespace on each line and trailing spaces
	reLead := regexp.MustCompile(`(?m)^[\t \p{Zs}]+`)
	s = reLead.ReplaceAllString(s, "")
	reTrail := regexp.MustCompile(`(?m)[\t \p{Zs}]+$`)
	s = reTrail.ReplaceAllString(s, "")

	// 6) Collapse internal multiple spaces
	reMultiSpace := regexp.MustCompile(`[ \t\p{Zs}]{2,}`)
	s = reMultiSpace.ReplaceAllString(s, " ")

	// 7) Collapse blank lines according to flag
	// Keep at most one empty line: collapse 3+ \n to exactly two \n
	reBlank3p := regexp.MustCompile(`\n{3,}`)
	s = reBlank3p.ReplaceAllString(s, "\n\n")

	s = strings.TrimSpace(s)

	// Write output
	return s
}

// normalizeSpacesAndNewlines replaces CRLF/CR with LF, tabs with space, various unicode spaces,
// and removes zero-width characters.
// Comments are in English per user request.
func normalizeSpacesAndNewlines(s string) string {
	// Newlines
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	// Tabs to single spaces
	s = strings.ReplaceAll(s, "\t", " ")

	// Map of odd spaces to normal space; zero-width to empty
	replacer := strings.NewReplacer(
		"\u00A0", " ", // NBSP
		"\u2000", " ", "\u2001", " ", "\u2002", " ", "\u2003", " ", "\u2004", " ",
		"\u2005", " ", "\u2006", " ", "\u2007", " ", "\u2008", " ", "\u2009", " ",
		"\u200A", " ", "\u202F", " ", "\u205F", " ", "\u3000", " ",
		"\u200B", "", // zero width space
		"\uFEFF", "", // zero width no-break space
	)
	s = replacer.Replace(s)

	return s
}

// dropNoiseLines removes known navigation/footer lines from the export.
func dropNoiseLines(s string) string {
	// (?im) = case-insensitive, multi-line
	re := regexp.MustCompile(`(?im)^[\t \p{Zs}]*(zur Navigation|Glossar|Kontakt|Datenschutz[- ]?Erklärung|Impressum|Barrierefreiheit|zum Suchergebnis|Lesezeichen)[\t \p{Zs}]*\n?`)
	return re.ReplaceAllString(s, "")
}

// breakBeforeLabels inserts a newline before any of the labels if they are not at line start.
func breakBeforeLabels(s string, labels []string) string {
	joined := strings.Join(labels, "|")
	// Capture a non-newline, then the label+colon, and insert a newline between
	re := regexp.MustCompile(`([^\n])((?:` + joined + `):)`)
	return re.ReplaceAllString(s, "$1\n$2")
}
