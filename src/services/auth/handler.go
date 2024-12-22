package auth

import (
	"music-recommender/db"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

// https://github.com/markbates/goth
// https://developer.mozilla.org/en-US/docs/Web/Security/Practical_implementation_guides/CSRF_prevention
// https://www.youtube.com/watch?v=OmLdoEMcr_Y&list=TLPQMTkxMjIwMjTQZqCE3K0mRg&index=5

// Have a cookie which is used in place for their password, that contains user ID, and such (JWT)
// This is cookie is same-origin only.

// Have another which is added as a header to every request, the header being X-CSRF
// This one contains session related information and has to refreshed more frequently

type Handler struct {
	musicDB *db.MusicDB
}

func NewHandler(mdb *db.MusicDB) *Handler {
	return &Handler{mdb}
}

func (h *Handler) RegisterCuratorRoutes(router *mux.Router) {
	router.HandleFunc("/login", h.login).Methods("POST")
	router.HandleFunc("/logout", RequireAuth(h.logout, h.musicDB)).Methods("POST")
	router.HandleFunc("/register", h.register).Methods("POST")
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	// provide credentials to gain an authentication token
	username := r.FormValue("username")
	password := r.FormValue("password")
	// Clean the strings
	username = strconv.Quote(username)
	password = strconv.Quote(password)

	hashedPassword, err := hashPassword(password)
	if err != nil {
		http.Error(w, "Invalid username/password", http.StatusNotAcceptable)
		return
	}

	////////////////////////////
	// Check User Credentials //
	///////////////////////////
	if h.musicDB.ValidUsernameAndCredentials(username, hashedPassword) {
		http.Error(w, "Invalid username/password", http.StatusNotAcceptable)
		return
	}

	sessionid, err := h.musicDB.GenerateAndStoreSessionID(h.musicDB.GetUserStructFromUsername(username))
	// csrfToken := ""
	if err != nil {
		http.Error(w, "Unable to login user.", http.StatusBadRequest)
	}

	var oneHundredDays time.Duration = 100 * (time.Hour * 24)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
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

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
	})

	h.musicDB.RemoveSessionTokens("username", "sessionid")
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var username string = r.FormValue("username")
	var password string = r.FormValue("password")
	var clean_username string = strconv.Quote(username)
	var clean_password string = strconv.Quote(password)
	if (len(username) < 4 || len(password) < 8) && (clean_username != username || clean_password != password) {
		http.Error(w, "Invalid username/password", http.StatusNotAcceptable)
		return
	}

	if user := h.musicDB.GetUserStructFromUsername(username); user.UserId != 0 {
		http.Error(w, "User already exists", http.StatusConflict)
		log.Info().Msgf("%v", user)
		return
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		http.Error(w, "Unknown error.", http.StatusInternalServerError)
		log.Err(err).Msg("Failed to hash password")
		return
	}

	h.musicDB.CreateUser(username, hashedPassword, "", db.LocalUserCreationSource.String())
	log.Info().Msgf("Created User: %s", username)
}

func StoreUserSession() {}

func GetSessionUser() {}
