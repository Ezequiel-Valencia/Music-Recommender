package auth

import (
	"bytes"
	"fmt"
	"io"
	"music-recommender/types/internal_types/auth_types"
	"music-recommender/utils/t_utils"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var registerTestCases = []struct {
	testCase         string
	request          *http.Request
	code             int
	expectedResponse string
}{
	{
		"Invalid Username Chars",
		httptest.NewRequest("POST", "/api/v1/register", bytes.NewBufferString(fmt.Sprintf("username=Ez@%s&password=password123&email=fake@gmail.com", t_utils.GenerateRandomRuneString(5, false)))),
		http.StatusNotAcceptable,
		"Invalid username/password\n",
	},
	{
		"Invalid Password Chars",
		httptest.NewRequest("POST", "/api/v1/register", bytes.NewBufferString(fmt.Sprintf("username=Ezequiel&password=password123@%s&email=fake@gmail.com", t_utils.GenerateRandomRuneString(5, false)))),
		http.StatusNotAcceptable,
		"Invalid username/password\n",
	},
	{
		"Invalid Email Chars",
		httptest.NewRequest("POST", "/api/v1/register", t_utils.CreateHTTPBodyURLEncoded(fmt.Sprintf("username=Ezequiel&password=password123&email=fake@@gmail.com%s", t_utils.GenerateRandomRuneString(5, true)))),
		http.StatusNotAcceptable,
		"Invalid username/password\n",
	},
	{
		"User Already Exists",
		httptest.NewRequest("POST", "/api/v1/register", t_utils.CreateHTTPBodyURLEncoded("username=Ezequiel&password=password123&email=fake2@gmail.com")),
		http.StatusConflict,
		"Username or email already exists\n",
	},
	{
		"Email Already Exists",
		httptest.NewRequest("POST", "/api/v1/register", t_utils.CreateHTTPBodyURLEncoded("username=Ezequiel2&password=password123&email=fake@gmail.com")),
		http.StatusConflict,
		"Username or email already exists\n",
	},
	{
		"Create User, All Good",
		httptest.NewRequest("POST", "/api/v1/register", bytes.NewBufferString("username=Ezequiel2&password=password123&email=fake2@gmail.com")),
		http.StatusOK,
		"Ezequiel2", // User object in JSON
	},
}

func TestRegister(t *testing.T) {
	handler := createAuthHandler()
	defer t_utils.ResetTestDB()

	hp, _ := hashPassword("password123")
	_ = handler.authTable.CreateUser("Ezequiel", "fake@gmail.com", hp, "", auth_types.LocalUserCreationSource.String())

	for _, tc := range registerTestCases {
		tc.request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler.register(rr, tc.request)

		assert.Equal(t, tc.code, rr.Code)
		if rr.Code == http.StatusOK {
			session := rr.Result().Header.Get("Set-Cookie")
			assert.NotEqual(t, "", session)

			user := handler.authTable.GetUserStructFromUsername("Ezequiel2")
			assert.Equal(t, 2, user.UserId) // Second user created
			bod, _ := io.ReadAll(rr.Body)
			assert.True(t, strings.Contains(string(bod), tc.expectedResponse))
		} else {
			bod, _ := io.ReadAll(rr.Body)
			assert.Equal(t, tc.expectedResponse, string(bod))
		}
	}

}

func TestPasswordUpdate(t *testing.T) {
	handler := createAuthHandler()
	defer t_utils.ResetTestDB()

	ogPassword, _ := hashPassword("password123")
	_ = handler.authTable.CreateUser("Ezequiel", "Ezequiel@gmail.com", ogPassword, "", auth_types.LocalUserCreationSource.String())
	_ = handler.authTable.CreateUser("Ezequiel2", "Ezequiel2@gmail.com", ogPassword, "", auth_types.LocalUserCreationSource.String())
	user := handler.authTable.GetUserStructFromUsername("Ezequiel2")

	// Both passwords are the same, nothing has changed
	assert.True(t, handler.authTable.CorrectEmailAndPassword("Ezequiel@gmail.com", "password123"), "")
	assert.True(t, handler.authTable.CorrectEmailAndPassword("Ezequiel2@gmail.com", "password123"), "")
	assert.False(t, handler.authTable.CorrectEmailAndPassword("Ezequiel2@gmail.com", "other890"), "")

	rr := httptest.NewRecorder()
	request := httptest.NewRequest("POST", "/api/v1/register", t_utils.CreateHTTPBodyURLEncoded("email=Ezequiel2@gmail.com&password=password123&newPassword=other890"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	handler.updatePassword(rr, request, user)

	// Only the password of the user making the update has changed, and to it's appropriate value
	assert.True(t, handler.authTable.CorrectEmailAndPassword("Ezequiel@gmail.com", "password123"), "")
	assert.False(t, handler.authTable.CorrectEmailAndPassword("Ezequiel2@gmail.com", "password123"), "")
	assert.True(t, handler.authTable.CorrectEmailAndPassword("Ezequiel2@gmail.com", "other890"), "")
}

func TestDeleteUser(t *testing.T) {
	handler := createAuthHandler()
	defer t_utils.ResetTestDB()

	request := httptest.NewRequest("POST", "/api/v1/register", t_utils.CreateHTTPBodyURLEncoded("username=Ezequiel&password=password123&email=fake@gmail.com"))
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
