package music_table_test

import (
	"database/sql"
	"music-recommender/db/music_table"
	"music-recommender/types/communication_types"
	"music-recommender/utils/t_utils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Checks voter role does nothing, curator can submit, limits are enforced, and everything submitted is correct
// Also ensuring the SQL function behaves appropriately
func TestInsertSet(t *testing.T) {
	adb, dbPointer := t_utils.GetTestDB()
	defer t_utils.ResetTestDB()
	///////////
	// Setup //
	///////////
	t_utils.CreateFakeUser(dbPointer, &t_utils.TestUserBob, "password")
	t_utils.CreateFakeUser(dbPointer, &t_utils.TestUserCuratorModerator, "password")
	driver := music_table.CreateMusicTableDriver(dbPointer, adb)

	fakeSong := communication_types.SubmitSong{
		Name:     "Fake",
		Artist:   "Artist",
		SongURL:  "Song url",
		Genre:    "Genre",
		Subgenre: "Sub",
	}
	fakeSongSet := []communication_types.SubmitSong{fakeSong, fakeSong, fakeSong}

	/////////////
	//! Tests !//
	/////////////

	// Voter role does no submission
	err := driver.InsertSongSet(&communication_types.SubmitSongSet{Description: "Fake Description",
		Songs: fakeSongSet}, t_utils.TestUserBob)

	expectedErrorAndRowCount(t, dbPointer, err, true, 0)

	/////////////////
	// Correctness //
	////////////////
	// Curator can submit, their submit limit is updated, and songs are queued
	err = driver.InsertSongSet(&communication_types.SubmitSongSet{Description: "Fake Description",
		Songs: fakeSongSet}, t_utils.TestUserCuratorModerator)
	expectedErrorAndRowCount(t, dbPointer, err, false, 3)

	var description string
	var curatorID int
	var timeSubmitted time.Time
	_ = dbPointer.QueryRow(`SELECT description, curator_id, date_submitted FROM toBeRanked WHERE song_id = 1`).Scan(&description, &curatorID, &timeSubmitted)
	assert.Equal(t, "Fake Description", description)
	assert.Equal(t, t_utils.TestUserCuratorModerator.UserId, curatorID)
	sqlRows, _ := dbPointer.Query(`SELECT date_submitted FROM toBeRanked`)

	// Make sure all of the time stamps for the set of 3 songs are the same
	for sqlRows.Next() {
		var timeGiven time.Time
		_ = sqlRows.Scan(&timeGiven)
		assert.True(t, timeSubmitted.Equal(timeGiven))
	}

	var submitCount int
	_ = dbPointer.QueryRow(`SELECT song_sets_submitted FROM users WHERE user_id = $1`, t_utils.TestUserCuratorModerator.UserId).Scan(&submitCount)
	assert.Equal(t, 1, submitCount)

	/////////////////
	// Limit Test //
	////////////////
	// Make sure that curator can not submit more than the limit for their role
	_, _ = dbPointer.Exec(`UPDATE users SET song_sets_submitted = $1 WHERE user_id = $2`,
		t_utils.TestUserCuratorModerator.UserRole.GetRolesSubmissionLimit(),
		t_utils.TestUserCuratorModerator.UserId)

	err = driver.InsertSongSet(&communication_types.SubmitSongSet{Description: "Fake Description",
		Songs: fakeSongSet}, t_utils.TestUserCuratorModerator)
	expectedErrorAndRowCount(t, dbPointer, err, true, 3)

	_, _ = dbPointer.Exec(`UPDATE users SET song_sets_submitted = $1 WHERE user_id = $2`,
		t_utils.TestUserCuratorModerator.UserRole.GetRolesSubmissionLimit()+1,
		t_utils.TestUserCuratorModerator.UserId)

	err = driver.InsertSongSet(&communication_types.SubmitSongSet{Description: "Fake Description",
		Songs: fakeSongSet}, t_utils.TestUserCuratorModerator)
	expectedErrorAndRowCount(t, dbPointer, err, true, 3)
}

func expectedErrorAndRowCount(t *testing.T, dbPointer *sql.DB, err error, expectError bool, expectedRowCount int) {
	if expectError {
		assert.Error(t, err)
		assert.ErrorContains(t, err, `user has reached song submission limit`)
	} else {
		assert.NoError(t, err)
	}
	sqlResult, _ := dbPointer.Exec(`SELECT * FROM toBeRanked`)
	numRows, _ := sqlResult.RowsAffected()
	assert.Equal(t, int64(expectedRowCount), numRows)
}
