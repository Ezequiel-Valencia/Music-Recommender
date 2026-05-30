package music_table

import (
	"database/sql"
	"errors"
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

// TODO: Check if the song URL is already present, and if so don't insert song again. 
func (mdb MusicTable) InsertNewSong(musicEntry *communication_types.SubmitSong, user auth_types.User) int {
	const executeString = `INSERT INTO music(insert_date, name, artist, songURL, genre, subgenre, submitter_id) 
	VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	var songID int
	err := mdb.db.QueryRow(executeString, time.Now().Format(config.StaticEnvs.TimeFormat), musicEntry.Name, musicEntry.Artist, musicEntry.SongURL,
		musicEntry.Genre, musicEntry.Subgenre, user.UserId).Scan(&songID)
	if err != nil {
		log.Err(err).Msg("Can't insert song.")
		return -1
	}
	return songID
}

func (mdb MusicTable) InsertSongSet(songSet *communication_types.SubmitSongSet, curator auth_types.User) error {
	timeInserted := time.Now()

	
	var hitLimit bool
	if err := mdb.db.QueryRow(`SELECT reached_submit_limit($1, $2)`, curator.UserId, curator.UserRole.GetRolesSubmissionLimit()).Scan(&hitLimit); err != nil {
		log.Err(err).Msg("Failed to check submission limit.")
	}

	if (hitLimit){
		return errors.New("user has reached song submission limit")
	}



	for _, song := range songSet.Songs{
		songID := mdb.InsertNewSong(&song, curator)
		if _, err := mdb.db.Exec(`INSERT INTO toBeRanked(song_id, description, curator_id, date_submitted)
		VALUES($1, $2, $3, $4)`, songID, songSet.Description, curator.UserId, timeInserted.Format(config.StaticEnvs.TimeFormat)); err != nil {
			log.Err(err).Msg("Can't queue song for ranking.")
		}
	}
	return nil
}

func (mdb MusicTable) GetUserSubmissionsToBeRanked(user auth_types.User) []communication_types.SubmitSongSet {
	result, _ := mdb.db.Query(`SELECT toBeRanked.description, 
			music.name, music.artist, music.songURL 
	FROM toBeRanked
	INNER JOIN music ON music.id = toBeRanked.song_id 
	WHERE toBeRanked.curator_id = $1;
	`, user.UserId)

	var submissions []communication_types.SubmitSongSet = []communication_types.SubmitSongSet{}

	for {
		var songSet communication_types.SubmitSongSet = communication_types.SubmitSongSet{}
		for range 3{
			if (!result.Next()){
				return submissions
			}
			var description, name, artist, songURL string
			if err := result.Scan(&description, &name, &artist, &songURL); err != nil {
				log.Err(err).Msg("Problem scanning song submission.")
				return submissions
			}
			songSet.Description = description
			songSet.Songs = append(
				songSet.Songs, 
				communication_types.SubmitSong{Name: name, Artist: artist, SongURL: songURL},
			)
		}
		submissions = append(submissions, songSet)
	}
}

