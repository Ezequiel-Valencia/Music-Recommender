package curator_service

import (
	"fmt"
	"html/template"
	"music-recommender/db/music_table"
	"music-recommender/services/auth"
	"music-recommender/types/communication_types"
	"music-recommender/types/internal_types/auth_types"
	"music-recommender/utils"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

type Handler struct {
	musicDB *music_table.MusicTable
}

func NewHandler(mdb *music_table.MusicTable) *Handler {
	return &Handler{mdb}
}

func (h *Handler) RegisterCuratorRoutes(apiRouter *mux.Router, webPageRouter *mux.Router) {
	apiRouter.HandleFunc("/submitMusic", auth.RequireAuth(h.submitMusic, h.musicDB.AbstractDB, auth_types.NoPrivileges, auth_types.OneSubmissionRole)).Methods("POST")

	webPageRouter.HandleFunc("/curatorPage", auth.RequireAuth(h.curatorPage, h.musicDB.AbstractDB, auth_types.NoPrivileges, auth_types.OneSubmissionRole)).Methods("GET")
}

func (h *Handler) curatorPage(w http.ResponseWriter, r *http.Request, user auth_types.User){
	dir, _ := os.Getwd()
	template := template.Must(template.ParseFiles(fmt.Sprintf("%s/templates/curator.html", dir)))
	templateMap := map[string]any{
		"User": user.Username,
		"Iterations": [...]int{1, 2, 3},
	}
	template.Execute(w, templateMap)
	
}


func (h *Handler) submitMusic(w http.ResponseWriter, r *http.Request, user auth_types.User) {
	// submit music to be chosen to the data base.
	var submitSong communication_types.SubmitSongSet
	err := utils.DecodeJSONBody(w, r, &submitSong)
	if err != nil {
		return
	}
	h.musicDB.InsertSongSet(&submitSong, user)
}
