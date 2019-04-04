package server

import (
    _ "encoding/json"
    "log"
    "net/http"
    "github.com/gorilla/mux"
    "encoding/json"
    "strconv"
    "fmt"
    "errors"
)

type APIResponse struct {
    Msg string `json:"msg"`
    AvgHeight float32 `json:"avg_height"`
}

func (resp * APIResponse) MarshalJSON() ([]byte, error) {
  if resp == nil {
    err := errors.New("attempt to marshal nil api response")
    return nil, err
  }

  json := fmt.Sprintf(`{"msg": "%s", "avg_height":%f}`,
    resp.Msg, resp.AvgHeight)

  return []byte(json), nil
}

// our main function
func ServeAPI(port int) {
    router := mux.NewRouter().StrictSlash(true)
    router.HandleFunc("/building", GetBuilding).Methods("GET")
    router.HandleFunc("/avg_height_between_years", GetAvgHeightYears).Methods("GET")
    err := http.ListenAndServe(fmt.Sprintf(":%d", port), router)
    if err != nil {
        log.Fatalf("http listen and serve error: %v\n", err)
    }

}


func GetBuilding(w http.ResponseWriter, r *http.Request) {
    w.Header().Add("Content-Type", "application/json")

	id_str := r.FormValue("doitt_id")
    doitt_id, err := strconv.Atoi(id_str)
    if err != nil {
        empty_feat := &DBFeature{}
        empty_feat.Msg = fmt.Sprintf("Invalid doitt_id: %s should be an integer", id_str)
        json.NewEncoder(w).Encode(empty_feat)
        return
    }

    db := ConnectToDB()
    defer db.Close()

    db_feature, noRows, err := QueryByDoittID(db, doitt_id)
    if noRows {
        empty_feat := &DBFeature{}
        empty_feat.Msg = fmt.Sprintf("no rows found matching doitt_id=%d", doitt_id)
        json.NewEncoder(w).Encode(empty_feat)
        return
    }
    if err != nil {
        log.Println("DB access error: %v", err)
        fmt.Println("zulu")
    }


	json.NewEncoder(w).Encode(db_feature)
}

func GetAvgHeightYears(w http.ResponseWriter, r *http.Request) {
    w.Header().Add("Content-Type", "application/json")

    yearMin, err := strconv.Atoi(r.FormValue("min"))
    if err != nil {
        resp := &APIResponse{}
        resp.Msg = "Invalid value for min (should be an integer)"
        json.NewEncoder(w).Encode(resp)
        return
    }
    yearMax, err := strconv.Atoi(r.FormValue("max"))
    if err != nil {
        resp := &APIResponse{}
        resp.Msg = "Invalid value for max (should be an integer)"
        json.NewEncoder(w).Encode(resp)
        return
    }

    db := ConnectToDB()
    defer db.Close()

    avg_height, err := QueryAvgHeightBetweenYears(db, yearMin, yearMax)
    if err != nil {
        fmt.Printf("query problem %v", err)
        resp := &APIResponse{}
        resp.Msg = fmt.Sprintf("Query received error: %v", err)
        resp.AvgHeight = 0
        json.NewEncoder(w).Encode(resp)
        return
    }
    resp := &APIResponse{}
    resp.Msg = ""
    resp.AvgHeight = avg_height

    json.NewEncoder(w).Encode(resp)
}