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
	"github.com/rs/zerolog/log"
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

func (h *Handler) curatorPage(w http.ResponseWriter, r *http.Request, user auth_types.User) {
	dir, _ := os.Getwd()

	incFunc := template.FuncMap{"inc": func(i int) int {
		return i + 1
	}}
	templateLocation := fmt.Sprintf("%s/templates/curator.html", dir)
	curatorTemplate, err := template.New("curator.html").Funcs(incFunc).ParseFiles(templateLocation)
	if err != nil {
		log.Err(err).Msg("Problem templating curator page.")
	}

	alreadySubmitted := h.musicDB.GetUserSubmissionsToBeRanked(user)
	submissionsString := fmt.Sprintf("%d / %d", len(alreadySubmitted), user.UserRole.GetRolesSubmissionLimit())

	templateMap := map[string]any{
		"User":             user.Username,
		"Iterations":       [...]int{1, 2, 3},
		"Submissions":      submissionsString,
		"AlreadySubmitted": alreadySubmitted,
	}
	if err := curatorTemplate.Execute(w, templateMap); err != nil {
		log.Err(err).Msg("Problem executing curator template.")
	}

}

func (h *Handler) submitMusic(w http.ResponseWriter, r *http.Request, user auth_types.User) {
	// submit music to be chosen to the data base.
	submitSong := communication_types.SubmitSongSet{}
	err := r.ParseForm()
	if err != nil {
		return
	}

	for i := range 3 {
		songName := r.Form.Get(fmt.Sprintf("song-name-%d", i+1))
		artist := r.Form.Get(fmt.Sprintf("artist-%d", i+1))
		url := r.Form.Get(fmt.Sprintf("url-%d", i+1))

		songCheck := len(songName) < 50 && len(songName) > 2 && utils.IsStringAlphaNumericWithPunctuation(songName)
		artistCheck := len(artist) < 50 && len(artist) > 2 && utils.IsStringAlphaNumericWithPunctuation(artist)
		urlCheck := len(url) < 75 && utils.IsProperYouTubeLink(url)
		if songCheck && artistCheck && urlCheck {
			submitSong.Songs = append(submitSong.Songs, communication_types.SubmitSong{Name: songName, Artist: artist, SongURL: url})
		} else {
			log.Error().Msgf("Improper submission by %s", user.Username)
			http.Error(w, "Improper submission.", http.StatusBadRequest)
			return
		}
	}

	description := r.Form.Get("description-box")

	if len(description) < 300 && len(description) > 5 && utils.IsStringAlphaNumericWithPunctuation(description) {
		submitSong.Description = description
	} else {
		log.Error().Msgf("Improper description by %s", user.Username)
		http.Error(w, "Improper description.", http.StatusBadRequest)
		return
	}

	if err := h.musicDB.InsertSongSet(&submitSong, user); err != nil {
		http.Error(w, "Failed to submit music.", http.StatusBadRequest)
		return
	}

	w.Header().Add("HX-Refresh", "true")
}
