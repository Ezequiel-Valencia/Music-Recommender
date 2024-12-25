package auth

import (
	"music-recommender/config"
	"music-recommender/db"
	"music-recommender/db/auth_table"
	"music-recommender/types"
	"music-recommender/utils"
	"net/http"
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
	authTable *auth_table.AuthTable
}

func NewHandler(mdb *auth_table.AuthTable) *Handler {
	return &Handler{mdb}
}

func (h *Handler) RegisterAuthRoutes(router *mux.Router) {
	router.HandleFunc("/user", RequireAuth(h.loggedInUserInfo, h.authTable.AbstractDB)).Methods("GET")
	router.HandleFunc("/user", RequireAuth(h.deleteUser, h.authTable.AbstractDB)).Methods("DELETE")
	router.HandleFunc("/login", h.login).Methods("POST")
	router.HandleFunc("/logout", RequireAuth(h.logout, h.authTable.AbstractDB)).Methods("POST")
	router.HandleFunc("/register", h.register).Methods("POST")
}

func (h *Handler) loggedInUserInfo(w http.ResponseWriter, r *http.Request, user db.User) {
	utils.WriteJSON(w, types.UserDTO{Username: user.Username,
		CreationDate: user.CreationDate.Format(time.DateOnly), Role: user.UserRole.String()}, 200)
}

func (h *Handler) deleteUser(w http.ResponseWriter, r *http.Request, user db.User) {

	h.authTable.DeleteUser(user.Username, user.UserId)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	// provide credentials to gain an authentication token
	username := r.FormValue("username")
	password := r.FormValue("password")

	if _, err := r.Cookie(config.Envs.SessionCookieName); err != http.ErrNoCookie {
		http.Error(w, "User already logged in.", http.StatusBadRequest)
		return
	}

	////////////////////////////
	// Check User Credentials //
	///////////////////////////
	// if not valid
	if !utils.IsStringAlphaNumeric(username) || !utils.IsStringAlphaNumeric(password) || !h.authTable.CorrectUsernameAndPassword(username, password) {
		http.Error(w, "Invalid username/password", http.StatusNotAcceptable)
		return
	}

	h.storeUserSession(w, username)

}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request, user db.User) {
	sessionCookie, err := r.Cookie(config.Envs.SessionCookieName)
	if err != nil {
		http.Error(w, "Can't logout", 500)
	}
	err = h.authTable.RemoveSessionTokens(user, sessionCookie.Value)
	if err != nil {
		http.Error(w, "Can't logout", 500)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     config.Envs.SessionCookieName,
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
	})

}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var username string = r.FormValue("username")
	var password string = r.FormValue("password")
	if len(username) < 4 || len(password) < 8 || !utils.IsStringAlphaNumeric(username) || !utils.IsStringAlphaNumeric(password) {
		http.Error(w, "Invalid username/password", http.StatusNotAcceptable)
		return
	}

	if user := h.authTable.GetUserStructFromUsername(username); user.UserId != 0 {
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

	h.authTable.CreateUser(username, hashedPassword, "", db.LocalUserCreationSource.String())
	h.storeUserSession(w, username)
}
