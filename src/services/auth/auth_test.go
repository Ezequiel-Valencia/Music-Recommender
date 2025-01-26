package auth

import (	
	"log"
	"music-recommender/config"
	"music-recommender/db/auth_table"
	"music-recommender/types/internal_types/auth_types"
	"music-recommender/utils/t_utils"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Use MockDB For these tests.

// The Teardown of the DB Won't Occur Until All Tests are Executed
func TestMain(m *testing.M) {
	m.Run()
	t_utils.TearDownTestDB()
}

const (
	noSession int = iota
	invalidSessionNotSigned
	invalidSignedName
	invalidCookieName
	validSessionNoUser
	onlySession
	invalidCSRFHeader
	sessionAndCSRF
)

var requireAuthTestCases = []struct {
	testCase     string
	sessionState int
}{
	{
		"No Session",
		noSession,
	},
	{
		"Invalid Session",
		invalidSessionNotSigned,
	},
	{
		"Valid Signature for session_token With Different Signed Name",
		invalidSignedName,
	},
	{
		"Valid Signature But for Cookie With Different Name",
		invalidCookieName,
	},
	{
		"Valid Signature But No User At the End",
		validSessionNoUser,
	},
	{
		"Only Session, no CSRF",
		onlySession,
	},
	{
		"Invalid CSRF Header",
		invalidCSRFHeader,
	},
	{
		"Valid Session and CSRF",
		sessionAndCSRF,
	},
}

func TestRequireAuth(t *testing.T) {
	handler := createAuthHandler()
	defer t_utils.ResetTestDB()

	handler.authTable.CreateUser("Ezequiel", "email", "pw", "", auth_types.LocalUserCreationSource.String())
	user := handler.authTable.GetUserStructFromUsername("Ezequiel")
	unEncodedSession, csrfUnEncode, _ := handler.authTable.GenerateAndStoreSessionID(user, time.Now().UTC().Format(config.StaticEnvs.TimeFormat))
	endPointWithAuth := RequireAuthMinimumPrivileges(handler.deleteUser, handler.authTable.AbstractDB)
	for _, tc := range requireAuthTestCases {
		request := httptest.NewRequest("POST", "/api/v1/delete", nil)
		switch tc.sessionState {
		case noSession:
		case invalidSessionNotSigned:
			request.AddCookie(&http.Cookie{Name: config.StaticEnvs.SessionCookieName, Value: t_utils.GenerateRandomRuneString(20, true)})
		case invalidSignedName:
			cookie, _ := config.SecureCookie.Encode("WrongName", unEncodedSession)
			request.AddCookie(&http.Cookie{Name: config.StaticEnvs.SessionCookieName, Value: cookie})
		case invalidCookieName:
			cookie, _ := config.SecureCookie.Encode(config.StaticEnvs.SessionCookieName, unEncodedSession)
			request.AddCookie(&http.Cookie{Name: "WrongName", Value: cookie})
		case validSessionNoUser:
			cookie, _ := config.SecureCookie.Encode(config.StaticEnvs.SessionCookieName, t_utils.GenerateRandomRuneString(20, true))
			request.AddCookie(&http.Cookie{Name: config.StaticEnvs.SessionCookieName, Value: cookie})
		case onlySession:
			cookie, _ := config.SecureCookie.Encode(config.StaticEnvs.SessionCookieName, unEncodedSession)
			request.AddCookie(&http.Cookie{Name: config.StaticEnvs.SessionCookieName, Value: cookie})
		case invalidCSRFHeader:
			csrfCookie, _ := config.SecureCookie.Encode(config.StaticEnvs.CSRFCookieName, t_utils.GenerateRandomRuneString(20, true))
			cookie, _ := config.SecureCookie.Encode(config.StaticEnvs.SessionCookieName, unEncodedSession)
			request.AddCookie(&http.Cookie{Name: config.StaticEnvs.SessionCookieName, Value: cookie})
			request.Header.Add(config.StaticEnvs.CSRFHeaderName, csrfCookie)
		case sessionAndCSRF:
			csrfCookie, _ := config.SecureCookie.Encode(config.StaticEnvs.CSRFCookieName, csrfUnEncode)
			cookie, _ := config.SecureCookie.Encode(config.StaticEnvs.SessionCookieName, unEncodedSession)
			request.AddCookie(&http.Cookie{Name: config.StaticEnvs.SessionCookieName, Value: cookie})
			request.Header.Add(config.StaticEnvs.CSRFHeaderName, csrfCookie)
		}

		rr := httptest.NewRecorder()
		endPointWithAuth(rr, request)
		log.Print(tc.testCase)
		if tc.sessionState == sessionAndCSRF {
			assert.NotEqual(t, http.StatusTemporaryRedirect, rr.Code)
		} else if (tc.sessionState == invalidCSRFHeader) {
			assert.Equal(t, http.StatusUnauthorized, rr.Code)
		} else {
			assert.Equal(t, http.StatusTemporaryRedirect, rr.Code)
		}
	}
}

const (
	minPrivileges int = iota
	privilegeIsToHigh
	roleIsToHigh
	bothToHigh
	elevatedPrivileges
)

var requireAuthPrivilegesTC = []struct {
	sessionState int
	privilege auth_types.UserPrivileges
	role auth_types.UserRoles
	redirection bool
}{
	{
		minPrivileges,
		auth_types.NoPrivileges,
		auth_types.VoterRole,
		false,
	},
	{
		privilegeIsToHigh,
		auth_types.AdminPrivileges,
		auth_types.VoterRole,
		true,
	},
	{
		roleIsToHigh,
		auth_types.NoPrivileges,
		auth_types.CuratorRole,
		true,
	},
	{
		bothToHigh,
		auth_types.AdminPrivileges,
		auth_types.CuratorRole,
		true,
	},
	{
		elevatedPrivileges,
		auth_types.AdminPrivileges,
		auth_types.CuratorRole,
		false,
	},
}

func TestRequireAuthPrivileges(t *testing.T){
	handler := createAuthHandler()
	defer t_utils.ResetTestDB()

	// Create user
	handler.authTable.CreateUser("Ezequiel", "email", "pw", "", auth_types.LocalUserCreationSource.String())
	user := handler.authTable.GetUserStructFromUsername("Ezequiel")

	// Create cookies
	unEncodedSession, csrfUnEncode, _ := handler.authTable.GenerateAndStoreSessionID(user, time.Now().UTC().Format(config.StaticEnvs.TimeFormat))
	csrfCookie, _ := config.SecureCookie.Encode(config.StaticEnvs.CSRFCookieName, csrfUnEncode)
	cookie, _ := config.SecureCookie.Encode(config.StaticEnvs.SessionCookieName, unEncodedSession)

	// Setup request
	request := httptest.NewRequest("GET", "/api/v1/user", nil)
	request.AddCookie(&http.Cookie{Name: config.StaticEnvs.SessionCookieName, Value: cookie})
	request.Header.Add(config.StaticEnvs.CSRFHeaderName, csrfCookie)

	for _, tc := range requireAuthPrivilegesTC{
		endPointWithAuth := RequireAuth(handler.loggedInUserInfo, handler.authTable.AbstractDB, tc.privilege, tc.role)
		rr := httptest.NewRecorder()

		if (tc.sessionState == elevatedPrivileges){
			handler.authTable.SetUserPrivilege("Ezequiel", auth_types.AdminPrivileges)
			handler.authTable.SetUserRole("Ezequiel", auth_types.CuratorRole)
		}

		endPointWithAuth(rr, request)

		if (tc.redirection){
			assert.Equal(t, http.StatusTemporaryRedirect, rr.Code)
		} else{
			assert.NotEqual(t, http.StatusTemporaryRedirect, rr.Code)
		}
	}
}


func createAuthHandler() *Handler {
	adb, dbPointer := t_utils.GetTestDB()
	at := auth_table.CreateAuthTableDriver(dbPointer, adb)
	return NewHandler(at)
}
