package daily_user

import (
	"music-recommender/db"
	"music-recommender/db/ranking_table"
	"music-recommender/services/auth"
	"music-recommender/types/communication_types"
	"music-recommender/utils"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

type Handler struct{
	rankingTable *ranking_table.RankingTable
	adb *db.AbstractDB
}


func NewHandler(mdb *ranking_table.RankingTable, adb *db.AbstractDB) *Handler{
	return &Handler{mdb, adb}
}


func (h *Handler) RegisterAnonymousUserRoutes(router *mux.Router){
	router.HandleFunc("/todaysMusic", h.handleGettingTodaysMusic).Methods("GET")
	router.HandleFunc("/calendar", auth.RequireAuth(h.handleGettingCalendar, h.adb)).Methods("GET")
	router.HandleFunc("/todaysMusic", auth.RequireAuth(h.submitAVote, h.adb)).Methods("POST")
}


// https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body

func (h *Handler) submitAVote(w http.ResponseWriter, r *http.Request, user db.User){
	var vote communication_types.SubmitVotePayload
	err := utils.DecodeJSONBody(w, r, vote)
	if err != nil{
		log.Error().Msg(err.Error())
		return
	}

	if (h.rankingTable.UserAlreadyVoteToday(user)){
		http.Error(w, "User already voted for today.", http.StatusBadRequest)
		return
	}
	
	if (vote.SongNumber > 2 || vote.SongNumber < 0){
		http.Error(w, "Bad song choice.", http.StatusNotAcceptable)
		return
	}

	h.rankingTable.UpdateTodaysRanking(vote, user)
	var todaysRanking *communication_types.TodaysRankingPayload = h.rankingTable.GetTodaysRanking()
	utils.WriteJSON(w, todaysRanking, 200)
}

func (h *Handler) handleGettingTodaysMusic(w http.ResponseWriter, r *http.Request){
	// get todays music from the DB and return the information
	todaysMusic := h.rankingTable.GetTodaysMusic()
	utils.WriteJSON(w, todaysMusic, 200)
}

func (h *Handler) handleGettingCalendar(w http.ResponseWriter, r *http.Request, user db.User){
	// get past music choices with their dates
	calendar := h.rankingTable.GetCalendarsMusic()
	utils.WriteJSON(w, calendar, 200)
}


