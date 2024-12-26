package auth

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"music-recommender/db/auth_table"
	"music-recommender/utils/t_utils"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Use MockDB For these tests.

var invalidUsernameOrPasswordTestCases = []struct{
	testCase		string
	request			*http.Request
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
}

func TestMain(m *testing.M){
	m.Run()
	t_utils.TearDownTestDB()
}

func TestInvalidLoginChars(t *testing.T){
	adb, dbPointer := t_utils.GetTestDB()
	at := auth_table.CreateAuthTableDriver(dbPointer, adb)
	handler := NewHandler(at)

	for _, tc := range invalidUsernameOrPasswordTestCases{
		log.Print(tc.testCase + " Test Case")
		tc.request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder();
		handler.login(rr, tc.request)

		assert.Equal(t, http.StatusNotAcceptable, rr.Code)
		bod, _ := io.ReadAll(rr.Body)
		assert.Equal(t, "Invalid username/password\n", string(bod))

	}
	t_utils.ResetTestDB()
}


