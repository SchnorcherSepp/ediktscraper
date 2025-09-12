package db

import (
	"log"
	"os"
)

const (
	NEW = iota
	CHANGED
	KNOWN
)

const path = "data/"

func init() {
	if err := os.MkdirAll(path, 777); err != nil {
		log.Fatal(err)
	}
}

func SetEditk(link, source string) int {
	return NEW
}
