package main_test

import (
	"bytes"
	"context"
	"fmt"
	"music-recommender/api"
	"music-recommender/config"
	"music-recommender/utils/t_utils"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)



var allEndPoints = []struct{
	Endpoint		string
	RequestType		string
	RequiresAuth	bool
}{
	// Auth Service
	{Endpoint: "/user", RequestType: http.MethodGet, RequiresAuth: true},
	{Endpoint: "/user", RequestType: http.MethodDelete, RequiresAuth: true},
	{Endpoint: "/logout", RequestType: http.MethodPost, RequiresAuth: true},

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

// Indirectly also tests the register endpoint and the sessions it creates
func TestAllEndpointsAuthRequirements(t *testing.T){
	req, _ := http.NewRequest(http.MethodPost,
		fmt.Sprintf("http://%s/api/v1/register", config.DynamicEnvs.HostAndPort),
		bytes.NewReader([]byte("username=Ezequiel&password=password123")))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, _ := http.DefaultClient.Do(req)
	sessionCookie := res.Cookies()[0]

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	for _, tc := range allEndPoints{
		reqString := fmt.Sprintf("http://%s%s%s", config.DynamicEnvs.HostAndPort, config.StaticEnvs.APIPrefix, tc.Endpoint)
		req, _ = http.NewRequest(tc.RequestType, reqString, nil)
		res, _ := client.Do(req)
		if tc.RequiresAuth{
			assert.Equal(t, http.StatusTemporaryRedirect, res.StatusCode)
			req.AddCookie(sessionCookie)
			res, _ = http.DefaultClient.Do(req)
			assert.NotEqual(t, http.StatusTemporaryRedirect, res.StatusCode)
		} else{
			assert.NotEqual(t, http.StatusTemporaryRedirect, res.StatusCode)
		}
	}
}

