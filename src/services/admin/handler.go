package admin

import (
	"fmt"
	"music-recommender/config"
	"music-recommender/db/auth_table"
	"music-recommender/db/ranking_table"
	"music-recommender/services/auth"
	"music-recommender/types/internal_types/auth_types"
	"music-recommender/utils"
	"net/http"
	"os"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	authTable          *auth_table.AuthTable
	todaysRankingTable *ranking_table.TodaysRankingDriver
}

func NewHandler(mdb *auth_table.AuthTable, trt *ranking_table.TodaysRankingDriver) *Handler {
	return &Handler{mdb, trt}
}

func (h *Handler) RegisterAdminRoutes(apiRouter *mux.Router, pageRouter *mux.Router) {

	pageRouter.HandleFunc("/adminPage", auth.RequireAuth(h.getCuratorPage, h.authTable.AbstractDB, auth_types.ModeratorPrivileges, auth_types.OneSubmissionRole))

	apiRouter.HandleFunc("/allowUserCreation", auth.RequireAuth(h.setUserCreationAllowance, h.authTable.AbstractDB, auth_types.AdminPrivileges, auth_types.VoterRole)).Methods(http.MethodPost)
	apiRouter.HandleFunc("/setUserRole", auth.RequireAuth(h.setUserRole, h.authTable.AbstractDB, auth_types.ModeratorPrivileges, auth_types.TrustedCuratorRole)).Methods(http.MethodPost)
	apiRouter.HandleFunc("/setUserPrivilege", auth.RequireAuth(h.setUserPrivilege, h.authTable.AbstractDB, auth_types.AdminPrivileges, auth_types.OneSubmissionRole)).Methods(http.MethodPost)
	apiRouter.HandleFunc("/skipSongSelection", auth.RequireAuth(h.skipSongSelection, h.authTable.AbstractDB, auth_types.AdminPrivileges, auth_types.OneSubmissionRole)).Methods(http.MethodPost)

}

func (h *Handler) setUserCreationAllowance(w http.ResponseWriter, r *http.Request, user auth_types.User) {
	if user.UserPrivileges != auth_types.OwnerPrivileges {
		log.Error().Msgf("Username %s, ID %d is attempting to disable user creation.", user.Username, user.UserId)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	err := r.ParseForm()
	if err != nil {
		log.Error().Msg("Problem parsing form to set user creation state.")
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	setState := r.Form.Get("allow-user-creation")
	allowUserCreationBool := setState != ""

	h.authTable.SetAbilityForUserCreation(allowUserCreationBool)
	w.Header().Add("HX-Refresh", "true")
}

func (h *Handler) setUserRole(w http.ResponseWriter, r *http.Request, user auth_types.User) {
	err := r.ParseForm()
	if err != nil {
		return
	}
	username := r.Form.Get("username-for-role")
	if !utils.IsStringAlphaNumeric(username) {
		http.Error(w, "Misformed username.", http.StatusBadRequest)
		return
	}
	role := auth_types.StringToUserRoles(r.Form.Get("user-role"))
	if role.EnumIndex() > user.UserRole.EnumIndex() {
		http.Error(w, "Bad request", http.StatusBadRequest)
		log.Error().Msgf("User %s just tried to set %s to a role higher than themselves!", user.Username, username)
		return
	}

	h.authTable.SetUserRole(username, role)
	w.Header().Add("HX-Refresh", "true")
}

func (h *Handler) setUserPrivilege(w http.ResponseWriter, r *http.Request, user auth_types.User) {
	err := r.ParseForm()
	if err != nil {
		return
	}

	username := r.Form.Get("username-for-privilege")
	if !utils.IsStringAlphaNumeric(username) {
		http.Error(w, "Misformed username.", http.StatusBadRequest)
		return
	}

	privilege := auth_types.StringToUserPrivileges(r.Form.Get("user-privilege"))
	if privilege.EnumIndex() > auth_types.AdminPrivileges.EnumIndex() {
		http.Error(w, "Bad request", http.StatusBadRequest)
		log.Error().Msgf("User %s just tried to set %s to a role higher than Admin!", user.Username, username)
		return
	}

	h.authTable.SetUserPrivilege(username, privilege)
	w.Header().Add("HX-Refresh", "true")
}

func (h *Handler) skipSongSelection(w http.ResponseWriter, r *http.Request, user auth_types.User) {
	log.Info().Msg("Skipped current song selection.")
	h.todaysRankingTable.CleanTodaysRanking()
	h.todaysRankingTable.SelectNewSongs()
}

func (h *Handler) getCuratorPage(w http.ResponseWriter, r *http.Request, user auth_types.User) {
	dir, _ := os.Getwd()

	tempFuncs := template.FuncMap{"privilegeToInt": func(i string) int {
		return auth_types.StringToUserPrivileges(i).EnumIndex()
	},
		"greaterThanOrEqual": func(x int, y int) bool {
			return x >= y
		},
	}

	templateLocation := fmt.Sprintf("%s/templates/admin.html", dir)
	adminTemplate, err := template.New("admin.html").Funcs(tempFuncs).ParseFiles(templateLocation)
	if err != nil {
		log.Err(err).Msg("Problem templating admin page.")
	}

	templateMap := map[string]any{
		"Username":          user.Username,
		"Privilege":         user.UserPrivileges.String(),
		"PrivilegeInt":      user.UserPrivileges.EnumIndex(),
		"CreationIsAllowed": config.DynamicEnvs.AllowUserCreation, // It gets initialized when server state is read from DB, and changes when updates are made
	}
	if err := adminTemplate.Execute(w, templateMap); err != nil {
		log.Err(err).Msg("Problem executing admin template.")
	}
}
