package music_curator

import (
	"music-recommender/db"
	"music-recommender/types"
	"music-recommender/utils"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

type Handler struct{
	musicDB *db.MusicDB
}


func NewHandler(mdb *db.MusicDB) *Handler{
	return &Handler{mdb}
}


func (h *Handler) RegisterCuratorRoutes(router *mux.Router){
	router.HandleFunc("/login", h.login).Methods("POST")
	router.HandleFunc("/submitMusic", h.submitMusic).Methods("POST")
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request){
	// provide credentials to gain an authentication token
}

func (h *Handler) submitMusic(w http.ResponseWriter, r *http.Request){
	// submit music to be chosen to the data base.
	var submitSong types.SubmitSong
	err := utils.DecodeJSONBody(w, r, &submitSong)
	if err != nil{
		log.Error().Msg(err.Error())
		return
	}
	h.musicDB.InsertNewSong(&submitSong)
}

