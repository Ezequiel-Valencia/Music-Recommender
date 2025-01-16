package main

import (
	"database/sql"
	api "music-recommender/api"
	"music-recommender/db"
	"music-recommender/utils"
	"net/http"

	"github.com/rs/zerolog/log"
)

func main() {
	log.Info().Msg("Starting Server")

	abstractDB, dbPointer, _ := db.CreateDB(false)

	// Task set in separate thread that runs once a day
	go DailyTaskSet(dbPointer)

	var server *http.Server = api.CreateMainServer(dbPointer, abstractDB) //Pointer to the API server struct
	if err := server.ListenAndServe(); err != nil {
		log.Fatal().AnErr("error", err).Msg("Server can't start.")
	}
	log.Info().Msg("Server has stopped.")
}

func DailyTaskSet(db *sql.DB) {
	for {
		utils.SleepUntilXHour(3)
		log.Info().Msg("Executing Daily Server Maintenance Tasks")
		dbCleanUp(db)
		if isThereMoreThanOneOwner(db) {
			log.Fatal().Msg("There seems to be more than one owner. Terminating server.")
		}
	}
}


