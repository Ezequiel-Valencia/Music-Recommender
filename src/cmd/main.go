package main

import (
	api "music-recommender/api"
	"music-recommender/db"
	"net/http"

	"github.com/rs/zerolog/log"
)

func main() {
	log.Info().Msg("Starting Server")

	abstractDB, dbPointer, _ := db.CreateDB(false)

	// Task set in separate thread that runs once a day
	go dailyTaskSet(dbPointer)

	var server *http.Server = api.CreateMainServer(dbPointer, abstractDB) //Pointer to the API server struct
	if err := server.ListenAndServe(); err != nil {
		log.Fatal().AnErr("error", err).Msg("Server can't start.")
	}
	log.Info().Msg("Server has stopped.")
}
