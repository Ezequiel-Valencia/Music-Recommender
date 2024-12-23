package daily_user

import (
	"music-recommender/db/ranking_table"
	"music-recommender/types"
	"music-recommender/utils"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

type Handler struct{
	rankingTable *ranking_table.RankingTable
}


func NewHandler(mdb *ranking_table.RankingTable) *Handler{
	return &Handler{mdb}
}


func (h *Handler) RegisterAnonymousUserRoutes(router *mux.Router){
	router.HandleFunc("/todaysMusic", h.handleGettingTodaysMusic).Methods("GET")
	router.HandleFunc("/calendar", h.handleGettingCalendar).Methods("GET")
	router.HandleFunc("/todaysMusic", h.submitAVote).Methods("POST")
}


// https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body

func (h *Handler) submitAVote(w http.ResponseWriter, r *http.Request){
	var vote types.SubmitVotePayload
	err := utils.DecodeJSONBody(w, r, vote)
	if err != nil{
		log.Error().Msg(err.Error())
		return
	}
	h.rankingTable.UpdateTodaysRanking(vote)
	var todaysRanking *types.TodaysRankingPayload = h.rankingTable.GetTodaysRanking()
	utils.WriteJSON(w, todaysRanking, 200)
}

func (h *Handler) handleGettingTodaysMusic(w http.ResponseWriter, r *http.Request){
	// get todays music from the DB and return the information
	todaysMusic := h.rankingTable.GetTodaysMusic()
	utils.WriteJSON(w, todaysMusic, 200)
}

func (h *Handler) handleGettingCalendar(w http.ResponseWriter, r *http.Request){
	// get past music choices with their dates
	calendar := h.rankingTable.GetCalendarsMusic()
	utils.WriteJSON(w, calendar, 200)
}


