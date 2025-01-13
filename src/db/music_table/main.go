package music_table

import (
	"database/sql"
	"music-recommender/config"
	"music-recommender/db"
	"music-recommender/types/communication_types"
	"music-recommender/types/internal_types/auth_types"
	"time"

	"github.com/rs/zerolog/log"
)

type MusicTable struct {
	db         *sql.DB
	AbstractDB *db.AbstractDB
}

func CreateMusicTableDriver(db *sql.DB, abd *db.AbstractDB) *MusicTable {
	return &MusicTable{db: db, AbstractDB: abd}
}

func (mdb MusicTable) InsertNewSong(musicEntry *communication_types.SubmitSong, user auth_types.User) {
	const executeString = `INSERT INTO music(insert_date, name, artist, songURL, genre, subgenre, submitter_id) 
	VALUES($1, $2, $3, $4, $5, $6, $7)`
	_, err := mdb.db.Exec(executeString, time.Now().Format(config.StaticEnvs.TimeFormat), musicEntry.Name, musicEntry.Artist, musicEntry.SongURL,
		musicEntry.Genre, musicEntry.Subgenre, user.UserId)
	if err != nil {
		log.Err(err).Msg("Can't insert song.")
	}
}
