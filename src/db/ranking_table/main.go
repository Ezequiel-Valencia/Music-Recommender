package ranking_table

import (
	"database/sql"
	"music-recommender/types"

	"github.com/rs/zerolog/log"
)


type RankingTable struct{
	db *sql.DB
}


func CreateRankingTableDriver(db *sql.DB) *RankingTable{
	return &RankingTable{db: db}
}


func (mdb RankingTable) GetTodaysRanking() *types.TodaysRankingPayload {
	res := mdb.db.QueryRow("SELECT * FROM todaysRanking WHERE insert_date = ?")
	res.Scan()
	return &types.TodaysRankingPayload{}
}

func (mdb RankingTable) UpdateTodaysRanking(submitVote types.SubmitVotePayload) {
	_, err := mdb.db.Exec("UPDATE todaysRanking SET num_votes = num_votes + 1 WHERE name = ?",
		submitVote.SongName)
	if err != nil{
		log.Err(err).Msg("Update ranking did not work")
	}
}

func (mdb RankingTable) GetTodaysMusic() *types.TodaysMusicPayload {
	return &types.TodaysMusicPayload{}
}

func (mdb RankingTable) GetCalendarsMusic() *types.CalendarPayload {
	return &types.CalendarPayload{}
}



