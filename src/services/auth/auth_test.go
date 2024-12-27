package auth

import (
	"bytes"
	"fmt"
	"io"
	"music-recommender/config"
	"music-recommender/db"
	"music-recommender/db/auth_table"
	"music-recommender/utils/t_utils"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Use MockDB For these tests.

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
		httptest.NewRequest("POST", "/api/v1/login", bytes.NewBufferString("username=Ezequiel&password=password1233")),
	},
}

// The Teardown of the DB Won't Occur Until All Tests are Executed
func TestMain(m *testing.M) {
	m.Run()
	t_utils.TearDownTestDB()
}

// By Having Valid User in this Test, ensures that all failures are because of their respective reasons and not because there isn't a valid user to begin with
func TestLogin(t *testing.T) {
	handler := createAuthHandler()
	defer t_utils.ResetTestDB()

	hp, _ := hashPassword("password123")
	handler.authTable.CreateUser("Ezequiel", hp, "", db.LocalUserCreationSource.String())

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
	assert.Equal(t, "User already logged in.\n", string(bod))

	// Valid Credentials Test Case
	request = httptest.NewRequest("POST", "/api/v1/login", bytes.NewBufferString("username=Ezequiel&password=password123"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	handler.login(rr, request)
	assert.Equal(t, http.StatusOK, rr.Code)
}

const (
	noCookie	int = iota
	inValidCookie
	validCookie
)

var logOutTestCases = []struct{
	testName	string
	cookieState		int
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


func TestLogOut(t *testing.T){
	handler := createAuthHandler()
	defer t_utils.ResetTestDB()

	req := httptest.NewRequest(http.MethodPost, "/logout", bytes.NewReader([]byte("username=Ezequiel&password=password123")))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	handler.register(rr, req)
	sessionCookie := rr.Result().Cookies()[0]

	for _, tc := range logOutTestCases{
		var testRecorder = httptest.NewRecorder()
		var request = httptest.NewRequest("POST", "/logout", bytes.NewBufferString(""))
		switch tc.cookieState{
		case inValidCookie:
			request.AddCookie(&http.Cookie{Name: config.StaticEnvs.SessionCookieName, Value: t_utils.GenerateRandomRuneString(14, true)})
		case validCookie:
			request.AddCookie(&http.Cookie{Name: config.StaticEnvs.SessionCookieName, Value: sessionCookie.Value})
		case noCookie:
		}

		handler.logout(testRecorder, request, db.User{UserId: 1})
		assert.Equal(t, tc.statusCode, testRecorder.Code)
		bod, _ := io.ReadAll(testRecorder.Body)
		assert.Equal(t, tc.bodResponse, string(bod))
	}
}


var registerTestCases = []struct {
	testCase string
	request  *http.Request
	code	 int
	expectedResponse	string
}{
	{
		"Invalid Username Chars",
		httptest.NewRequest("POST", "/api/v1/register", bytes.NewBufferString(fmt.Sprintf("username=Ez@%s&password=password123", t_utils.GenerateRandomRuneString(5, false)))),
		http.StatusNotAcceptable,
		"Invalid username/password\n",
	},
	{
		"Invalid Password Chars",
		httptest.NewRequest("POST", "/api/v1/register", bytes.NewBufferString(fmt.Sprintf("username=Ezequiel&password=password123@%s", t_utils.GenerateRandomRuneString(5, false)))),
		http.StatusNotAcceptable,
		"Invalid username/password\n",
	},
	{
		"User Already Exists",
		httptest.NewRequest("POST", "/api/v1/register", bytes.NewBufferString("username=Ezequiel&password=password123")),
		http.StatusConflict,
		"User already exists\n",
	},
	{
		"Create User, All Good",
		httptest.NewRequest("POST", "/api/v1/register", bytes.NewBufferString("username=Ezequiel2&password=password123")),
		http.StatusOK,
		"",
	},
}


func TestRegister(t *testing.T){
	handler := createAuthHandler()
	defer t_utils.ResetTestDB()


	hp, _ := hashPassword("password123")
	handler.authTable.CreateUser("Ezequiel", hp, "", db.LocalUserCreationSource.String())

	for _, tc := range registerTestCases {
		tc.request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler.register(rr, tc.request)

		assert.Equal(t, tc.code, rr.Code)
		if rr.Code == http.StatusOK{
			session := rr.Result().Header.Get("Set-Cookie")
			assert.NotEqual(t, "", session)

			user := handler.authTable.GetUserStructFromUsername("Ezequiel2")
			assert.Equal(t, 2, user.UserId) // Second user created
		}

		bod, _ := io.ReadAll(rr.Body)
		assert.Equal(t, tc.expectedResponse, string(bod))
	}

}

func TestDeleteUser(t *testing.T){
	handler := createAuthHandler()
	
	request := httptest.NewRequest("POST", "/api/v1/register", bytes.NewBufferString("username=Ezequiel&password=password123"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler.register(rr, request)
	user := handler.authTable.GetUserStructFromUsername("Ezequiel")
	assert.Equal(t, 1, user.UserId)
	assert.Equal(t, "Ezequiel", user.Username)

	handler.deleteUser(rr, request, user)
	user = handler.authTable.GetUserStructFromUsername("Ezequiel")
	assert.Equal(t, 0, user.UserId)
}


const (
	noSession int	= iota
	invalidSession
	validSession
)

var requireAuthTestCases = []struct {
	testCase string
	sessionState int
}{
	{
		"No Session",
		noSession,
	},
	{
		"Invalid Session",
		invalidSession,
	},
	{
		"Valid Session",
		validSession,
	},
}

func TestRequireAuth(t *testing.T){
	handler := createAuthHandler()

	handler.authTable.CreateUser("Ezequiel", "pw", "", db.LocalUserCreationSource.String())
	user := handler.authTable.GetUserStructFromUsername("Ezequiel")
	session, _ := handler.authTable.GenerateAndStoreSessionID(user)
	endPointWithAuth := RequireAuth(handler.deleteUser, handler.authTable.AbstractDB)
	for _, tc := range requireAuthTestCases{
		request := httptest.NewRequest("POST", "/api/v1/delete", nil)
		switch tc.sessionState{
		case noSession:
		case invalidSession:
			request.AddCookie(&http.Cookie{Name: config.StaticEnvs.SessionCookieName, Value: t_utils.GenerateRandomRuneString(20, true)})
		case validSession:
			request.AddCookie(&http.Cookie{Name: config.StaticEnvs.SessionCookieName, Value: session})
		}

		rr := httptest.NewRecorder()
		endPointWithAuth(rr, request)
		if (tc.sessionState != validCookie){
			assert.Equal(t, http.StatusTemporaryRedirect, rr.Code)
		} else{
			assert.NotEqual(t, http.StatusTemporaryRedirect, rr.Code)
		}
	}
}

func createAuthHandler() *Handler{
	adb, dbPointer := t_utils.GetTestDB()
	at := auth_table.CreateAuthTableDriver(dbPointer, adb)
	return NewHandler(at)
}

