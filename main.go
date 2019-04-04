package main

import (
	"log"
	"os"
	"time"
	"fmt"

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

	url := "https://data.cityofnewyork.us/api/geospatial/nqwf-w8eh?method=export&format=GeoJSON"
	fname := "data.geojson"

	// NYC Open Data's public version of Building Footprints is updated
	// once every week.
	updateFreq := 7 * 24 * time.Hour
	retryFreq := 500 * time.Millisecond

	etl := server.NewETL(url, fname, updateFreq, retryFreq);
	go etl.DoETL()
	
	fmt.Println("Footprints Server. Type 'help' for CLI instructions.")
	for true {
		fmt.Print(">")
		var input string
    	_, err = fmt.Scanln(&input)
    	if err != nil {
    		log.Printf("CLI Error: %v\n", err)
    	} else if input == "help" {
    		fmt.Println("help - get instructions")
    		fmt.Println("exit - stop server")
    		fmt.Println("last - print timestamp of most recent data extraction")
    		fmt.Println("status - print ETL process's current activity (one of: download, extract, transform/load, or wait)")
    	} else if input == "exit" {
    		etl.StopETL()
    		os.Exit(0)
    	} else if input == "last" {
    		etl.PrintLast()
    	} else if input == "status" {
    		etl.PrintStatus()
    	} else {
    		fmt.Println("Invalid command.")
    	}
	}
}