package server

import (
  "fmt"
  "log"

  "database/sql"
  _ "github.com/lib/pq"
)

const (
  host     = "localhost"
  port     = 5432
  user     = "postgres"
  password = "092125" // TODO: Change postgres password before deployment.
  dbname   = "nyc_buildings"
)

// Connect to the postgres database.
// Assume the required database has already been created and configured.
func ConnectToDB() (db * sql.DB) {
  // TODO: Configure postgres with SSL support
  psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
    "password=%s dbname=%s sslmode=disable",
    host, port, user, password, dbname)

  db, err := sql.Open("postgres", psqlInfo)
  if err != nil {
    log.Fatal(err)
  }

  err = db.Ping()
  if err != nil {
    log.Panicln(err)
  }

  log.Println("Successfully connected to database.")

  return db
}