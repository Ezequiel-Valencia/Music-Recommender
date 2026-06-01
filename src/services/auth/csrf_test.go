package auth

import (
	"music-recommender/config"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
