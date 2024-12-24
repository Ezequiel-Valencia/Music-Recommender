package db

import (
	"database/sql"
	"fmt"
	"music-recommender/utils"
	"net/http"
	"time"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3" // Used to import the side effects of a package. Allows for SQlit3 Driver to be known
	"github.com/rs/zerolog/log"
)

type AbstractDB struct {
	db *sql.DB
}

func CreateSQLiteStorage() (*AbstractDB, *sql.DB) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
    "password=%s dbname=%s sslmode=disable",
    "localhost", 5432, "postgres", "passwd", "postgres")
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	createTables(db)
	return &AbstractDB{db}, db
}

func (adb AbstractDB) SessionTokenIsValid(username string) (bool, error) {
	adb.db.Exec("")
	return false, nil
}
