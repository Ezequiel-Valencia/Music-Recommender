package ranking_table

import (
	"fmt"
	"music-recommender/types/communication_types"
	"music-recommender/types/internal_types"
	"music-recommender/utils/t_utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Users get only a single vote
func TestOnlyOneVote(t *testing.T) {
	adb, dbPointer := t_utils.GetTestDB()
	rankTable := TodaysRankingDriver{db: dbPointer}
	defer t_utils.ResetTestDB()

	t_utils.CreateFakeUser(dbPointer, &t_utils.TestUserBob, "password123")
	assert.False(t, rankTable.UserAlreadyVoteToday(t_utils.TestUserBob))
	t_utils.FillDBWithFakeSongsAndDescription(dbPointer, adb, &t_utils.TestUserBob, "yo")

	rankTable.UpdateTodaysVoteCount(communication_types.SubmitVotePayload{SongOrder: 1}, t_utils.TestUserBob)

	assert.True(t, rankTable.UserAlreadyVoteToday(t_utils.TestUserBob))
}

// Correctly set and get todays ranking
func TestTodaysRanking(t *testing.T) {
	adb, dbPointer := t_utils.GetTestDB()
	rankTable := TodaysRankingDriver{db: dbPointer}
	defer t_utils.ResetTestDB()

	t_utils.CreateFakeUser(dbPointer, &t_utils.TestUserBob, "password123")
	const description string = "I chose these songs for testing."
	t_utils.FillDBWithFakeSongsAndDescription(dbPointer, adb, &t_utils.TestUserBob, description)
	songIDs := [3]int{1, 2, 3}
	rankTable.setTodaysRanking(&internal_types.TodaysRankingSubmission{CuratorId: t_utils.TestUserBob.UserId, Description_Id: 1,
		SongIDs: songIDs[:]})

	musicPayload := rankTable.GetTodaysMusic()

	assert.Equal(t, description, musicPayload.CuratorDescription)
	assert.Equal(t, musicPayload.CuratorName, t_utils.TestUserBob.Username)
	for i, me := range musicPayload.MusicEntries {
		assert.Equal(t, fmt.Sprintf("Song %d", i), me.Title)
		assert.Equal(t, fmt.Sprintf("Artist %d", i), me.Artist)
		assert.Equal(t, i, me.SongOrder)
	}
}

// Correctly update the vote counts, and retrieve their percentage
func TestUpdateRanking(t *testing.T) {
	adb, dbPointer := t_utils.GetTestDB()
	rankTable := TodaysRankingDriver{db: dbPointer}
	defer t_utils.ResetTestDB()

	t_utils.CreateFakeUser(dbPointer, &t_utils.TestUserBob, "password123")
	t_utils.FillDBWithFakeSongsAndDescription(dbPointer, adb, &t_utils.TestUserBob, "yo")
	rankTable.setTodaysRanking(&internal_types.TodaysRankingSubmission{CuratorId: t_utils.TestUserBob.UserId, Description_Id: 1, SongIDs: []int{1, 2, 3}})

	currentRanking := rankTable.GetTodaysVotes()
	assert.Equal(t, map[int]float64{0: 0, 1: 0, 2: 0}, currentRanking.RankingMap)

	rankTable.UpdateTodaysVoteCount(communication_types.SubmitVotePayload{SongOrder: 1}, t_utils.TestUserBob)
	currentRanking = rankTable.GetTodaysVotes()
	assert.Equal(t, float64(1), currentRanking.RankingMap[1])
	assert.Equal(t, float64(0), currentRanking.RankingMap[0])

	rankTable.UpdateTodaysVoteCount(communication_types.SubmitVotePayload{SongOrder: 0}, t_utils.TestUserBob)
	currentRanking = rankTable.GetTodaysVotes()
	assert.Equal(t, float64(0.5), currentRanking.RankingMap[1])
	assert.Equal(t, float64(0.5), currentRanking.RankingMap[0])
	assert.Equal(t, float64(0.0), currentRanking.RankingMap[2])
}


