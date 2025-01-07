package main

import (
	"bytes"
	"context"
	"fmt"
	"music-recommender/api"
	"music-recommender/config"
	"music-recommender/db"
	"music-recommender/utils/t_utils"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var allEndPoints = []struct {
	Endpoint     string
	RequestType  string
	RequiresAuth bool
}{
	// Auth Service
	{Endpoint: "/user", RequestType: http.MethodGet, RequiresAuth: true},
	{Endpoint: "/user", RequestType: http.MethodDelete, RequiresAuth: true},
	{Endpoint: "/logout", RequestType: http.MethodPost, RequiresAuth: true},
	{Endpoint: "/passwd", RequestType: http.MethodPatch, RequiresAuth: true},
	{Endpoint: "/allowUserCreation", RequestType: http.MethodPost, RequiresAuth: true},

	{Endpoint: "/login", RequestType: http.MethodPost, RequiresAuth: false},
	{Endpoint: "/register", RequestType: http.MethodPost, RequiresAuth: false},
}

func TestMain(m *testing.M) {
	config.DynamicEnvs.HostAndPort = "localhost:9999"
	adb, db := t_utils.GetTestDB()
	apiSever := api.CreateMainServer(db, adb)
	go apiSever.ListenAndServe()

	m.Run()

	apiSever.Shutdown(context.TODO())
	t_utils.TearDownTestDB()
}

func getUserCookies(login bool)(*http.Cookie, *http.Cookie){
	var endpoint string
	if (login){
		endpoint = "/login"
	} else{
		endpoint = "/register"
	}
	req, _ := http.NewRequest(http.MethodPost,
		fmt.Sprintf("http://%s/api/v1%s", config.DynamicEnvs.HostAndPort, endpoint),
		bytes.NewReader([]byte("username=Ezequiel&password=password123")))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, _ := http.DefaultClient.Do(req)
	return res.Cookies()[0], res.Cookies()[1]
}

// Indirectly also tests the register endpoint and the sessions it creates
func TestAllEndpointsAuthRequirements(t *testing.T) {
	defer t_utils.ResetTestDB()
	var req *http.Request
	sessionCookie, csrfCookie := getUserCookies(false)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	for _, tc := range allEndPoints {
		reqString := fmt.Sprintf("http://%s%s%s", config.DynamicEnvs.HostAndPort, config.StaticEnvs.APIPrefix, tc.Endpoint)
		req, _ = http.NewRequest(tc.RequestType, reqString, nil)
		res, _ := client.Do(req)
		if tc.RequiresAuth {
			assert.Equal(t, http.StatusTemporaryRedirect, res.StatusCode)
			req, _ = http.NewRequest(tc.RequestType, reqString, nil)
			req.AddCookie(sessionCookie)
			req.Header.Add(config.StaticEnvs.CSRFHeaderName, csrfCookie.Value)
			res, _ = http.DefaultClient.Do(req)
			assert.NotEqual(t, http.StatusTemporaryRedirect, res.StatusCode)

			// If the endpoint deletes the user or logs them out, we have to validate it again
			if ((tc.Endpoint == "/user" && tc.RequestType == "DELETE") || tc.Endpoint == "/logout"){
				sessionCookie, csrfCookie = getUserCookies(tc.Endpoint == "/logout")
			}
		} else {
			assert.NotEqual(t, http.StatusTemporaryRedirect, res.StatusCode)
		}
	}
}

func TestDailyDBTask(t *testing.T) {
	_, dbPointer := t_utils.GetTestDB()
	defer t_utils.ResetTestDB()

	nowTime := time.Now()
	twoHundredDaysAgo := -1 * (200 * 24) * time.Hour
	testUser := db.User{UserId: 1, Username: "Tester", CreationDate: nowTime.Add(twoHundredDaysAgo)}
	t_utils.CreateFakeUser(dbPointer, &testUser, "password")

	// Old Session
	dbPointer.Exec(`INSERT INTO sessions(user_id, session_id, csrf_token, creation_date) 
	VALUES($1, $2, $3, $4)`, testUser.UserId, "ses", "crf", nowTime.Add(twoHundredDaysAgo).UTC().Format(config.StaticEnvs.TimeFormat))

	// New Session
	dbPointer.Exec(`INSERT INTO sessions(user_id, session_id, csrf_token, creation_date) 
	VALUES($1, $2, $3, $4)`, testUser.UserId, "ses", "crf", nowTime.UTC().Format(config.StaticEnvs.TimeFormat))

	res, _ := dbPointer.Exec("SELECT * FROM sessions")
	resNum, _ := res.RowsAffected()
	var expectedNum int64 = 2
	assert.Equal(t, expectedNum, resNum)

	dbCleanUp(dbPointer)

	res, _ = dbPointer.Exec("SELECT * FROM sessions")
	resNum, _ = res.RowsAffected()
	expectedNum = 1
	assert.Equal(t, expectedNum, resNum)

	var sessionCreationDate string
	dbPointer.QueryRow("SELECT creation_date FROM sessions").Scan(&sessionCreationDate)
	assert.Equal(t, nowTime.UTC().Format(config.StaticEnvs.TimeFormat), sessionCreationDate)
}

func TestMoreThanOneOwnerChecker(t *testing.T){
	_, dbPointer := t_utils.GetTestDB()
	defer t_utils.ResetTestDB()


	t_utils.CreateFakeUser(dbPointer, &db.User{Username: "OGOwner", UserRole: db.UnlimitedRole, UserPrivileges: db.OwnerPrivileges}, "passwd")
	assert.False(t, isThereMoreThanOneOwner(dbPointer), "There is one owner")
	t_utils.CreateFakeUser(dbPointer, &db.User{Username: "Oscar", UserPrivileges: db.OwnerPrivileges}, "passwd")
	assert.True(t, isThereMoreThanOneOwner(dbPointer), "Oscar wants to be owner too.")
	dbPointer.Exec("DELETE FROM users WHERE username = 'Oscar'")
	assert.False(t, isThereMoreThanOneOwner(dbPointer), "Back to one. Oscar is banned.")
}
