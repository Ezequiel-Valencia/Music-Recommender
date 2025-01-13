package auth

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"music-recommender/config"
	"music-recommender/db/auth_table"
	"music-recommender/types/internal_types/auth_types"
	"music-recommender/utils/t_utils"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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
		httptest.NewRequest("POST", "/api/v1/login", bytes.NewBufferString("username=Ezequiel&password=password1233&email=fake@gmail.com")),
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
	handler.authTable.CreateUser("Ezequiel", "fake@gmail.com", hp, "", auth_types.LocalUserCreationSource.String())

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
	handler.authTable.CreateUser("Ezequiel", "fake@gmail.com", hp, "", auth_types.LocalUserCreationSource.String())

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
	endPointWithAuth := RequireAuth(handler.deleteUser, handler.authTable.AbstractDB)
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
		} else {
			assert.Equal(t, http.StatusTemporaryRedirect, rr.Code)
		}
	}
}

const (
	getRequestWorks int = iota
	loginWorks
	registerWorks
	csrfCookieFails
	csrfHeaderButNotSigned
	csrfHeaderWorks
)

var csrfChecking = []struct {
	testCase     string
	csrfState    int
	csrfRequired bool
}{
	{
		"Get Request Simply Works",
		getRequestWorks,
		false,
	},
	// These cases aren't hit since the require auth middleware is not required for login, and register
	// {
	// 	"Login Works",
	// 	loginWorks,
	// 	false,
	// },
	// {
	// 	"Register Works",
	// 	registerWorks,
	// 	false,
	// },
	{
		"CSRF Fails if it's only the cookie",
		csrfCookieFails,
		true,
	},
	{
		"CSRF Fails if it's only the cookie",
		csrfHeaderButNotSigned,
		true,
	},
	{
		"CSRF Header Works",
		csrfHeaderWorks,
		true,
	},
}

func TestCSRFChecking(t *testing.T) {

	for _, tc := range csrfChecking {
		var request *http.Request
		fakeValue := "value is not checked in this function, but in DB call"
		switch tc.csrfState {
		case getRequestWorks:
			request = httptest.NewRequest(http.MethodGet, "/test", nil)
		case loginWorks:
			request = httptest.NewRequest(http.MethodPost, "/login", nil)
		case registerWorks:
			request = httptest.NewRequest(http.MethodPost, "/register", nil)
		case csrfCookieFails:
			request = httptest.NewRequest(http.MethodPatch, "/test/register", nil)
			request.AddCookie(&http.Cookie{Name: config.StaticEnvs.CSRFCookieName, Value: "value doesn't matter for this function"})
		case csrfHeaderButNotSigned:
			request = httptest.NewRequest(http.MethodDelete, "/test/register", nil)
			request.AddCookie(&http.Cookie{Name: config.StaticEnvs.CSRFCookieName, Value: fakeValue})
		case csrfHeaderWorks:
			request = httptest.NewRequest(http.MethodPost, "/test", nil)
			signedValue, _ := config.SecureCookie.Encode(config.StaticEnvs.CSRFCookieName, fakeValue)
			request.Header.Add(config.StaticEnvs.CSRFHeaderName, signedValue)
		}

		decodedCSRF, isRequired := retrieveCSRFToken(request)
		if tc.csrfState == csrfHeaderWorks {
			assert.NotEqual(t, "", decodedCSRF)
			assert.Equal(t, tc.csrfRequired, isRequired)
		} else {
			assert.Equal(t, "", decodedCSRF)
			assert.Equal(t, tc.csrfRequired, isRequired)
		}
	}

}

func TestPasswordUpdate(t *testing.T) {
	handler := createAuthHandler()
	defer t_utils.ResetTestDB()

	ogPassword, _ := hashPassword("password123")
	handler.authTable.CreateUser("Ezequiel", "Ezequiel@gmail.com", ogPassword, "", auth_types.LocalUserCreationSource.String())
	handler.authTable.CreateUser("Ezequiel2", "Ezequiel2@gmail.com", ogPassword, "", auth_types.LocalUserCreationSource.String())
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
	notPrivilegedUser int = iota
	queryParamNotPresent
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
		queryParamNotPresent,
		http.StatusBadRequest,
		true,
	},
	{
		queryParamNotBoolean,
		http.StatusBadRequest,
		true,
	},
	{
		validRequestToFalse,
		http.StatusOK,
		false,
	},
	{
		validRequestToTrue,
		http.StatusOK,
		true,
	},
}

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
			request = httptest.NewRequest("POST", "/allowCreation", nil)
		case queryParamNotPresent:
			testUser = owner
			request = httptest.NewRequest("POST", "/allowCreation?notPresent=True", nil)
		case queryParamNotBoolean:
			testUser = owner
			request = httptest.NewRequest("POST", "/allowCreation?allowUserCreation=80", nil)
		case validRequestToFalse:
			testUser = owner
			request = httptest.NewRequest("POST", "/allowCreation?allowUserCreation=False", nil)
		case validRequestToTrue:
			testUser = owner
			request = httptest.NewRequest("POST", "/allowCreation?allowUserCreation=True", nil)
		}

		handler.setUserCreationAllowance(rr, request, testUser)
		assert.Equal(t, tc.code, rr.Code)
		assert.Equal(t, tc.allowCreation, config.DynamicEnvs.AllowUserCreation)
	}
}

func createAuthHandler() *Handler {
	adb, dbPointer := t_utils.GetTestDB()
	at := auth_table.CreateAuthTableDriver(dbPointer, adb)
	return NewHandler(at)
}
