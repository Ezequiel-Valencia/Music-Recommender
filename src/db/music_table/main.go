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

func (mdb MusicTable) InsertNewSong(musicEntry *communication_types.SubmitSong, user db.User) {
	const executeString = `INSERT INTO music(name, artist, songURL, genre, subgenre, description, submitter_id) 
	VALUES($1, $2, $3, $4, $5, $6, $7)`
	_, err := mdb.db.Exec(executeString, musicEntry.Name, musicEntry.Artist, musicEntry.SongURL,
		musicEntry.Genre, musicEntry.Subgenre, musicEntry.Description, user.UserId)
	if err != nil {
		log.Err(err).Msg("Can't insert song.")
	}
}




