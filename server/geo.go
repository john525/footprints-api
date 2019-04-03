package server

import (
    "encoding/json"
    "fmt"
    "errors"
)

type GeoJSON struct {
    Type string `json:"type"`
    Features []Feature `json:"features"`
}

type Feature struct {
    Type   string  `json:"type"`
    Properties Props `json:"properties"`
    Geometry Shape `json:"geometry"`
}

type Props struct {
    Cnstrct_yr string `json:"feat_code"`
    Doitt_id string `json:"doitt_id"`
    Heightroof string `json:"heightroof"`
    Lstmoddate string `json:"lstmoddate"`
}

type Shape struct {
    Type string `json:"type"`
    Coordinates Point `json:"coordinates"`
}

type Point struct {
    X float64
    Y float64
}

func (p *Point) UnmarshalJSON(data []byte) error {
    var geometry []interface{}
    if err := json.Unmarshal(data, &geometry); err != nil {
        fmt.Printf("Error while decoding multipolygon (error: %v)\n", err)
        fmt.Printf("%s\n", string(data))
        return err
    }

    //Find any point in or on the building so we can approximate its location
    for {
        if len(geometry) == 2 {
            x, ok1 := geometry[0].(float64)
            y, ok2 := geometry[1].(float64)
            if ok1 && ok2 {
                p.X = x
                p.Y = y
                break
            }
        }
        newArray, ok := geometry[0].([]interface{})
        if !ok {
            return errors.New("Could not find any floats in geometry")
        }
        geometry = newArray
    }

    return nil
}