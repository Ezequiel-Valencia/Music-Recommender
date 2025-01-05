package auth

import (
	"fmt"
	"music-recommender/config"
	"music-recommender/db"
	"music-recommender/utils"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthHandlerFunc func(w http.ResponseWriter, r *http.Request, user db.User)

func RequireAuth(handlerFunc AuthHandlerFunc, adb *db.AbstractDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Do you even have a session cookie?
		encodedSessionCookie, err := r.Cookie(config.StaticEnvs.SessionCookieName)
		csrfToken, requiresCSRF := retrieveCSRFToken(r)
		
		if err != nil || (csrfToken == "" && requiresCSRF) {
			http.Redirect(w, r, fmt.Sprintf("%s/login", config.StaticEnvs.APIPrefix), http.StatusTemporaryRedirect)
			return
		}

		// Is your session cookie valid, and have a user tied to it?
		var sessionCookie string
		config.SecureCookie.Decode(config.StaticEnvs.SessionCookieName, encodedSessionCookie.Value, &sessionCookie)
		user, err := adb.GetUserFromSessionID(sessionCookie, csrfToken, requiresCSRF)
		if err != nil {
			http.Redirect(w, r, fmt.Sprintf("%s/login", config.StaticEnvs.APIPrefix), http.StatusTemporaryRedirect)
			return
		}

		// If all is true continue
		handlerFunc(w, r, user)
	}
}


// CSRF Has to be set as a header through JS. Otherwise it's still vulnerable to CSRF. Based on assumption that malicious user can't run 
// scripts on browser that impersonate origin of my domain
func retrieveCSRFToken(r *http.Request) (string, bool){
	if r.Method != http.MethodGet && r.Method != http.MethodHead{
		header := r.Header.Get(config.StaticEnvs.CSRFHeaderName)
		if header == ""{
			return "", true
		}
		var decoded string
		config.SecureCookie.Decode(config.StaticEnvs.CSRFCookieName, header, &decoded)
		return decoded, true 
	}
	return "", false
}

// If I do contains instead of equal, it can allow for more white listed path than intended
// func isWhiteListFromCSRFPath(url *url.URL) bool{
// 	return url.Path == "/register" || url.Path == "/login" || url.Path == "register" || url.Path == "login"
// }

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// Stores an unencoded session within the DB, and sends a signed version to the user
func (h *Handler) storeUserSessionAsCookie(w http.ResponseWriter, username string) {
	now := time.Now()
	sessionid, csrfToken, err := h.authTable.GenerateAndStoreSessionID(h.authTable.GetUserStructFromUsername(username), now.UTC().Format(config.StaticEnvs.TimeFormat))
	signedSession, _ := config.SecureCookie.Encode(config.StaticEnvs.SessionCookieName, sessionid)
	signedCSRF, _ := config.SecureCookie.Encode(config.StaticEnvs.CSRFCookieName, csrfToken)
	// csrfToken := ""
	if err != nil {
		http.Error(w, "Unable to login user.", http.StatusBadRequest)
	}

	var oneHundred50Days time.Duration = 150 * (time.Hour * 24)

	// Long enough that users don't have to login every time, but also not to long where someone attempting brute force can get in.
	http.SetCookie(w, &http.Cookie{
		Name:     config.StaticEnvs.SessionCookieName,
		Value:    signedSession,
		Expires:  now.UTC().Add(oneHundred50Days),
		HttpOnly: true, // Prevents malicious
		Secure:   config.DynamicEnvs.CookieDomain != "", // if theres a domain, secure transfer only
		SameSite: http.SameSiteLaxMode,
		Path: "/", // Accessible on all paths
		Domain: config.DynamicEnvs.CookieDomain,
	})

	http.SetCookie(w, &http.Cookie{
		Name: config.StaticEnvs.CSRFCookieName,
		Value: signedCSRF,
		Expires: now.UTC().Add(oneHundred50Days),
		HttpOnly: false, // Needs to be accessed on the client side JS, and put as the X-CSRF-Token
		Secure: config.DynamicEnvs.CookieDomain != "",
		SameSite: http.SameSiteLaxMode,
		Path: "/", // Accessible on all paths
		Domain: config.DynamicEnvs.CookieDomain,
	})
}

func validUsernameAndPasswordChars(username string, password string) bool {
	lengthCheck := len(username) > 4 && len(username) < 15 && len(password) > 8 && len(password) < 30
	charsUsedCheck := utils.IsStringAlphaNumeric(username) && utils.IsStringANWithExtraChars(password)
	return lengthCheck && charsUsedCheck
}
