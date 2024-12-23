package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3" // Used to import the side effects of a package. Allows for SQlit3 Driver to be known
	"github.com/rs/zerolog/log"
)

type AbstractDB struct {
	db *sql.DB
}

func CreateSQLiteStorage() (*AbstractDB, *sql.DB) {
	db, err := sql.Open("sqlite3", "./music.db")
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
