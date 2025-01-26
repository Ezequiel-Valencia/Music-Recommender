package auth

import (
	"music-recommender/config"
	"music-recommender/db/auth_table"
	"music-recommender/types/communication_types"
	"music-recommender/types/internal_types/auth_types"
	"music-recommender/utils"
	"net/http"
	"time"

	// "github.com/badoux/checkmail"
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
	router.HandleFunc("/user", RequireAuthMinimumPrivileges(h.loggedInUserInfo, h.authTable.AbstractDB)).Methods(http.MethodGet)
	router.HandleFunc("/user", RequireAuthMinimumPrivileges(h.deleteUser, h.authTable.AbstractDB)).Methods(http.MethodDelete)

	router.HandleFunc("/login", h.login).Methods(http.MethodPost)
	router.HandleFunc("/logout", RequireAuthMinimumPrivileges(h.logout, h.authTable.AbstractDB)).Methods(http.MethodPost)
	router.HandleFunc("/register", h.register).Methods(http.MethodPost)

	router.HandleFunc("/passwd", RequireAuthMinimumPrivileges(h.updatePassword, h.authTable.AbstractDB)).Methods(http.MethodPatch)
}

func (h *Handler) loggedInUserInfo(w http.ResponseWriter, r *http.Request, user auth_types.User) {
	utils.WriteJSON(w, communication_types.UserDTO{Username: user.Username,
		CreationDate: user.CreationDate.Format(time.DateOnly), Role: user.UserRole.String()}, 200)
}

func (h *Handler) deleteUser(w http.ResponseWriter, r *http.Request, user auth_types.User) {
	h.authTable.DeleteUser(user.Username, user.UserId)
}

func (h *Handler) updatePassword(w http.ResponseWriter, r *http.Request, user auth_types.User) {
	// provide credentials to assure it is them
	email := r.FormValue("email")
	password := r.FormValue("password")
	newPassword := r.FormValue("newPassword")

	////////////////////////////
	// Check User Credentials //
	///////////////////////////
	// if not valid
	validChars := validPasswordChars(password) && validEmailChars(email)
	if !validChars || !h.authTable.CorrectEmailAndPassword(email, password) {
		http.Error(w, "Invalid username/password. Can't update password.", http.StatusNotAcceptable)
		return
	}

	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		http.Error(w, "Invalid username/password", http.StatusNotAcceptable)
	}
	h.authTable.UpdatePassword(user, hashedPassword)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	// provide credentials to gain an authentication token
	email := r.FormValue("email")
	password := r.FormValue("password")

	if _, err := r.Cookie(config.StaticEnvs.SessionCookieName); err != http.ErrNoCookie {
		http.Error(w, "User already logged in. Please clear cookies for a new valid session.", http.StatusBadRequest)
		return
	}

	////////////////////////////
	// Check User Credentials //
	///////////////////////////
	// if not valid
	validChars := validEmailChars(email) && validPasswordChars(password)
	if !validChars || !h.authTable.CorrectEmailAndPassword(email, password) {
		http.Error(w, "Invalid username/password", http.StatusNotAcceptable)
		return
	}

	loginUser := h.authTable.GetUserStructFromEmail(email)
	if h.authTable.ReachedMaxNumberOfSessionsForUser(loginUser) {
		http.Error(w, "Max number of logins reached. Logout on one device.", http.StatusTooManyRequests)
		return
	}

	h.storeUserSessionAsCookie(w, loginUser.Username)
	utils.WriteJSON(w, communication_types.UserDTO{Username: loginUser.Username,
		CreationDate: loginUser.CreationDate.Format(config.StaticEnvs.TimeFormat),
		Role:         loginUser.UserRole.String()}, 200)

}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request, user auth_types.User) {
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
		Expires:  time.Now().AddDate(0, 0, -1),
		HttpOnly: true,
		Secure:   config.DynamicEnvs.CookieDomain != "",
		SameSite: http.SameSiteLaxMode,
		Path:     "/", // Accessible on all paths
		Domain:   config.DynamicEnvs.CookieDomain,
		MaxAge:   -1,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     config.StaticEnvs.CSRFCookieName,
		Value:    "",
		Expires:  time.Now().AddDate(0, 0, -1),
		HttpOnly: false,
		Secure:   config.DynamicEnvs.CookieDomain != "",
		SameSite: http.SameSiteLaxMode,
		Path:     "/", // Accessible on all paths
		Domain:   config.DynamicEnvs.CookieDomain,
		MaxAge:   -1,
	})

}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var username string = r.FormValue("username")
	var email string = r.FormValue("email")
	var password string = r.FormValue("password")

	if !config.DynamicEnvs.AllowUserCreation {
		http.Error(w, "User creation is not allowed.", http.StatusMethodNotAllowed)
		return
	}

	validChars := validEmailChars(email) && validUsernameChars(username) && validPasswordChars(password)
	if !validChars {
		http.Error(w, "Invalid username/password", http.StatusNotAcceptable)
		return
	}

	// if emailErr := checkmail.ValidateHost(email); emailErr != nil {
	// 	http.Error(w, "Invalid email.", http.StatusNotAcceptable)
	// 	return
	// }

	if !h.authTable.IsTheUsernameAndEmailUnique(username, email) {
		http.Error(w, "Username or email already exists", http.StatusConflict)
		return
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		http.Error(w, "Unknown error.", http.StatusInternalServerError)
		log.Err(err).Msg("Failed to hash password")
		return
	}

	h.authTable.CreateUser(username, email, hashedPassword, "", auth_types.LocalUserCreationSource.String())
	h.storeUserSessionAsCookie(w, username)
	user := h.authTable.GetUserStructFromUsername(username)
	utils.WriteJSON(w, communication_types.UserDTO{Username: username,
		CreationDate: user.CreationDate.Format(config.StaticEnvs.TimeFormat),
		Role:         user.UserRole.String()}, 200)
}

