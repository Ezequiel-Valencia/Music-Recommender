package db

import (
	"database/sql"

	"github.com/rs/zerolog/log"
)

// Good enough table SCHEMA for now

/*
rank_id is a comma separated list of the foreign keys associated with ranking table

num_ranks is the number of times a specific song has been ranked.
Useful in case a special day is made where previous songs can be ranked again
(ex. Christmas songs on christmas, where instead of three it's every submitters favorite christmas song)
*/
const createMusicTable string = `CREATE TABLE IF NOT EXISTS music (
	id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1),
	insert_date TIMESTAMP NOT NULL,
	name TEXT,
	songURL TEXT,
	genre TEXT,
	subgenre TEXT,
	description TEXT,
	submitterID INTEGER references users(user_id),
	rank_id TEXT,
	num_ranks INTEGER
)`

/*
Ranking can include multiple songs that have been ranked, thus it will be
a string that has specific format easy for tokenization, with each token
being a reference to a music ID
*/

const createRankingTable string = `CREATE TABLE IF NOT EXISTS ranking (
	id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1),
	date_ranked TIMESTAMP NOT NULL,
	ranking TEXT
)`

/*
Everyday take the total of yesterdays music choices (if there is any)
Sum up the choices and put them in the ranking table as appropriate
Clean the table, and place the new songs which will be ranked within the table
*/
const createTodaysRankingTable string = `CREATE TABLE IF NOT EXISTS todaysRanking (
	songID INTEGER references music(id),
	name TEXT,
	num_votes INTEGER
)`

// User Identity instead of serial for UID's cause it's SQL compliant
// https://stackoverflow.com/questions/55300370/postgresql-serial-vs-identity
const createUserTable string = `CREATE TABLE IF NOT EXISTS users (
	user_id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1),
	username TEXT,
	password_hash TEXT,
	subject_identifier TEXT,
	creation_source TEXT,
	creation_date TIMESTAMP NOT NULL,
	user_role TEXT,
	user_privileges TEXT
)`

const createSessionIDTable string = `CREATE TABLE IF NOT EXISTS sessions (
	entry INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1),
	user_id INTEGER references users(user_id) ON DELETE CASCADE,
	session_id TEXT NOT NULL,
	csrf_token TEXT NOT NULL,
	creation_date TIMESTAMP NOT NULL
)`



// 0 for table primary key is special value, will not be used and can be assumed as NULL
func CreateTables(db *sql.DB, testMode bool) error{
	tables := [...]string{createUserTable, createMusicTable, createSessionIDTable, createRankingTable, createTodaysRankingTable}
	for _, v := range tables{
		_, err := db.Exec(v)
		if err != nil{
			if testMode{return err}
			log.Fatal().AnErr("err", err).Msg("Table Creation Error")
		}
	}
	log.Info().Msg("Created Tables")
	return nil
}



