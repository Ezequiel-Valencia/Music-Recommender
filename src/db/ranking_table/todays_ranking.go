package ranking_table

import (
	"database/sql"
	"music-recommender/config"
	"music-recommender/types/communication_types"
	"music-recommender/types/internal_types"
	"music-recommender/types/internal_types/auth_types"
	"music-recommender/utils"
	"time"

	"github.com/rs/zerolog/log"
)

type TodaysRankingDriver struct {
	db *sql.DB
}


func CreateTodaysRankingDriver(db *sql.DB) *TodaysRankingDriver {
	return &TodaysRankingDriver{db: db}
}

func (mdb TodaysRankingDriver) GetTodaysVotes() communication_types.TodaysRankingPayload {
	res, err := mdb.db.Query("SELECT song_order, num_votes FROM todaysRanking")
	if err != nil {
		log.Err(err).Msg("Can't get todays ranking.")
		return communication_types.TodaysRankingPayload{}
	}
	todaysRanking := communication_types.TodaysRankingPayload{
		RankingMap: make(map[int]float64),
	}
	var totalVotes int = 0
	for res.Next() {
		var order, numVotes int
		res.Scan(&order, &numVotes)
		totalVotes += numVotes
		todaysRanking.RankingMap[order] += float64(numVotes)
	}
	for key := range 3 {
		if totalVotes == 0 {
			todaysRanking.RankingMap[key] = 0
		} else {
			todaysRanking.RankingMap[key] = todaysRanking.RankingMap[key] / float64(totalVotes)
		}
	}
	return todaysRanking
}

func (mdb TodaysRankingDriver) UserAlreadyVoteToday(user auth_types.User) bool {
	var lastVote time.Time
	mdb.db.QueryRow("SELECT last_vote FROM users WHERE user_id = $1", user.UserId).Scan(&lastVote)
	return !isYesterdayOrBefore(lastVote)
}

// Song order is based on the 0 indexing of the three songs being voted on.
func (mdb TodaysRankingDriver) UpdateTodaysVoteCount(submitVote communication_types.SubmitVotePayload, user auth_types.User) {
	_, err := mdb.db.Exec("UPDATE users SET last_vote = $1 WHERE user_id = $2",
		time.Now().Format(config.StaticEnvs.TimeFormat), user.UserId)

	if err != nil {
		log.Err(err).Msg("Can't update users last vote date.")
		return
	}

	_, err = mdb.db.Exec("UPDATE todaysRanking SET num_votes = num_votes + 1 WHERE song_order = $1",
		submitVote.SongOrder)
	if err != nil {
		log.Err(err).Msg("Update ranking did not work")
	}
}

func (mdb TodaysRankingDriver) GetTodaysMusic() *communication_types.TodaysMusicPayload {

	rows, err := mdb.db.Query(`SELECT song_id, curator_id, description, song_order, song_name, song_artist, song_path_resource
	FROM todaysRanking`)
	if err != nil {
		log.Err(err).Msg("Can't Get Todays Music")
		return &communication_types.TodaysMusicPayload{}
	}

	var musicPayload communication_types.TodaysMusicPayload
	musicPayload.MusicEntries = []communication_types.MusicPayloadEntry{}

	var curatorID int
	for rows.Next() {
		var songID, order int
		var description, songName, songArtist, songResource string
		rows.Scan(&songID, &curatorID, &description, &order, &songName, &songArtist, &songResource)
		musicPayload.CuratorDescription = description

		musicEntry := communication_types.MusicPayloadEntry{Title: songName, Artist: songArtist, SongOrder: order, PathResource: songResource}
		musicPayload.MusicEntries = append(musicPayload.MusicEntries, musicEntry)
	}
	var curatorName string
	mdb.db.QueryRow(`SELECT username FROM users WHERE user_id = $1`, curatorID).Scan(&curatorName)
	musicPayload.CuratorName = curatorName

	return &musicPayload
}

func (mdb TodaysRankingDriver) AnySongsToBeRanked() bool{
	sqlRows, _ := mdb.db.Exec(`SELECT * FROM toBeRanked`)
	res, _ := sqlRows.RowsAffected()
	return res > 0
}

// Dumb for now
func (mdb TodaysRankingDriver) SelectNewSongs() {
	var newSongListTime time.Time
	mdb.db.QueryRow(`SELECT date_submitted 
	FROM toBeRanked ORDER BY RANDOM()
	LIMIT 1`).Scan(&newSongListTime)

	
	sqlRows, _ := mdb.db.Query(`SELECT song_id, description, curator_id FROM toBeRanked 
		WHERE date_submitted = $1`, newSongListTime.Format(config.StaticEnvs.TimeFormat))
	
	var whatWillBeRankedToday internal_types.TodaysRankingSubmission = internal_types.TodaysRankingSubmission{}
	for sqlRows.Next(){
		var description string
		var curatorId, songId int
		sqlRows.Scan(&songId, &description, &curatorId)
		whatWillBeRankedToday.CuratorId = curatorId
		whatWillBeRankedToday.Description = description
		whatWillBeRankedToday.SongIDs = append(whatWillBeRankedToday.SongIDs, songId)
	}
	mdb.db.Exec(`DELETE FROM toBeRanked 
	WHERE date_submitted = $1`, newSongListTime.Format(config.StaticEnvs.TimeFormat))

	mdb.setTodaysRanking(&whatWillBeRankedToday)
}

func (mdb TodaysRankingDriver) setTodaysRanking(submission *internal_types.TodaysRankingSubmission) {
	if len(submission.SongIDs) > 3 {
		log.Error().Msg("More than three songs will be set for ranking.")
		return
	}

	for i, songID := range submission.SongIDs {
		var name, url, artist string
		err := mdb.db.QueryRow("SELECT name, artist, songURL FROM music WHERE id = $1", songID).Scan(
			&name, &artist, &url,
		)
		if err != nil {
			log.Err(err).Msg("Problem setting todays ranking.")
			return
		}
		resource, err := utils.GetResourceFromYouTubeLink(url)
		if (err != nil){
			log.Err(err).Msg("Resource for todays ranking has a problem.")
		}
		_, err = mdb.db.Exec(`INSERT INTO todaysRanking(
			song_id, curator_id, description, song_name, song_artist,
			song_path_resource, song_order
		) 
		VALUES($1, $2, $3, $4, $5, $6, $7)`,
			songID, submission.CuratorId, submission.Description, name, artist,
			resource, i,
		)

		if err != nil {
			log.Err(err).Msg("Problem setting todays ranking.")
			return
		}
	}
}

func isYesterdayOrBefore(date time.Time) bool {
	yesterday := time.Now().AddDate(0, 0, -1)

	return date.Year() <= yesterday.Year() &&
		date.Month() <= yesterday.Month() &&
		date.Day() <= yesterday.Day()
}
