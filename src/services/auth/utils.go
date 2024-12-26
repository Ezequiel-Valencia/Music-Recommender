package auth

import (
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
		user, err := adb.GetUserFromSessionID(r)
		if err != nil || user.Username == "" {
			if err != nil {
				log.Err(err).Msg("User is not authenticated!")
			} else {
				log.Error().Msg("User session is not valid!")
			}
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		handlerFunc(w, r, user)
	}
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func (h *Handler) storeUserSession(w http.ResponseWriter, username string) {
	sessionid, err := h.authTable.GenerateAndStoreSessionID(h.authTable.GetUserStructFromUsername(username))
	// csrfToken := ""
	if err != nil {
		http.Error(w, "Unable to login user.", http.StatusBadRequest)
	}

	var oneHundredDays time.Duration = 100 * (time.Hour * 24)

	http.SetCookie(w, &http.Cookie{
		Name:     config.Envs.SessionCookieName,
		Value:    sessionid,
		Expires:  time.Now().Add(oneHundredDays),
		HttpOnly: true, // Prevents malicious
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
