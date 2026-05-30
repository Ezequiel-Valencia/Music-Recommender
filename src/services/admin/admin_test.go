package admin

import (
	"music-recommender/config"
	"music-recommender/db/auth_table"
	"music-recommender/db/ranking_table"
	"music-recommender/types/internal_types/auth_types"
	"music-recommender/utils/t_utils"
	"net/http"
	"net/http/httptest"
	"testing"

	"gotest.tools/assert"
)

const (
	notPrivilegedUser int = iota
	formBodyNotPresent
	queryParamNotBoolean
	validRequestToFalse
	validRequestToTrue
)

var disableCreationCheck = []struct {
	allowanceState int
	code           int
	allowCreation  bool
}{
	{
		notPrivilegedUser,
		http.StatusUnauthorized,
		true,
	},
	{
		formBodyNotPresent,
		http.StatusOK,
		false,
	},
	{
		validRequestToTrue,
		http.StatusOK,
		true,
	},
}

/*
Disabling is determined whether or not the form entity 'allow-user-creation' is present or not.
*/
func TestDisablingUserCreation(t *testing.T) {
	handler := createAuthHandler()
	_, dbPointer := t_utils.GetTestDB()
	defer t_utils.ResetTestDB()

	owner := auth_types.User{Username: "Ezequiel", UserRole: auth_types.UnlimitedRole,
		UserPrivileges: auth_types.OwnerPrivileges, UserId: 1}
	badActor := auth_types.User{Username: "Couch", UserRole: auth_types.CuratorRole,
		UserPrivileges: auth_types.AdminPrivileges, UserId: 2}
	t_utils.CreateFakeUser(dbPointer, &owner, "pass")
	t_utils.CreateFakeUser(dbPointer, &badActor, "pass")

	for _, tc := range disableCreationCheck {
		rr := httptest.NewRecorder()
		var request *http.Request
		var testUser auth_types.User
		switch tc.allowanceState {
		case notPrivilegedUser:
			testUser = badActor
			request = httptest.NewRequest("POST", "/allowCreation?allow-user-creation", nil)
		case formBodyNotPresent:
			testUser = owner
			request = httptest.NewRequest("POST", "/allowCreation", nil)
		case validRequestToTrue:
			testUser = owner
			request = httptest.NewRequest("POST", "/allowCreation?allow-user-creation=true", nil)
		}

		handler.setUserCreationAllowance(rr, request, testUser)
		assert.Equal(t, tc.code, rr.Code)
		assert.Equal(t, tc.allowCreation, config.DynamicEnvs.AllowUserCreation)
	}
}

func createAuthHandler() *Handler {
	adb, dbPointer := t_utils.GetTestDB()
	at := auth_table.CreateAuthTableDriver(dbPointer, adb)
	trt := ranking_table.CreateTodaysRankingDriver(dbPointer)
	return NewHandler(at, trt)
}
