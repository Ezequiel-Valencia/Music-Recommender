package ranking_table

import (
	"database/sql"
	"music-recommender/config"
	"music-recommender/types/communication_types"
	"music-recommender/types/internal_types"
	"time"

	"github.com/rs/zerolog/log"
)

type RankedDriver struct {
	dbPointer *sql.DB
}

func CreateRankedDriver(db *sql.DB) *RankedDriver{
	return &RankedDriver{dbPointer: db}
}

// Assumes the top song ID is accurate, and is given directly from Calculate Todays Rank
func (rd RankedDriver) InsertAlreadyRankedSongs(topSongId int, rankedSongs []internal_types.RankedSong){
	voteDate := time.Now().AddDate(0, 0, -1)
	for _, rs := range rankedSongs{
		rd.dbPointer.Exec(`INSERT INTO ranked(
			song_id, curator_id, date_ranked, num_votes, winner
		) VALUES($1, $2, $3, $4, $5)`, rs.SongID, rs.CuratorId, 
		voteDate.Format(config.StaticEnvs.TimeFormat), rs.NumVotes, rs.SongID == topSongId)
	}
}

// All the songs ranked, top songs id, error
func (rd TodaysRankingDriver) CalculateTodaysRank()([]internal_types.RankedSong, int, error){
	res, err := rd.db.Query("SELECT song_id, num_votes, curator_id FROM todaysRanking")
	if err != nil {
		log.Err(err).Msg("Can't compute todays ranking.")
		return nil, -1, err
	}
	var rankedSongs []internal_types.RankedSong = []internal_types.RankedSong{}
	var topSongId int = 1
	var topVotes int = -1
	for res.Next() {
		var song_id, numVotes, curator_id int

		res.Scan(&song_id, &numVotes, &curator_id)
		if (numVotes > topVotes){
			topVotes = numVotes
			topSongId = song_id
		}
		rankedSongs = append(rankedSongs, 
			internal_types.RankedSong{SongID: song_id, NumVotes: numVotes, CuratorId: curator_id},
		)
	}
	return rankedSongs, topSongId, nil
}

func (rd TodaysRankingDriver) CleanTodaysRanking(){
	_, err := rd.db.Exec("DELETE FROM todaysRanking")
	if (err != nil){
		log.Err(err).Msg("Didn't clean todays ranking.")
	}
}

func (mdb TodaysRankingDriver) GetCalendarsMusic() *communication_types.CalendarPayload {
	return &communication_types.CalendarPayload{}
}


