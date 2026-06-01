package auth_table

import (
	"music-recommender/types/internal_types/auth_types"
	"music-recommender/utils/t_utils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var uniqueUsernameAndEmailCases = []struct {
	testCase string
	username string
	email    string
}{
	{
		"Email Exists",
		"hello2",
		"fake@gmail.com",
	},
	{
		"Username Exists",
		"hello",
		"fake2@gmail.com",
	},
	{
		"Both Exists",
		"hello",
		"fake@gmail.com",
	},
}

// The Teardown of the DB Won't Occur Until All Tests are Executed
func TestMain(m *testing.M) {
	m.Run()
	t_utils.TearDownTestDB()
}

func TestDisablingUserCreation(t *testing.T) {
	adb, dbPointer := t_utils.GetTestDB()
	at := CreateAuthTableDriver(dbPointer, adb)

	username := "hello"
	email := "fake@gmail.com"

	assert.True(t, at.IsTheUsernameAndEmailUnique(username, email))

	t_utils.CreateFakeUser(dbPointer, &auth_types.User{Username: username, Email: email}, "password")

	for _, tc := range uniqueUsernameAndEmailCases {
		assert.False(t, at.IsTheUsernameAndEmailUnique(tc.username, tc.email))
	}

	assert.True(t, at.IsTheUsernameAndEmailUnique("hello2", "fake2@gmail.com"))

}

// Test that ensures if any unknown value is pulled from the permissions table, the lowest form of it is returned
func TestLeastPrivilegeForUnknowns(t *testing.T) {
	adb, dbPointer := t_utils.GetTestDB()
	var userID int
	_ = dbPointer.QueryRow(`INSERT INTO users(username, email, password_hash, subject_identifier, creation_source, creation_date)
	VALUES ($1, $2, $3, $4, $5, $6) RETURNING user_id`, "Ezequiel", "fake1@gmail.fake", "f", "f", "f", time.Now()).Scan(&userID)

	_, _ = dbPointer.Exec(`INSERT INTO userPrivileges(user_id, moderator, music_submission) VALUES($1, $2, $3)`, userID, "OwnerOfTheWorld", "AllTheSongs")

	authDriver := CreateAuthTableDriver(dbPointer, adb)
	user := authDriver.GetUserStructFromUsername("Ezequiel")

	// User should still be returned since there is a privilege entry, however, all of their privileges should be the lowest tier.
	assert.Equal(t, "Ezequiel", user.Username)
	assert.Equal(t, auth_types.NoPrivileges, user.UserPrivileges)
	assert.Equal(t, auth_types.VoterRole, user.UserRole)
}
