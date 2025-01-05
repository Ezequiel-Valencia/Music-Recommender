package api

import (
	"database/sql"
	"music-recommender/config"
	"music-recommender/db"
	"music-recommender/db/auth_table"
	"music-recommender/db/music_table"
	"music-recommender/db/ranking_table"
	"music-recommender/services/auth"
	"music-recommender/services/curator_service"
	daily_user "music-recommender/services/daily-user"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type APIServer struct {
	Server *http.Server
}

func CreateMainServer(db *sql.DB, abd *db.AbstractDB) *http.Server {
	router := mux.NewRouter()
	subrouter := router.PathPrefix(config.StaticEnvs.APIPrefix).Subrouter()
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowCredentials: true, // Allow cookies from other origins to be sent
	})

	anonymous_user_handler := daily_user.NewHandler(ranking_table.CreateRankingTableDriver(db))
	anonymous_user_handler.RegisterAnonymousUserRoutes(subrouter)

	curator_handler := curator_service.NewHandler(music_table.CreateMusicTableDriver(db, abd))
	curator_handler.RegisterCuratorRoutes(subrouter)

	auth_handler := auth.NewHandler(auth_table.CreateAuthTableDriver(db, abd))
	auth_handler.RegisterAuthRoutes(subrouter)

	
	handlerWithCors := c.Handler(subrouter)
	return &http.Server{Addr: config.DynamicEnvs.HostAndPort, Handler: handlerWithCors}
}
