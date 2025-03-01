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
	mdb.db.QueryRow(`SELECT reached_submit_limit($1, $2)`, curator.UserId, curator.UserRole.GetRolesSubmissionLimit()).Scan(&hitLimit)

	if (hitLimit){
		log.Warn().Msgf("User %s has hit their submission limit", curator.Username)
		return errors.New("user has reached song submission limit")
	}

	var descriptionID int
	const submissionInsertion = `INSERT INTO submissionDescriptions(description) VALUES($1) RETURNING id`
	mdb.db.QueryRow(submissionInsertion, songSet.Description).Scan(&descriptionID)

	for _, song := range songSet.Songs{
		songID := mdb.InsertNewSong(&song, curator)
		_, err := mdb.db.Exec(`INSERT INTO toBeRanked(song_id, description_id, curator_id, date_submitted)
		VALUES($1, $2, $3, $4)`, songID, descriptionID, curator.UserId, timeInserted.Format(config.StaticEnvs.TimeFormat))
		if (err != nil){
			log.Err(err).Msg("Problem inserting song set to be ranked.")
			return err
		}
	}
	return nil
}

func (mdb MusicTable) GetUserSubmissionsToBeRanked(user auth_types.User) []communication_types.SubmitSongSet {
	result, err := mdb.db.Query(`SELECT des.description, 
			music.name, music.artist, music.songURL
	FROM toBeRanked
	INNER JOIN music ON music.id = toBeRanked.song_id 
	INNER JOIN submissionDescriptions des ON des.id = toBeRanked.description_id
	WHERE toBeRanked.curator_id = $1;
	`, user.UserId)

	if (err != nil){
		log.Err(err).Msg("Can't get user sets to be ranked.")
		return []communication_types.SubmitSongSet{}
	}

	var submissions []communication_types.SubmitSongSet = []communication_types.SubmitSongSet{}

	for {
		var songSet communication_types.SubmitSongSet = communication_types.SubmitSongSet{}
		for range 3{
			if (!result.Next()){
				return submissions
			}
			var description, name, artist, songURL string
			result.Scan(&description, &name, &artist, &songURL)
			songSet.Description = description
			songSet.Songs = append(
				songSet.Songs, 
				communication_types.SubmitSong{Name: name, Artist: artist, SongURL: songURL},
			)
		}
		submissions = append(submissions, songSet)
	}
}

