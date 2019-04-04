package server

import (
    _ "encoding/json"
    "log"
    "net/http"
    "github.com/gorilla/mux"
    "encoding/json"
    "strconv"
    "fmt"
)

// our main function
func ServeAPI(port int) {
    router := mux.NewRouter().StrictSlash(true)
    router.HandleFunc("/building", GetBuilding).Methods("GET")
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
    }

	json.NewEncoder(w).Encode(db_feature)
}