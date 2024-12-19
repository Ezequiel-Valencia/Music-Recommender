package api

import (
	"music-recommender/services/daily-user"
	"net/http"
	"music-recommender/services/curator"
	"music-recommender/db"
	"github.com/gorilla/mux"
)

type APIServer struct {
	db   *db.MusicDB// Pointer to SQL DB
	addr string
}

func CreateMainServer(addr string, db *db.MusicDB) *APIServer {
	return &APIServer{
		db:   db,
		addr: addr,
	}
}

func (a APIServer) Run() error {
	router := mux.NewRouter()
	subrouter := router.PathPrefix("/api/v1").Subrouter()
	
	anonymous_user_handler := daily_user.NewHandler(a.db)
	anonymous_user_handler.RegisterAnonymousUserRoutes(subrouter)

	curator_handler := music_curator.NewHandler(a.db)
	curator_handler.RegisterCuratorRoutes(subrouter)

	return http.ListenAndServe(a.addr, subrouter)
}
