package db

import (
	"database/sql"
	"music-recommender/config"
	"time"

	"github.com/rs/zerolog/log"
)

// Good enough table SCHEMA for now

// User Identity instead of serial for UID's cause it's SQL compliant
// https://stackoverflow.com/questions/55300370/postgresql-serial-vs-identity
const createUserTable string = `CREATE TABLE IF NOT EXISTS users (
	user_id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1),
	username TEXT NOT NULL,
	email TEXT NOT NULL,
	password_hash TEXT NOT NULL,
	subject_identifier TEXT,
	creation_source TEXT NOT NULL,
	creation_date TIMESTAMP NOT NULL,
	user_role TEXT NOT NULL,
	user_privileges TEXT NOT NULL,
	song_sets_submitted INTEGER NOT NULL DEFAULT(0),
	last_vote TIMESTAMP
)`

/*
rank_id is a comma separated list of the foreign keys associated with ranking table

num_ranks is the number of times a specific song has been ranked.
Useful in case a special day is made where previous songs can be ranked again
(ex. Christmas songs on christmas, where instead of three it's every submitters favorite christmas song)
*/
const createMusicTable string = `CREATE TABLE IF NOT EXISTS music (
	id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1),
	insert_date TIMESTAMP NOT NULL,
	name TEXT NOT NULL,
	artist TEXT NOT NULL,
	songURL TEXT NOT NULL,
	genre TEXT,
	subgenre TEXT,
	submitter_id INTEGER references users(user_id),
	rank_id TEXT,
	num_ranks INTEGER
)`


const createToBeRankedTable = `CREATE TABLE IF NOT EXISTS toBeRanked (
	id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1),
	song_id INTEGER references music(id),
	description TEXT NOT NULL,
	curator_id INTEGER references users(user_id),
	date_submitted TIMESTAMP NOT NULL
)`

/*
Ranking can include multiple songs that have been ranked, thus it will be
a string that has specific format easy for tokenization, with each token
being a reference to a (music ID, rank, 

Better yet assign the same day to all rankings 
*/

const createRankingTable string = `CREATE TABLE IF NOT EXISTS ranked (
	id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1),
	song_id INTEGER references music(id),
	curator_id INTEGER references users(user_id),
	date_ranked TIMESTAMP NOT NULL,
	num_votes INTEGER,
	winner BOOLEAN
)`

/*
Everyday take the total of yesterdays music choices (if there is any)
Sum up the choices and put them in the ranking table as appropriate
Clean the table, and place the new songs which will be ranked within the table
*/
const createTodaysRankingTable string = `CREATE TABLE IF NOT EXISTS todaysRanking (
	song_id INTEGER references music(id),
	curator_id INTEGER references users(user_id),
	description TEXT NOT NULL,
	song_name TEXT NOT NULL,
	song_artist TEXT NOT NULL,
	song_path_resource TEXT NOT NULL,
	song_order INTEGER NOT NULL,
	num_votes INTEGER DEFAULT 0
)`



const createSessionIDTable string = `CREATE TABLE IF NOT EXISTS sessions (
	entry INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1),
	user_id INTEGER references users(user_id) ON DELETE CASCADE,
	session_id TEXT NOT NULL,
	csrf_token TEXT NOT NULL,
	creation_date TIMESTAMP NOT NULL
)`

const serverState string = `CREATE TABLE IF NOT EXISTS server_state (
	single_row INTEGER PRIMARY KEY NOT NULL UNIQUE DEFAULT 1,
	allow_user_creation BOOLEAN NOT NULL,
	update_date TIMESTAMP NOT NULL,
	CONSTRAINT single_row_uni CHECK (single_row = 1)
); REVOKE DELETE, TRUNCATE ON server_state FROM public;`

// 0 for table primary key is special value, will not be used and can be assumed as NULL
func CreateTablesAndFunctions(db *sql.DB, testMode bool) error {
	tables := [...]string{createUserTable, createMusicTable, createSessionIDTable,
		createRankingTable, createTodaysRankingTable, serverState, createToBeRankedTable}
	functions := [...]string{hasUserSubmitCountHitLimit}
	createDBHelper(db, testMode, tables[:], "Tables")
	createDBHelper(db, testMode, functions[:], "Functions")
	log.Info().Msg("Created Tables and Functions")
	return nil
}

func createDBHelper(db *sql.DB, testMode bool, creationSet []string, creationType string) error{
	for _, v := range creationSet {
		_, err := db.Exec(v)
		if err != nil {
			if testMode {
				log.Err(err).Msgf("%s creation error.", creationType)
				return err
			}
			log.Fatal().AnErr("err", err).Msgf("%s Creation Error", creationType)
		}
	}
	return nil
}

func initializeOrGetServerState(db *sql.DB) {
	res, _ := db.Exec("SELECT * FROM server_state")
	if res == nil{
		return
	}
	resNum, _ := res.RowsAffected()
	if resNum > 1 {
		log.Fatal().Msg("There is more than one row for server state.")
	} else if resNum == 1{
		db.QueryRow("SELECT allow_user_creation FROM server_state").Scan(&config.DynamicEnvs.AllowUserCreation)
	} else if resNum == 0{
		db.Exec(`INSERT INTO server_state(allow_user_creation, update_date)
		VALUES($1, $2)`, true, time.Now().UTC().Format(config.StaticEnvs.TimeFormat))
		return
	}
}
