package server

import (
    _ "encoding/json"
    "log"
    "net/http"
    "github.com/gorilla/mux"
)

// our main function
func serve() {
    router := mux.NewRouter()
    router.HandleFunc("/buildings", GetBuilding).Methods("GET")
    log.Fatal(http.ListenAndServe(":8000", router))
}

func GetBuilding(w http.ResponseWriter, r *http.Request) {
	//params := mux.Vars(r)
    //doitt_id := params["doitt_id"]

	//json.NewEncoder(w).Encode(QueryByDoittID(doitt_id))
}