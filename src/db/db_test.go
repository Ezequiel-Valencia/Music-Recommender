package db_test

import (
	"music-recommender/db/auth_table"
	"music-recommender/types/internal_types/auth_types"
	"music-recommender/utils/t_utils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)


func TestMain(m *testing.M) {
	m.Run()
}

func TestUserPrivilegeEntryRequirement(t *testing.T) {
	adb, dbPointer := t_utils.GetTestDB()
	defer t_utils.TearDownTestDB()

	// Users with no privileges reference return empty user, and logs error
	dbPointer.Exec(`INSERT INTO users(username, email, password_hash, subject_identifier, creation_source, creation_date) 
	VALUES ($1, $2, $3, $4, $5, $6)`, "Ezequiel", "fake1@gmail.fake", "f", "f", "f", time.Now())
	
	at := auth_table.CreateAuthTableDriver(dbPointer, adb)
	ezUser := at.GetUserStructFromUsername("Ezequiel")
	assert.Equal(t, auth_types.NoPrivileges, ezUser.UserPrivileges)
	assert.Equal(t, auth_types.VoterRole, ezUser.UserRole)
	assert.Equal(t, ezUser, auth_types.User{})


	t_utils.CreateFakeUser(dbPointer, &t_utils.TestUserCuratorModerator, "password123")
	curUser := at.GetUserStructFromUsername(t_utils.TestUserCuratorModerator.Username)
	assert.Equal(t, auth_types.ModeratorPrivileges, curUser.UserPrivileges)
	assert.Equal(t, auth_types.CuratorRole, curUser.UserRole)
}


