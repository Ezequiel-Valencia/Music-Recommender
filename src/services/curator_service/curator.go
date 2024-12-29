package curator_service

import (
	"music-recommender/db"
	"music-recommender/db/music_table"
	"music-recommender/services/auth"
	"music-recommender/types"
	"music-recommender/utils"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

type Handler struct{
	musicDB *music_table.MusicTable
}


func NewHandler(mdb *music_table.MusicTable) *Handler{
	return &Handler{mdb}
}


func (h *Handler) RegisterCuratorRoutes(router *mux.Router){
	router.HandleFunc("/submitMusic", auth.RequireAuth(h.submitMusic, h.musicDB.AbstractDB)).Methods("POST")
}

func (h *Handler) submitMusic(w http.ResponseWriter, r *http.Request, user db.User){
	// submit music to be chosen to the data base.
	var submitSong types.SubmitSong
	err := utils.DecodeJSONBody(w, r, &submitSong)
	if err != nil{
		log.Error().Msg(err.Error())
		return
	}
	h.musicDB.InsertNewSong(&submitSong)
}

