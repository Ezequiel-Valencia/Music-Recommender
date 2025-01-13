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

type RankingTable struct {
	db *sql.DB
}

func CreateRankingTableDriver(db *sql.DB) *RankingTable {
	return &RankingTable{db: db}
}

func (mdb RankingTable) SetTodaysRanking(submission *internal_types.TodaysRankingSubmission) {
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
		_, err = mdb.db.Exec(`INSERT INTO todaysRanking(
			song_id, curator_name, description, song_name, song_artist,
			song_path_resource, song_order
		) 
		VALUES($1, $2, $3, $4, $5, $6, $7)`,
			songID, submission.CuratorName, submission.Description, name, artist,
			utils.GetResourceFromYouTubeLink(&url), i,
		)

		if err != nil {
			log.Err(err).Msg("Problem setting todays ranking.")
			return
		}
	}
}

func (mdb RankingTable) GetTodaysRanking() communication_types.TodaysRankingPayload {
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

func (mdb RankingTable) UserAlreadyVoteToday(user auth_types.User) bool {
	var lastVote time.Time
	mdb.db.QueryRow("SELECT last_vote FROM users WHERE user_id = $1", user.UserId).Scan(&lastVote)
	return !isYesterdayOrBefore(lastVote)
}

func (mdb RankingTable) UpdateTodaysRanking(submitVote communication_types.SubmitVotePayload, user auth_types.User) {
	_, err := mdb.db.Exec("UPDATE users SET last_vote = $1 WHERE user_id = $2",
		time.Now().Format(config.StaticEnvs.TimeFormat), user.UserId)

	if err != nil {
		log.Err(err).Msg("Can't update users last vote date.")
		return
	}

	_, err = mdb.db.Exec("UPDATE todaysRanking SET num_votes = num_votes + 1 WHERE song_order = $1",
		submitVote.SongNumber)
	if err != nil {
		log.Err(err).Msg("Update ranking did not work")
	}
}

func (mdb RankingTable) GetTodaysMusic() *communication_types.TodaysMusicPayload {

	rows, err := mdb.db.Query(`SELECT song_id, curator_name, description, song_order, song_name, song_artist, song_path_resource
	FROM todaysRanking`)
	if err != nil {
		log.Err(err).Msg("Can't Get Todays Music")
		return &communication_types.TodaysMusicPayload{}
	}

	var musicPayload communication_types.TodaysMusicPayload
	musicPayload.MusicEntries = []communication_types.MusicPayloadEntry{}

	for rows.Next() {
		var songID, order int
		var curatorName, description, songName, songArtist, songResource string
		rows.Scan(&songID, &curatorName, &description, &order, &songName, &songArtist, &songResource)
		musicPayload.CuratorDescription = description
		musicPayload.CuratorName = curatorName

		musicEntry := communication_types.MusicPayloadEntry{Title: songName, Artist: songArtist, SongOrder: order, PathResource: songResource}
		musicPayload.MusicEntries = append(musicPayload.MusicEntries, musicEntry)
	}

	return &musicPayload
}

func (mdb RankingTable) GetCalendarsMusic() *communication_types.CalendarPayload {
	return &communication_types.CalendarPayload{}
}

func isYesterdayOrBefore(date time.Time) bool {
	yesterday := time.Now().AddDate(0, 0, -1)

	return date.Year() <= yesterday.Year() &&
		date.Month() <= yesterday.Month() &&
		date.Day() <= yesterday.Day()
}
