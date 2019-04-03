package server

import (
  "fmt"
  "log"
  "time"
  _ "strings"
  _ "strconv"

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

type DBFeature struct {
  ID int
  DoittID int
  Year int
  LastMod time.Time
  RoofHeight float32
  X float64
  Y float64
}

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

func InsertIntoDB(db *sql.DB, feature * DBFeature) error {
  // 4269 represents NAD83 spatial reference system
  point := fmt.Sprintf("ST_GeometryFromText('POINT (%f %f)', 4269)", feature.X, feature.Y)

  //layout := "2006-01-02T15:04:05.000Z"
  format := "2006-01-02:04:05-0700"
  lastmod := fmt.Sprintf("TIMESTAMP WITH TIME ZONE '%s'", feature.LastMod.Format(format))

  query := fmt.Sprintf(`INSERT INTO BUILDINGS (DOITT_ID, YEAR, ROOF_HEIGHT, LASTMOD, COORDS)
    VALUES (%d, %d, %f, %s, %s)`, feature.DoittID,feature.Year, feature.RoofHeight,
    lastmod, point)
  _, err := db.Exec(query)
  if err != nil {
    return fmt.Errorf("Error inserting (%s) into db: %s", query, err)
  }

  return nil
}

func UpdateDBEntry(db *sql.DB, db_id int, feature * DBFeature) error {
  // 4269 represents NAD83 spatial reference system
  point := fmt.Sprintf("ST_GeometryFromText('POINT (%f %f)', 4269)", feature.X, feature.Y)

  //layout := "2006-01-02T15:04:05.000Z"
  format := "2006-01-02:04:05-0700"
  lastmod := fmt.Sprintf("TIMESTAMP WITH TIME ZONE '%s'", feature.LastMod.Format(format))

  query := fmt.Sprintf(`UPDATE BUILDINGS SET DOITT_ID=%d YEAR=%d ROOF_HEIGHT=%f LASTMOD=%s
    COORDS=%s WHERE ID=%d`, feature.DoittID,feature.Year, feature.RoofHeight,
    lastmod, point, db_id)
  _, err := db.Exec(query)
  if err != nil {
    return fmt.Errorf("Error updating (%s) into db: %s", query, err)
  }

  return nil
}

func QueryByDoittID(db *sql.DB, doitt_id int) (feature * DBFeature, err error) {
  query := fmt.Sprintf(`SELECT ID, DOITT_ID, YEAR, LASTMOD, ROOF_HEIGHT, ST_X(COORDS), ST_Y(COORDS)
    FROM BUILDINGS WHERE DOITT_ID=%d`, doitt_id)
  
  feature = &DBFeature{}

  err = db.QueryRow(query).Scan(
    &feature.ID, &feature.DoittID, &feature.Year, &feature.LastMod,
    &feature.RoofHeight, &feature.X, &feature.Y)

  if err != nil {
    return nil, fmt.Errorf("SELECT QUERY error: %v", err)
  }

  err = nil
  return
}