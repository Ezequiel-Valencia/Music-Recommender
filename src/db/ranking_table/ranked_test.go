package ranking_table

import (
	"music-recommender/types/communication_types"
	"music-recommender/types/internal_types"
	"music-recommender/utils/t_utils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Insert into ranked table correctly
func TestInsertRankedSongs(t *testing.T) {
	adb, dbPointer := t_utils.GetTestDB()
	rankTable := RankedDriver{dbPointer: dbPointer}
	defer t_utils.ResetTestDB()

	t_utils.CreateFakeUser(dbPointer, &t_utils.TestUserBob, "password")
	fillDBWithFakeSongs(dbPointer, adb, &t_utils.TestUserBob)
	ranked_songs := []internal_types.RankedSong{}
	for i := range 3{
		ranked_songs = append(ranked_songs, internal_types.RankedSong{
			SongID: i + 1,
			CuratorId: t_utils.TestUserBob.UserId,
			NumVotes: i,
		})
	}
	var winnerID = 2
	rankTable.InsertAlreadyRankedSongs(winnerID, ranked_songs)

	sqlRows, _ := dbPointer.Query(`SELECT song_id, date_ranked, curator_id, winner FROM ranked`)
	
	i := 1
	for sqlRows.Next(){
		var songID, curatorID int
		var dateRanked time.Time
		var winner bool
		sqlRows.Scan(&songID, &dateRanked, &curatorID, &winner)

		assert.Equal(t, i, songID)
		assert.Equal(t, i == winnerID, winner)
		assert.Equal(t, time.Now().AddDate(0, 0, -1).Day(), dateRanked.Day())
		assert.Equal(t, curatorID, t_utils.TestUserBob.UserId)
		i += 1
	}
}

// 
func TestCalculateTodaysRank(t *testing.T){
	adb, dbPointer := t_utils.GetTestDB()
	defer t_utils.ResetTestDB()

	t_utils.CreateFakeUser(dbPointer, &t_utils.TestUserBob, "password")
	fillDBWithFakeSongs(dbPointer, adb, &t_utils.TestUserBob)

	todaysSubmission := internal_types.TodaysRankingSubmission{Description: "Fake", CuratorId: t_utils.TestUserBob.UserId}
	for i := range 3{
		todaysSubmission.SongIDs = append(todaysSubmission.SongIDs, i + 1)
	}

	rankingDriver := CreateTodaysRankingDriver(dbPointer)
	rankingDriver.setTodaysRanking(&todaysSubmission)
	rankingDriver.UpdateTodaysVoteCount(communication_types.SubmitVotePayload{SongOrder: 0}, t_utils.TestUserBob)
	rankingDriver.UpdateTodaysVoteCount(communication_types.SubmitVotePayload{SongOrder: 2}, t_utils.TestUserBob)
	rankingDriver.UpdateTodaysVoteCount(communication_types.SubmitVotePayload{SongOrder: 2}, t_utils.TestUserBob)

	rankedSongs, topSongID, _ := rankingDriver.CalculateTodaysRank()

	assert.Equal(t, 3, topSongID)
	for i := range 3{
		assert.Equal(t, i, rankedSongs[i].NumVotes)
	}
	
}

