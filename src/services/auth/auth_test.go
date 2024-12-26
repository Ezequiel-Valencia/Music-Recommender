package auth

import (
	"bytes"
	"fmt"
	"io"
	"log"
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
	adb, dbPointer := t_utils.GetTestDB()
	at := auth_table.CreateAuthTableDriver(dbPointer, adb)
	handler := NewHandler(at)
	hp, _ := hashPassword("password123")
	handler.authTable.CreateUser("Ezequiel", hp, "", db.LocalUserCreationSource.String())

	for _, tc := range invalidUsernameOrPasswordTestCases {
		log.Print(tc.testCase + " Test Case")
		tc.request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler.login(rr, tc.request)

		assert.Equal(t, http.StatusNotAcceptable, rr.Code)
		bod, _ := io.ReadAll(rr.Body)
		assert.Equal(t, "Invalid username/password\n", string(bod))

	}

	// Doesn't matter if session is valid. If the cookie is there don't allow login.
	log.Print("Already Logged In Failure Test Case")
	request := httptest.NewRequest("POST", "/api/v1/login", bytes.NewBufferString("username=Ezequiel&password=password123"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.AddCookie(&http.Cookie{Name: config.StaticEnvs.SessionCookieName, Value: "anything"})
	rr := httptest.NewRecorder()
	handler.login(rr, request)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	bod, _ := io.ReadAll(rr.Body)
	assert.Equal(t, "User already logged in.\n", string(bod))

	log.Print("Valid Credentials Test Case")
	request = httptest.NewRequest("POST", "/api/v1/login", bytes.NewBufferString("username=Ezequiel&password=password123"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	handler.login(rr, request)
	assert.Equal(t, http.StatusOK, rr.Code)

	t_utils.ResetTestDB()
}
