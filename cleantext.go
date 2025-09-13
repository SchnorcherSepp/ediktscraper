package main

import (
	"regexp"
	"strings"
)

// Text cleanup utilities for normalizing and structuring extracted text.

// Labels that should start on a new line if they appear mid-line.
// Some contain regex escapes (e.g., "\\." to match a literal dot).
var fieldLabels = []string{
	"Dienststelle", "Aktenzeichen", "wegen", "Grundbuch", "EZ", "Grundstücksnr\\.", "BLNr",
	"Adresse", "PLZ/Ort", "Kategorie\\(n\\)", "Beschreibung \\(WE\\)", "Grundstücksgröße",
	"Stichtag", "Schätzwert", "Wert des mitzuversteigernden Zubehörs", "erstellt von", "Ausdruck vom",
}

// CleanText applies a deterministic cleanup pipeline to free-form text:
//  1. Normalize newlines and spaces (including Unicode variants).
//  2. Remove known navigation/footer noise lines.
//  3. Ensure a single space after a colon ("Key:Value" -> "Key: Value").
//  4. Force a line break before known field labels if they appear mid-line.
//  5. Trim leading and trailing whitespace on each line.
//  6. Collapse internal runs of spaces to a single space.
//  7. Collapse 3+ blank lines to exactly one empty line (two '\n').
//
// Returns the cleaned string.
func CleanText(s string) string {

	// 1) Normalize newlines and spaces
	s = normalizeSpacesAndNewlines(s)

	// 2) Drop known navigation/footer noise lines
	s = dropNoiseLines(s)

	// 3) Ensure a space after colon like "Key: Value"
	// Pattern: a colon followed by a non-whitespace character -> insert a space.
	reAfterColon := regexp.MustCompile(`:([^\s])`)
	s = reAfterColon.ReplaceAllString(s, `: $1`)

	// 4) Insert line breaks before known field labels if not at line start
	s = breakBeforeLabels(s, fieldLabels)

	// 5) Trim leading whitespace on each line and trailing spaces
	// (?m) enables ^ and $ to match line starts/ends within the multiline string.
	reLead := regexp.MustCompile(`(?m)^[\t \p{Zs}]+`)
	s = reLead.ReplaceAllString(s, "")
	reTrail := regexp.MustCompile(`(?m)[\t \p{Zs}]+$`)
	s = reTrail.ReplaceAllString(s, "")

	// 6) Collapse internal multiple spaces (ASCII spaces, tabs, Unicode spaces) to one space
	reMultiSpace := regexp.MustCompile(`[ \t\p{Zs}]{2,}`)
	s = reMultiSpace.ReplaceAllString(s, " ")

	// 7) Collapse blank lines: 3+ '\n' -> exactly two '\n' (one empty line)
	reBlank3p := regexp.MustCompile(`\n{3,}`)
	s = reBlank3p.ReplaceAllString(s, "\n\n")

	// Final trim of leading/trailing whitespace
	s = strings.TrimSpace(s)

	// Write output
	return s
}

// normalizeSpacesAndNewlines replaces CRLF/CR with LF, tabs with a space,
// normalizes various Unicode spaces to ASCII space, and removes zero-width characters.
func normalizeSpacesAndNewlines(s string) string {
	// Normalize newlines to LF
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	// Convert tabs to single spaces
	s = strings.ReplaceAll(s, "\t", " ")

	// Replace odd Unicode spaces with a normal space; drop zero-width chars
	replacer := strings.NewReplacer(
		"\u00A0", " ", // NBSP
		"\u2000", " ", "\u2001", " ", "\u2002", " ", "\u2003", " ", "\u2004", " ",
		"\u2005", " ", "\u2006", " ", "\u2007", " ", "\u2008", " ", "\u2009", " ",
		"\u200A", " ", "\u202F", " ", "\u205F", " ", "\u3000", " ",
		"\u200B", "", // zero-width space
		"\uFEFF", "", // zero-width no-break space
	)
	s = replacer.Replace(s)

	return s
}

// dropNoiseLines removes known navigation/footer boilerplate from the text.
// Regex uses (?im): case-insensitive and multiline. It matches whole lines
// containing items like "zur Navigation", "Kontakt", "Impressum", etc., and deletes them.
func dropNoiseLines(s string) string {
	re := regexp.MustCompile(`(?im)^[\t \p{Zs}]*(zur Navigation|Glossar|Kontakt|Datenschutz[- ]?Erklärung|Impressum|Barrierefreiheit|zum Suchergebnis|Lesezeichen)[\t \p{Zs}]*\n?`)
	return re.ReplaceAllString(s, "")
}

// breakBeforeLabels inserts a newline before any provided label if the label
// occurs mid-line. It leaves labels at the start of a line unchanged.
// Pattern explanation:
//
//	([^\n])           capture a non-newline character (ensures mid-line)
//	((?:LABELS):)     capture the label (from the alternation) followed by a colon
//
// Replacement inserts '\n' between the two captured groups.
func breakBeforeLabels(s string, labels []string) string {
	joined := strings.Join(labels, "|")
	re := regexp.MustCompile(`([^\n])((?:` + joined + `):)`)
	return re.ReplaceAllString(s, "$1\n$2")
}
