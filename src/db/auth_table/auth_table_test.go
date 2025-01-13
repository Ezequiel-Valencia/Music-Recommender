package auth_table

import (
	"music-recommender/types/internal_types/auth_types"
	"music-recommender/utils/t_utils"
	"testing"

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
