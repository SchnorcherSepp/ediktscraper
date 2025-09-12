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

func (db *DB) AddEdikt(alldocURL string) (isKnown bool) {
	if db.Edikt == nil {
		db.Edikt = make(map[string]bool)
	}
	_, isKnown = db.Edikt[alldocURL]
	db.Edikt[alldocURL] = true
	db.Save()
	return isKnown
}

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
