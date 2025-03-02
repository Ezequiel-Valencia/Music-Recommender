package auth_types_test

import (
	"music-recommender/types/internal_types/auth_types"
	"testing"
	"github.com/stretchr/testify/assert"
)

// The Teardown of the DB Won't Occur Until All Tests are Executed
func TestMain(m *testing.M) {
	m.Run()
}


// Test that ensures if any unknown value is pulled from the permissions table, the lowest form of it is returned
func TestLeastPrivilegeForUnknowns(t *testing.T){
	assert.Equal(t, auth_types.NoPrivileges, auth_types.StringToUserPrivileges("OwnerOfTheWorld"))
	assert.Equal(t, auth_types.VoterRole, auth_types.StringToUserRoles("AllSongsCanSubmit"))
}

