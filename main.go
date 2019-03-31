package main

import (
	"log"
	"os"

    "github/john525/footprints-api/server"
)

func main() {
	f, err := os.OpenFile("logfile", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
	    log.Fatalf("error creating log file: %v", err)
	}
	defer f.Close()
	
	log.SetOutput(f)
	log.Println("Building Footprints API Server Log")

	etl := server.NewETL();
	err = etl.ExtractFeatures()
	if err != nil {
		log.Printf("Error parsing geojson file: %v\n", err)
	}
	etl.Close()
}