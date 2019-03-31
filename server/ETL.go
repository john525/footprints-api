package server

import (
    "fmt"
    "log"
    "errors"
    "sync"

    "io/ioutil"
    "encoding/json"

    "database/sql"
)

type ETL struct {
    db *sql.DB
    mutex sync.Mutex
}

func NewETL() (etl *ETL) {
    etl = &ETL{}
    etl.db = ConnectToDB();
    return etl;
}

func (etl * ETL) Close() {
    etl.db.Close()
    log.Println("Database connection closed.")
}

func LoadFeature(data interface{}) (added bool, err error) {
    added = false

    feature, ok := data.(Feature)
    if !ok {
        err = errors.New("Unable to extract feature.")
        return
    }

    if feature.Type != "Feature" {
        msg := fmt.Sprintf("Invalid type attribute: %v.", feature.Type)
        err = errors.New(msg)
        return
    }

    fmt.Printf("geotype %v last: %v\n", feature.Geometry.Type, feature.Properties.LastMod)

    added = true
    return
}

func (etl * ETL) ExtractFeatures() (err error) {
    // The geojson file is stored as a single object of type "FeatureCollection"
    // whose fields contain its type name and a single array (named "features").

    data, err := ioutil.ReadFile("data.geojson")
    if err != nil {
        return errors.New("Unable to read geojson file.")
    }

    featureCollection := GeoJSON{}
    err = json.Unmarshal(data, &featureCollection)
    if err != nil {
        return errors.New("Unable to read geojson file.")
    }

    objType := featureCollection.Type
    features := featureCollection.Features

    if objType != "FeatureCollection" {
        msg := fmt.Sprintf("GeoJSON object has unexpected type %v.", objType)
        return errors.New(msg)
    }

    for i := 0; i < len(features); i++ {
        _, err = LoadFeature(features[i])
        if err != nil {
            log.Printf("LoadFeature err: %v\n", err)
        }
    }

    return nil
}

