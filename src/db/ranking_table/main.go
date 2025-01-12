package ranking_table

import (
	"database/sql"
	"music-recommender/config"
	"music-recommender/db"
	"music-recommender/types/communication_types"
	"time"

	"github.com/rs/zerolog/log"
)


type RankingTable struct{
	db *sql.DB
}


func CreateRankingTableDriver(db *sql.DB) *RankingTable{
	return &RankingTable{db: db}
}


func (mdb RankingTable) GetTodaysRanking() *communication_types.TodaysRankingPayload {
	res, err := mdb.db.Query("SELECT song_order, num_votes FROM todaysRanking")
	if err != nil {
		log.Err(err).Msg("Can't get todays ranking.")
	}
	todaysRanking := communication_types.TodaysRankingPayload{}
	var totalVotes int = 0
	for res.Next(){
		var order, numVotes int
		res.Scan(&order, &numVotes)
		totalVotes += numVotes
		todaysRanking.RankingMap[order] += numVotes
	}
	for key := range todaysRanking.RankingMap {		
		todaysRanking.RankingMap[key] /= totalVotes
	}
	return &todaysRanking
}

func (mdb RankingTable) UserAlreadyVoteToday(user db.User) bool{
	var lastVote time.Time
	mdb.db.QueryRow("SELECT last_vote FROM users WHERE user_id = $1", user.UserId).Scan(&lastVote)
	return !isYesterdayOrBefore(lastVote)
}

func (mdb RankingTable) UpdateTodaysRanking(submitVote communication_types.SubmitVotePayload, user db.User) {
	_, err := mdb.db.Exec("UPDATE users SET last_vote = $1 WHERE user_id = $2", 
	time.Now().Format(config.StaticEnvs.TimeFormat), user.UserId)

	if err != nil{
		log.Err(err).Msg("Can't update users last vote date.")
		return
	}

	_, err = mdb.db.Exec("UPDATE todaysRanking SET num_votes = num_votes + 1 WHERE song_order = $1",
		submitVote.SongNumber)
	if err != nil{
		log.Err(err).Msg("Update ranking did not work")
	}
}

func (mdb RankingTable) GetTodaysMusic() *communication_types.TodaysMusicPayload {

	rows, err := mdb.db.Query(`SELECT songID, curator_name, description, song_order, song_name, song_artist, song_path_resource
	FROM todaysRanking`)
	if (err != nil){
		log.Err(err).Msg("Can't Get Todays Rankings")
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


