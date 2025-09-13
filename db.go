package main

import (
	"encoding/gob"
	"errors"
	"os"
)

const dbPath = "db.dat"

type DB struct {
	Edikt map[string]bool
}

// AddEdikt adds the given alldocURL to the set of known entries.
// It returns true if the URL was already present before this call.
// The method initializes the map on first use and persists the DB to disk.
func (db *DB) AddEdikt(alldocURL string) (isKnown bool) {
	if db.Edikt == nil {
		db.Edikt = make(map[string]bool)
	}
	_, isKnown = db.Edikt[alldocURL]
	db.Edikt[alldocURL] = true
	db.Save()
	return isKnown
}

// Save writes the DB to disk at dbPath using gob encoding.
// It panics on I/O or encoding errors.
func (db *DB) Save() {
	f, err := os.Create(dbPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := gob.NewEncoder(f).Encode(db); err != nil {
		panic(err)
	}
}

// LoadDB loads a DB from dbPath using gob decoding.
// If the file does not exist, it returns a new empty DB.
// It panics on any other error.
func LoadDB() *DB {
	f, err := os.Open(dbPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return new(DB)
		}
		panic(err)
	}
	defer f.Close()

	var db DB
	if err := gob.NewDecoder(f).Decode(&db); err != nil {
		panic(err)
	}
	return &db
}
