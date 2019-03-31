package server

type GeoJSON struct {
    Type string `json:"type"`
    Features []Feature `json:"features"`
}

type Feature struct {
    Type   string  `json:"type"`
    Properties Props `json:"properties"`
    Geometry MultiPolygon `json:"geometry"`
}

type Props struct {
    // `json:"shape_area"`
    // `json:"shape_len"`
    // `json:"feat_code"`
    // `json:"doitt_id"`
    // `json:"groundelev"`
    LastMod string `json:"lstmoddate"`
}

type MultiPolygon struct {
    Type string `json:"type"`
    Coordinates [][][][]float64 `json:"coordinates"`
}