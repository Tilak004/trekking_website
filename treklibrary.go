package main

import (
	_ "embed"
	"encoding/json"
	"log"
)

// Known treks: the fixed facts of treks we run (region, distance, altitude,
// summary, description, itinerary, inclusions, season…), kept in
// treklibrary.json. "New trek" in the admin console can start from one of
// these, so only the price, photo and dates need filling in.

//go:embed treklibrary.json
var trekLibraryJSON []byte

var trekLibrary []Trek

func init() {
	if err := json.Unmarshal(trekLibraryJSON, &trekLibrary); err != nil {
		log.Fatalf("treklibrary.json: %v", err)
	}
}
