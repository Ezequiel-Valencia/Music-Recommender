package music_table

import (
	"database/sql"
	"music-recommender/db"
	"music-recommender/types/communication_types"

	"github.com/rs/zerolog/log"
)



type MusicTable struct{
	db *sql.DB
	AbstractDB *db.AbstractDB
}

func CreateMusicTableDriver(db *sql.DB, abd *db.AbstractDB) *MusicTable{
	return &MusicTable{db: db, AbstractDB: abd}
}

func (mdb MusicTable) InsertNewSong(musicEntry *communication_types.SubmitSong) {
	const executeString = `INSERT INTO music(name, songURL, genre, subgenre, description, submitterID) 
	VALUES(?, ?, ?, ?, ?, ?)`
	_, err := mdb.db.Exec(executeString, musicEntry.Name, musicEntry.SongURL,
		musicEntry.Genre, musicEntry.Subgenre, musicEntry.Description, 20)
	if err != nil {
		log.Err(err).Msg("Can't insert song.")
	}
}




