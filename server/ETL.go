package server

import (
    "fmt"
    "log"
    "errors"
    "sync"

    "io/ioutil"
    "encoding/json"
    "time"
    "strconv"

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

func (etl *ETL) LoadFeature(data interface{}) (added bool, err error) {
    added = false

    feature, ok := data.(Feature)
    if !ok {
        err = errors.New("Unable to extract feature.")
        return
    }

    if feature.Type != "Feature" {
        err = fmt.Errorf("Invalid type attribute: %v.", feature.Type)
        return
    }

    // Validate feature timestamp
    layout := "2006-01-02T15:04:05.000Z"
    str := feature.Properties.Lstmoddate
    t, err := time.Parse(layout, str)
    if err != nil {
        err = fmt.Errorf("Could not parse last mod date (%v)", str)
        return
    }

    //Validate other props
    doitt_id, err := strconv.Atoi(feature.Properties.Doitt_id)
    if err != nil {
        err = fmt.Errorf("Could not parse doitt_id (%v)", err)
        return
    }
    year, err := strconv.Atoi(feature.Properties.Cnstrct_yr)
    if err != nil {
        err = fmt.Errorf("Could not parse constructed year (%v)", err)
        return
    }
    height, err := strconv.ParseFloat(feature.Properties.Heightroof, 32)
    if err != nil {
        err = fmt.Errorf("Could not parse height of roof (%v)", err)
        return
    }

    newRow := &DBFeature{}
    newRow.DoittID = doitt_id
    newRow.Year = year
    newRow.LastMod = t
    newRow.RoofHeight = float32(height)
    newRow.X = feature.Geometry.Coordinates.X
    newRow.Y = feature.Geometry.Coordinates.Y

    
    oldFeature, err := QueryByDoittID(etl.db, doitt_id)
    if err != nil {
        log.Printf("Error from SELECT query: %v\n", err)
    }
    if oldFeature != nil {
        u := oldFeature.LastMod
        if u.Before(t) {
            err = UpdateDBEntry(etl.db, oldFeature.ID, newRow)
            if err != nil {
                return
            }
        } else {
            err = errors.New("Cannot overwrite database with older data.")
            return
        }

    } else {
        err = InsertIntoDB(etl.db, newRow)
        if err != nil {
            return
        }
    }

    added = true
    err = nil
    return
}

func (etl * ETL) ExtractFeatures() (err error) {
    // The geojson file is stored as a single object of type "FeatureCollection"
    // whose fields contain its type name and a single array (named "features").

    data, err := ioutil.ReadFile("data.geojson")
    if err != nil {
        return fmt.Errorf("Unable to read geojson file (%v).", err)
    }

    featureCollection := GeoJSON{}
    err = json.Unmarshal(data, &featureCollection)
    if err != nil {
        return fmt.Errorf("Unable to unmarshal geojson (%v).", err)
    }

    objType := featureCollection.Type
    features := featureCollection.Features

    if objType != "FeatureCollection" {
        return fmt.Errorf("GeoJSON object has unexpected type %v.", objType)
    }

    for i := 0; i < len(features); i++ {
        _, err = etl.LoadFeature(features[i])
        if err != nil {
            log.Printf("LoadFeature err: %v\n", err)
        }
    }

    return nil
}

