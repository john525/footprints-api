package server

import (
  "fmt"
  "errors"
  "log"
  "time"
  _ "strings"
  _ "strconv"
  _ "encoding/json"

  "database/sql"
  _ "github.com/lib/pq"
)

const (
  host     = "localhost"
  port     = 5432
  user     = "postgres"
  password = "092125" // TODO: Strengthen postgres password before deployment.
  dbname   = "nyc_buildings"
)

type DBFeature struct {
  ID int `json:"id"`
  DoittID int `json:"doitt_id"`
  Year int `json:"year"`
  LastMod time.Time `json:"last_mod"`
  RoofHeight float32 `json:"roof_height"`
  X float64 `json:"coord_x"`
  Y float64 `json:"coord_y"`
  Msg string `json:"msg"`
}

func (feature * DBFeature) MarshalJSON() ([]byte, error) {
  if feature == nil {
    err := errors.New("attempt to marshal nil db feature")
    return nil, err
  }

  json := fmt.Sprintf(`{"id": %d, "doitt_id": %d,
    "msg": "%s", "Year": %d, "LastMod":"%s", "RoofHeight":%f,
    "coord_x":%f, "coord_y":%f}`,
    feature.ID, feature.DoittID, feature.Msg, feature.Year,
    feature.LastMod, feature.RoofHeight, feature.X, feature.Y)

  return []byte(json), nil
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
    log.Fatalln(err) // TODO: add better error-handling
  }

  err = db.Ping()
  if err != nil {
    log.Panicln(err)
  }

  log.Println("Connecting to database...")

  return db
}

func InsertIntoDB(db *sql.DB, feature * DBFeature) error {
  // 4269 represents NAD83 spatial reference system
  point := fmt.Sprintf("ST_GeometryFromText('POINT (%f %f)', 4269)", feature.X, feature.Y)

  //layout := "2006-01-02T15:04:05.000Z"
  format := "2006-01-02:04:05-0700"
  lastmod := fmt.Sprintf("TIMESTAMP WITH TIME ZONE '%s'", feature.LastMod.Format(format))

  query := fmt.Sprintf(`INSERT INTO BUILDINGS (DOITT_ID, YEAR, ROOF_HEIGHT, LASTMOD, COORDS)
    VALUES (%d, %d, %f, %s, %s)`, feature.DoittID, feature.Year, feature.RoofHeight,
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

func QueryAvgHeightInBoundingBox(db *sql.DB, xmin float32, ymin float32, xmax float32, ymax float32) {
  // TODO: implement
  //region := fmt.Sprintf("POLYGON((%f %f, %f %f, %f %f, %f %f, %f %f))",
  //    xmin, ymin, xmin, ymax, xmax, ymax, xmax, ymin, xmin, ymin)
}

func QueryAvgHeightBetweenYears(db *sql.DB, yearMin int, yearMax int) (height float32, err error) {
  query := fmt.Sprintf(`SELECT AVG(ROOF_HEIGHT) FROM BUILDINGS WHERE
    (ROOF_HEIGHT IS NOT NULL) AND (YEAR >= %d) AND (YEAR <= %d);`, yearMin, yearMax)
  
  var avgHeight *float32
  err = db.QueryRow(query).Scan(&avgHeight)
  // if avgHeight == nil {
  //   return 0, errors.New("no data found in interval")
  // }

  if err == sql.ErrNoRows {
    return 0, fmt.Errorf("No data entries within selected interval.")
  } else if err != nil {
    return 0, fmt.Errorf("SELECT error: %v", err)
  }

  return *avgHeight, nil
  }

func QueryByDoittID(db *sql.DB, doitt_id int) (feature * DBFeature, noRows bool, err error) {
  query := fmt.Sprintf(`SELECT ID, DOITT_ID, YEAR, LASTMOD, ROOF_HEIGHT, ST_X(COORDS), ST_Y(COORDS)
    FROM BUILDINGS WHERE DOITT_ID=%d`, doitt_id)
  
  feature = &DBFeature{}

  err = db.QueryRow(query).Scan(
    &feature.ID, &feature.DoittID, &feature.Year, &feature.LastMod,
    &feature.RoofHeight, &feature.X, &feature.Y)

  if err == sql.ErrNoRows {
    return nil, true, err
  } else if err != nil {
    return nil, false, fmt.Errorf("SELECT QUERY error: %v", err)
  }

  return feature, false, nil
}