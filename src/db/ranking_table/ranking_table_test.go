package ranking_table

import (
	"database/sql"
	"fmt"
	"music-recommender/db"
	"music-recommender/db/music_table"
	"music-recommender/types/communication_types"
	"music-recommender/types/internal_types"
	"music-recommender/types/internal_types/auth_types"
	"music-recommender/utils/t_utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOnlyOneVote(t *testing.T) {
	adb, dbPointer := t_utils.GetTestDB()
	rankTable := TodaysRankingDriver{db: dbPointer}
	defer t_utils.ResetTestDB()

	t_utils.CreateFakeUser(dbPointer, &t_utils.TestUserBob, "password123")
	assert.False(t, rankTable.UserAlreadyVoteToday(t_utils.TestUserBob))
	fillDBWithFakeSongs(dbPointer, adb, &t_utils.TestUserBob)

	rankTable.UpdateTodaysVoteCount(communication_types.SubmitVotePayload{SongNumber: 1}, t_utils.TestUserBob)

	assert.True(t, rankTable.UserAlreadyVoteToday(t_utils.TestUserBob))
}

func TestTodaysRanking(t *testing.T) {
	adb, dbPointer := t_utils.GetTestDB()
	rankTable := TodaysRankingDriver{db: dbPointer}
	defer t_utils.ResetTestDB()

	t_utils.CreateFakeUser(dbPointer, &t_utils.TestUserBob, "password123")
	fillDBWithFakeSongs(dbPointer, adb, &t_utils.TestUserBob)
	const description string = "I chose these songs for testing."
	songIDs := [3]int{1, 2, 3}
	rankTable.setTodaysRanking(&internal_types.TodaysRankingSubmission{CuratorId: t_utils.TestUserBob.UserId, Description: description,
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

func TestUpdateRanking(t *testing.T) {
	adb, dbPointer := t_utils.GetTestDB()
	rankTable := TodaysRankingDriver{db: dbPointer}
	defer t_utils.ResetTestDB()

	t_utils.CreateFakeUser(dbPointer, &t_utils.TestUserBob, "password123")
	fillDBWithFakeSongs(dbPointer, adb, &t_utils.TestUserBob)
	rankTable.setTodaysRanking(&internal_types.TodaysRankingSubmission{CuratorId: t_utils.TestUserBob.UserId, Description: "yo", SongIDs: []int{1, 2, 3}})

	currentRanking := rankTable.GetTodaysVotes()
	assert.Equal(t, map[int]float64{0: 0, 1: 0, 2: 0}, currentRanking.RankingMap)

	rankTable.UpdateTodaysVoteCount(communication_types.SubmitVotePayload{SongNumber: 1}, t_utils.TestUserBob)
	currentRanking = rankTable.GetTodaysVotes()
	assert.Equal(t, float64(1), currentRanking.RankingMap[1])
	assert.Equal(t, float64(0), currentRanking.RankingMap[0])

	rankTable.UpdateTodaysVoteCount(communication_types.SubmitVotePayload{SongNumber: 0}, t_utils.TestUserBob)
	currentRanking = rankTable.GetTodaysVotes()
	assert.Equal(t, float64(0.5), currentRanking.RankingMap[1])
	assert.Equal(t, float64(0.5), currentRanking.RankingMap[0])
	assert.Equal(t, float64(0.0), currentRanking.RankingMap[2])
}

func fillDBWithFakeSongs(dbPointer *sql.DB, adb *db.AbstractDB, user *auth_types.User) {
	musicDriver := music_table.CreateMusicTableDriver(dbPointer, adb)
	for i := range 10 {
		submitSong := communication_types.SubmitSong{Name: fmt.Sprintf("Song %d", i),
			Artist: fmt.Sprintf("Artist %d", i)}
		musicDriver.InsertNewSong(&submitSong, *user)
	}
}
