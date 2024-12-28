package auth

import (
	"fmt"
	"music-recommender/config"
	"music-recommender/db"
	"music-recommender/utils"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandlerFunc func(w http.ResponseWriter, r *http.Request, user db.User)

func RequireAuth(handlerFunc AuthHandlerFunc, adb *db.AbstractDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Do you even have a session cookie?
		encodedSessionCookie, err := r.Cookie(config.StaticEnvs.SessionCookieName)
		if err != nil {
			http.Redirect(w, r, fmt.Sprintf("%s/login", config.StaticEnvs.APIPrefix), http.StatusTemporaryRedirect)
			return
		}

		// Is your session cookie valid, and have a user tied to it?
		var sessionCookie string
		config.SecureCookie.Decode(config.StaticEnvs.SessionCookieName, 
			encodedSessionCookie.Value, &sessionCookie)
		user, err := adb.GetUserFromSessionID(sessionCookie)
		if err != nil || user.Username == "" || sessionCookie == "" {
			if sessionCookie == "" {
				log.Error().Msg("Invalid Signature for Cookie")
			} else{
				log.Error().Msg("No user for session")
			}
			http.Redirect(w, r, fmt.Sprintf("%s/login", config.StaticEnvs.APIPrefix), http.StatusTemporaryRedirect)
			return
		}

		// If all is true continue
		handlerFunc(w, r, user)
	}
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// Stores an unencoded session within the DB, and sends a signed version to the user
func (h *Handler) storeUserSessionAsCookie(w http.ResponseWriter, username string) {
	sessionid, err := h.authTable.GenerateAndStoreSessionID(h.authTable.GetUserStructFromUsername(username))
	signedSession, _ := config.SecureCookie.Encode(config.StaticEnvs.SessionCookieName, sessionid)
	// csrfToken := ""
	if err != nil {
		http.Error(w, "Unable to login user.", http.StatusBadRequest)
	}

	var oneHundred50Days time.Duration = 150 * (time.Hour * 24)

	// Long enough that users don't have to login every time, but also not to long where someone attempting brute force can get in.
	http.SetCookie(w, &http.Cookie{
		Name:     config.StaticEnvs.SessionCookieName,
		Value:    signedSession,
		Expires:  time.Now().Add(oneHundred50Days),
		HttpOnly: true, // Prevents malicious
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		// Path: "/", // Accessible on all paths
		// Domain: "go-server-domain",
	})

	// http.SetCookie(w, &http.Cookie{
	// 	Name: "csrf_token",
	// 	Value: csrfToken,
	// 	Expires: time.Now().Add(oneHundredDays),
	// 	HttpOnly: false, // Needs to be accessed on the client side JS, and put as the X-CSRF-Token
	// 	Path: "/", // Accessible on all paths
	// 	Domain: "go-server-domain",
	// })
}

func GetSessionUser() {}

func validUsernameAndPasswordChars(username string, password string) bool {
	lengthCheck := len(username) > 4 && len(username) < 20 && len(password) > 8 && len(password) < 50
	charsUsedCheck := utils.IsStringAlphaNumeric(username) && utils.IsStringAlphaNumeric(password)
	return lengthCheck && charsUsedCheck
}
