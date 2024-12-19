package db

import (
	"database/sql"
	"music-recommender/types"

	_ "github.com/mattn/go-sqlite3" // Used to import the side effects of a package. Allows for SQlit3 Driver to be known
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
	id INTEGER NOT NULL PRIMARY KEY,
)`


type MusicDB struct {
	db *sql.DB
}


func CreateSQLiteStorage() *MusicDB{
	db, err := sql.Open("sqlite3", "./music.db")
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	_, err = db.Exec(createUserTable)
	if err != nil {
		log.Fatal().AnErr("Create user table fail", err)
	}
	_, err = db.Exec(createMusicTable)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	db.Exec(createRankingTable)
	db.Exec(createTodaysRankingTable)
	mdb := MusicDB{db}
	return &mdb
}

func (mdb MusicDB) CreateNewCurator(musicEntry MusicEntry) error{
	return nil
}

func (mdb MusicDB) InsertNewSong(musicEntry *types.SubmitSong){
	const executeString = `INSERT INTO music(name, songURL, genre, subgenre, description, submitterID) 
	VALUES(?, ?, ?, ?, ?, ?)`
	_, err := mdb.db.Exec(executeString, musicEntry.Name, musicEntry.SongURL, 
		musicEntry.Genre, musicEntry.Subgenre, musicEntry.Description, 20)
	handleMusicDBError(err)
}

func (mdb MusicDB) GetTodaysRanking() *types.TodaysRankingPayload{
	res := mdb.db.QueryRow("SELECT * FROM todaysRanking WHERE insert_date = ?")
	res.Scan()
	return &types.TodaysRankingPayload{}
}

func (mdb MusicDB) UpdateTodaysRanking(submitVote types.SubmitVotePayload){
	_, err := mdb.db.Exec("UPDATE todaysRanking SET num_votes = num_votes + 1 WHERE name = ?", 
							submitVote.SongName)
	handleMusicDBError(err)
}

func (mdb MusicDB) GetTodaysMusic() *types.TodaysMusicPayload{
	return &types.TodaysMusicPayload{}
}

func (mdb MusicDB) GetCalendarsMusic() *types.CalendarPayload{
	return &types.CalendarPayload{}
}

func handleMusicDBError(err error){
	if err != nil{
		log.Error().AnErr("musicDB", err)
	}
}
