package auth

import (
	"music-recommender/config"
	"music-recommender/db"
	"music-recommender/db/auth_table"
	"music-recommender/types"
	"music-recommender/utils"
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
	authTable *auth_table.AuthTable
}

func NewHandler(mdb *auth_table.AuthTable) *Handler {
	return &Handler{mdb}
}

func (h *Handler) RegisterAuthRoutes(router *mux.Router) {
	router.HandleFunc("/user", RequireAuth(h.loggedInUserInfo, h.authTable.AbstractDB)).Methods(http.MethodGet)
	router.HandleFunc("/user", RequireAuth(h.deleteUser, h.authTable.AbstractDB)).Methods(http.MethodDelete)
	
	router.HandleFunc("/login", h.login).Methods(http.MethodPost)
	router.HandleFunc("/logout", RequireAuth(h.logout, h.authTable.AbstractDB)).Methods(http.MethodPost)
	router.HandleFunc("/register", h.register).Methods(http.MethodPost)
	
	router.HandleFunc("/allowUserCreation", RequireAuth(h.setUserCreationAllowance, h.authTable.AbstractDB)).Methods(http.MethodPost)

	router.HandleFunc("/passwd", RequireAuth(h.updatePassword, h.authTable.AbstractDB)).Methods(http.MethodPatch)
}

func (h *Handler) loggedInUserInfo(w http.ResponseWriter, r *http.Request, user db.User) {
	utils.WriteJSON(w, types.UserDTO{Username: user.Username,
		CreationDate: user.CreationDate.Format(time.DateOnly), Role: user.UserRole.String()}, 200)
}

func (h *Handler) deleteUser(w http.ResponseWriter, r *http.Request, user db.User) {
	h.authTable.DeleteUser(user.Username, user.UserId)
}

func (h *Handler) updatePassword(w http.ResponseWriter, r *http.Request, user db.User){
	// provide credentials to assure it is them
	username := r.FormValue("username")
	password := r.FormValue("password")
	newPassword := r.FormValue("newPassword")

	////////////////////////////
	// Check User Credentials //
	///////////////////////////
	// if not valid
	if !validUsernameAndPasswordChars(username, password) || !h.authTable.CorrectUsernameAndPassword(username, password) {
		http.Error(w, "Invalid username/password. Can't update password.", http.StatusNotAcceptable)
		return
	}

	hashedPassword, err := hashPassword(newPassword)
	if err != nil{
		http.Error(w, "Invalid username/password", http.StatusNotAcceptable)
	}
	h.authTable.UpdatePassword(user, hashedPassword)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	// provide credentials to gain an authentication token
	username := r.FormValue("username")
	password := r.FormValue("password")

	if _, err := r.Cookie(config.StaticEnvs.SessionCookieName); err != http.ErrNoCookie {
		http.Error(w, "User already logged in.", http.StatusBadRequest)
		return
	}

	////////////////////////////
	// Check User Credentials //
	///////////////////////////
	// if not valid
	if !validUsernameAndPasswordChars(username, password) || !h.authTable.CorrectUsernameAndPassword(username, password) {
		http.Error(w, "Invalid username/password", http.StatusNotAcceptable)
		return
	}

	if h.authTable.ReachedMaxNumberOfSessionsForUser(h.authTable.GetUserStructFromUsername(username)){
		http.Error(w, "Max number of logins reached. Logout on one device.", http.StatusTooManyRequests)
		return
	}

	h.storeUserSessionAsCookie(w, username)
	user := h.authTable.GetUserStructFromUsername(username)
	utils.WriteJSON(w, types.UserDTO{Username: username,
		CreationDate: user.CreationDate.Format(config.StaticEnvs.TimeFormat), 
		Role: user.UserRole.String()}, 200)

}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request, user db.User) {
	sessionCookie, err := r.Cookie(config.StaticEnvs.SessionCookieName)
	if err != nil {
		http.Error(w, "Can't logout", http.StatusUnauthorized)
		return
	}
	var decodedCookie string = ""
	config.SecureCookie.Decode(config.StaticEnvs.SessionCookieName, sessionCookie.Value, &decodedCookie)
	err = h.authTable.RemoveSessionTokens(user, decodedCookie)
	if err != nil {
		http.Error(w, "Can't logout", http.StatusBadRequest)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     config.StaticEnvs.SessionCookieName,
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
		Domain: config.DynamicEnvs.CookieDomain,
	})

}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var username string = r.FormValue("username")
	var password string = r.FormValue("password")

	if !config.DynamicEnvs.AllowUserCreation{
		http.Error(w, "User creation is not allowed.", http.StatusMethodNotAllowed)
		return
	}

	if !validUsernameAndPasswordChars(username, password) {
		http.Error(w, "Invalid username/password", http.StatusNotAcceptable)
		return
	}

	if user := h.authTable.GetUserStructFromUsername(username); user.UserId != 0 {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		http.Error(w, "Unknown error.", http.StatusInternalServerError)
		log.Err(err).Msg("Failed to hash password")
		return
	}

	h.authTable.CreateUser(username, hashedPassword, "", db.LocalUserCreationSource.String())
	h.storeUserSessionAsCookie(w, username)
	user := h.authTable.GetUserStructFromUsername(username)
	utils.WriteJSON(w, types.UserDTO{Username: username,
		CreationDate: user.CreationDate.Format(config.StaticEnvs.TimeFormat), 
		Role: user.UserRole.String()}, 200)
}


func (h *Handler) setUserCreationAllowance(w http.ResponseWriter, r *http.Request, user db.User){
	if (user.UserPrivileges != db.OwnerPrivileges){
		log.Error().Msgf("Username %s, ID %d is attempting to disable user creation.", user.Username, user.UserId)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	queryValues := r.URL.Query()
	setState := queryValues.Get("allowUserCreation")
	if setState == ""{
		log.Error().Msg("Malformed request to disable account creation.")
		http.Error(w, "Malformed request", http.StatusBadRequest)
		return
	}
	allowUserCreationBool, err := strconv.ParseBool(setState)
	if err != nil {
		log.Error().Msg("Problem parsing boolean to set user creation state.")
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	h.authTable.SetAbilityForUserCreation(allowUserCreationBool)
}
