package auth

import (
	"bytes"
	"fmt"
	"io"
	"music-recommender/config"
	"music-recommender/types/internal_types/auth_types"
	"music-recommender/utils/t_utils"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)


var invalidUsernameOrPasswordTestCases = []struct {
	testCase string
	request  *http.Request
}{
	{
		"Invalid Username Chars",
		httptest.NewRequest("POST", "/api/v1/login", bytes.NewBufferString(fmt.Sprintf("username=Ez@%s&password=password123", t_utils.GenerateRandomRuneString(5, false)))),
	},
	{
		"Invalid Password Chars",
		httptest.NewRequest("POST", "/api/v1/login", bytes.NewBufferString(fmt.Sprintf("username=Ezequiel&password=password123@%s", t_utils.GenerateRandomRuneString(5, false)))),
	},
	{
		"Username To Short",
		httptest.NewRequest("POST", "/api/v1/login", bytes.NewBufferString("username=Ez&password=password123@")),
	},
	{
		"Password To Short",
		httptest.NewRequest("POST", "/api/v1/login", bytes.NewBufferString("username=Ezequiel&password=pas")),
	},
	{
		"Username To Long",
		httptest.NewRequest("POST", "/api/v1/login", bytes.NewBufferString(fmt.Sprintf("username=Ezequiel%s&password=password123", t_utils.GenerateRandomRuneString(25, true)))),
	},
	{
		"Password To Long",
		httptest.NewRequest("POST", "/api/v1/login", bytes.NewBufferString(fmt.Sprintf("username=Ezequiel&password=password123%s", t_utils.GenerateRandomRuneString(60, true)))),
	},
	{
		"Incorrect Username",
		httptest.NewRequest("POST", "/api/v1/login", bytes.NewBufferString("username=Ezequiel9&password=password123")),
	},
	{
		"Incorrect Password",
		httptest.NewRequest("POST", "/api/v1/login", bytes.NewBufferString("username=Ezequiel&password=password1233&email=fake@gmail.com")),
	},
}

// By Having Valid User in this Test, ensures that all failures are because of their respective reasons and not because there isn't a valid user to begin with
func TestLogin(t *testing.T) {
	handler := createAuthHandler()
	defer t_utils.ResetTestDB()

	testUser := auth_types.User{Username: "Ezequiel", Email: "fake@gmail.com", CreationSource: auth_types.LocalUserCreationSource}
	_, db := t_utils.GetTestDB()
	t_utils.CreateFakeUser(db, &testUser, "password123")

	for _, tc := range invalidUsernameOrPasswordTestCases {
		tc.request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler.login(rr, tc.request)

		assert.Equal(t, http.StatusNotAcceptable, rr.Code)
		bod, _ := io.ReadAll(rr.Body)
		assert.Equal(t, "Invalid username/password\n", string(bod))

	}

	// Doesn't matter if session is valid. If the cookie is there don't allow login.
	// Already Logged In Failure Test Case
	request := httptest.NewRequest("POST", "/api/v1/login", bytes.NewBufferString("username=Ezequiel&password=password123"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.AddCookie(&http.Cookie{Name: config.StaticEnvs.SessionCookieName, Value: "anything"})
	rr := httptest.NewRecorder()
	handler.login(rr, request)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	bod, _ := io.ReadAll(rr.Body)
	assert.Equal(t, "User already logged in. Please clear cookies for a new valid session.\n", string(bod))

	// Valid Credentials Test Case
	request = httptest.NewRequest("POST", "/api/v1/login", t_utils.CreateHTTPBodyURLEncoded("username=Ezequiel&password=password123&email=fake@gmail.com"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	handler.login(rr, request)
	assert.Equal(t, http.StatusOK, rr.Code)
}


func TestMaxNumberOfSessions(t *testing.T) {
	handler := createAuthHandler()
	defer t_utils.ResetTestDB()

	hashedPW, _ := hashPassword("password123")
	handler.authTable.CreateUser("Ezequiel", "fake@gmail.com", hashedPW, "", auth_types.LocalUserCreationSource.String())

	request := httptest.NewRequest("POST", "/api/v1/register", t_utils.CreateHTTPBodyURLEncoded("username=Ezequiel&password=password123&email=fake@gmail.com"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for i := range 15 {
		rr := httptest.NewRecorder()
		handler.login(rr, request)
		if i <= 9 {
			assert.Equal(t, http.StatusOK, rr.Code)
		} else {
			assert.Equal(t, http.StatusTooManyRequests, rr.Code)
		}
	}
}


const (
	noCookie int = iota
	inValidCookie
	validCookie
)

var logOutTestCases = []struct {
	testName    string
	cookieState int
	statusCode  int
	bodResponse string
}{
	{
		testName: "No Cookie, No Auth", bodResponse: "Can't logout\n", cookieState: noCookie,
		statusCode: http.StatusUnauthorized,
	},
	{
		testName: "Invalid Session Cookie", bodResponse: "Can't logout\n", cookieState: inValidCookie,
		statusCode: http.StatusBadRequest,
	},
	{testName: "Valid Cookie", cookieState: validCookie, statusCode: http.StatusOK},
}

func TestLogOut(t *testing.T) {
	handler := createAuthHandler()
	defer t_utils.ResetTestDB()

	req := httptest.NewRequest(http.MethodPost, "/logout", bytes.NewReader([]byte("username=Ezequiel&password=password123&email=fake@gmail.com")))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	handler.register(rr, req)
	sessionCookie := rr.Result().Cookies()[0]
	var decodedSession string
	config.SecureCookie.Decode(config.StaticEnvs.SessionCookieName, sessionCookie.Value, &decodedSession)

	for _, tc := range logOutTestCases {
		var testRecorder = httptest.NewRecorder()
		var request = httptest.NewRequest("POST", "/logout", bytes.NewBufferString(""))
		switch tc.cookieState {
		case inValidCookie:
			request.AddCookie(&http.Cookie{Name: config.StaticEnvs.SessionCookieName, Value: t_utils.GenerateRandomRuneString(14, true)})
		case validCookie:
			request.AddCookie(&http.Cookie{Name: config.StaticEnvs.SessionCookieName, Value: sessionCookie.Value})
		case noCookie:
		}

		handler.logout(testRecorder, request, auth_types.User{UserId: 1})
		user := handler.authTable.GetUserStructFromUsername("Ezequiel")

		sessionUser, _ := handler.authTable.AbstractDB.GetUserFromSessionID(decodedSession, "", false)

		// Assure you don't actually delete the user
		assert.Equal(t, "Ezequiel", user.Username)
		assert.Equal(t, 1, user.UserId)

		// Check if session is gone
		if tc.cookieState == validCookie {
			assert.Equal(t, "", sessionUser.Username)
			assert.Equal(t, 0, sessionUser.UserId)
		} else {
			assert.Equal(t, "Ezequiel", sessionUser.Username)
			assert.Equal(t, 1, sessionUser.UserId)
		}

		assert.Equal(t, tc.statusCode, testRecorder.Code)
		bod, _ := io.ReadAll(testRecorder.Body)
		assert.Equal(t, tc.bodResponse, string(bod))
	}
}

