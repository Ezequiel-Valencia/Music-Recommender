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
	id INTEGER NOT NULL PRIMARY KEY,
	insert_date DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	name TEXT,
	songURL TEXT,
	genre TEXT,
	subgenre TEXT,
	description TEXT,
	submitterID,
	rank_id TEXT
	num_ranks INTEGER,
	FOREIGN KEY (submitterID) REFERENCES users(id)
)`

/*
Ranking can include multiple songs that have been ranked, thus it will be
a string that has specific format easy for tokenization, with each token
being a reference to a music ID
*/

const createRankingTable string = `CREATE TABLE IF NOT EXISTS ranking (
	id INTEGER NOT NULL PRIMARY KEY,
	date_ranked DATETIME NOT NULL,
	ranking TEXT
)`

/*
Everyday take the total of yesterdays music choices (if there is any)
Sum up the choices and put them in the ranking table as appropriate
Clean the table, and place the new songs which will be ranked within the table
*/
const createTodaysRankingTable string = `CREATE TABLE IF NOT EXISTS todaysRanking (
	songID INTEGER,
	name TEXT,
	num_votes INTEGER,
	FOREIGN KEY (songID) REFERENCES music(id)
)`

const createUserTable string = `CREATE TABLE IF NOT EXISTS users (
	user_id INTEGER NOT NULL PRIMARY KEY,
	username TEXT,
	passwd_hash TEXT,
	subject_identifier TEXT,
	creation_source TEXT,
	creation_date DATETIME NOT NULL,
	user_role TEXT,
	user_privileges TEXT
)`

const createSessionIDTable string = `CREATE TABLE IF NOT EXISTS sessions (
	user_id,
	session_id TEXT,
	FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
)`




func createTables(db *sql.DB){
	tables := [...]string{createUserTable, createMusicTable, createSessionIDTable, createRankingTable, createTodaysRankingTable}
	for _, v := range tables{
		_, err := db.Exec(v)
		if err != nil{
			log.Fatal().AnErr("Table Creation Error", err)
		}
	}
}



