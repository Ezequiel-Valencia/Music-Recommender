package api

import (
	"database/sql"
	"music-recommender/db"
	"music-recommender/db/auth_table"
	"music-recommender/db/music_table"
	"music-recommender/db/ranking_table"
	"music-recommender/services/auth"
	"music-recommender/services/curator_service"
	daily_user "music-recommender/services/daily-user"
	"net/http"

	"github.com/gorilla/mux"
)

type APIServer struct {
	db   *sql.DB // Pointer to SQL DB
	abd  *db.AbstractDB
	addr string
}

func CreateMainServer(addr string, db *sql.DB, abd *db.AbstractDB) *APIServer {
	return &APIServer{
		db:   db,
		addr: addr,
		abd:  abd,
	}
}

func (a APIServer) Run() error {
	router := mux.NewRouter()
	subrouter := router.PathPrefix("/api/v1").Subrouter()

	anonymous_user_handler := daily_user.NewHandler(ranking_table.CreateRankingTableDriver(a.db))
	anonymous_user_handler.RegisterAnonymousUserRoutes(subrouter)

	curator_handler := curator_service.NewHandler(music_table.CreateMusicTableDriver(a.db, a.abd))
	curator_handler.RegisterCuratorRoutes(subrouter)

	auth_handler := auth.NewHandler(auth_table.CreateAuthTableDriver(a.db, a.abd))
	auth_handler.RegisterAuthRoutes(subrouter)

	return http.ListenAndServe(a.addr, subrouter)
}
