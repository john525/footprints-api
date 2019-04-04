package server

import (
    "fmt"
    "log"
    "errors"

    "net/http"
    "io"
    "os"
    "io/ioutil"
    "encoding/json"
    "time"
    "strconv"

    "database/sql"
)

type Status int
const (
    Download Status = 0
    Extract Status = 1
    TransformLoad Status = 2
    Waiting Status = 3
)

type ETL struct {
    db *sql.DB
    fname string
    url string
    updateFreq time.Duration
    retryFreq time.Duration
    last *time.Time
    stopETL chan bool
    downloadDone chan bool
    extractDone chan bool
    status Status
}

func NewETL(url string, fname string, updateFreq time.Duration, retryFreq time.Duration) (etl *ETL) {
    etl = &ETL{}
    etl.fname = fname
    etl.url = url
    etl.updateFreq = updateFreq
    etl.retryFreq = retryFreq
    etl.db = ConnectToDB();
    etl.stopETL = make(chan bool)
    etl.last = nil
    etl.downloadDone = make(chan bool)
    etl.extractDone = make(chan bool)
    return etl;
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func (etl * ETL) DownloadFile(filepath string, url string) error {

    // Get the data
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // Create the file
    out, err := os.Create(filepath)
    if err != nil {
        return err
    }
    defer out.Close()

    // Write the body to file
    _, err = io.Copy(out, resp.Body)
    return err
}

func (etl * ETL) DoETL() {
    timer := time.NewTimer(etl.updateFreq)

    etlLoop:
    for {
        downloadLoop:
        for {
           etl.status = Download
            go etl.DownloadSourceData()

            select {
            case <-etl.stopETL:
                return
            case success := <-etl.downloadDone:
                if !success {
                    log.Printf("Could not download from %s.\n", etl.url)
                } else {
                    break downloadLoop
                }
            }

            time.Sleep(etl.retryFreq)
        }

        extractLoop:
        for {
            etl.status = Extract
            go etl.ExtractFeatures()

            select {
            case <-etl.stopETL:
                return
            case success := <-etl.extractDone:
                if !success {
                    log.Printf("Could not extract features from %s.\n", etl.fname)
                } else {
                    break extractLoop
                }
            }

            time.Sleep(etl.retryFreq)
        }

        etl.status = Waiting
        select {
        case <-etl.stopETL:
            if !timer.Stop() {
                <-timer.C
            }
            return
        case <-timer.C:
            continue etlLoop
        }
    }
}

func (etl * ETL) DownloadSourceData() {
    _, err := os.Stat(etl.fname)
    if err == nil {
        err := os.Remove(etl.fname)
        if err != nil {
            log.Printf("Remove file error %v\n", err)
            etl.downloadDone <- false
            return
        }
    } else if err != os.ErrNotExist {
        log.Printf("Find file error %v\n", err)
        etl.downloadDone <- false
        return
    }

    err = etl.DownloadFile(etl.fname, etl.url)
    if err != nil {
        log.Printf("Download file error %v\n", err)
        etl.downloadDone <- false
        return
    }


    etl.downloadDone <- true
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

    
    oldFeature, noRows, err := QueryByDoittID(etl.db, doitt_id)
    if err != nil && !noRows {
        log.Printf("Error from SELECT query: %v\n", err)
    }
    if noRows {
        err = InsertIntoDB(etl.db, newRow)
        if err != nil {
            return
        }
    } else {
        u := oldFeature.LastMod
        if u.Before(t) {
            err = UpdateDBEntry(etl.db, oldFeature.ID, newRow)
            if err != nil {
                return
            }
        } else {
            log.Println("Ignoring out-of-date feature data.")
            return
        }
    }

    added = true
    err = nil
    return
}

func (etl * ETL) ExtractFeatures() {
    // The geojson file is stored as a single object of type "FeatureCollection"
    // whose fields contain its type name and a single array (named "features").

    data, err := ioutil.ReadFile(etl.fname)
    if err != nil {
        log.Printf("Unable to read geojson file (%v).\n", err)
        etl.extractDone <- false
        return
    }

    featureCollection := GeoJSON{}
    err = json.Unmarshal(data, &featureCollection)
    if err != nil {
        log.Printf("Unable to unmarshal geojson (%v).\n", err)
        etl.extractDone <- false
        return
    }

    objType := featureCollection.Type
    features := featureCollection.Features

    if objType != "FeatureCollection" {
        log.Printf("GeoJSON object has unexpected type %v.\n", objType)
        etl.extractDone <- false
        return
    }

    for i := 0; i < len(features); i++ {
        _, err = etl.LoadFeature(features[i])
        if err != nil {
            log.Printf("LoadFeature err: %v\n", err)
            // no need to cancel whole extraction process just because one feature failed
        }
    }

    etl.extractDone <- true

}

func (etl * ETL) StopETL() {
    etl.Close()
    etl.stopETL <- true
}

func (etl * ETL) PrintLast() {
    if etl.last == nil {
        fmt.Println("ETL has not yet loaded any data.")
    } else {
        fmt.Println("Last extraction, transformation, and load occurred at %v.", etl.last)
    }
}

func (etl * ETL) PrintStatus() {
    fmt.Println("ETL STATUS:")
    if etl.status == Download {
        fmt.Println("Downloading geojson from NYC Open Data.")
    } else if etl.status == Extract {
        fmt.Println("Extraction phase: unmarshalling local geojson file.")
    } else if etl.status == TransformLoad {
        fmt.Println("Transform & Load phase: modifying data and inserting into postgres.")
    } else if etl.status == Waiting {
        fmt.Println("Waiting for next ETL cycle. Ready to serve API requests.")
    }
}
