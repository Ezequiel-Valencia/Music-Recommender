package main

import (
	api "music-recommender/api"
	"music-recommender/config"
	"music-recommender/db"

	"github.com/rs/zerolog/log"
)

func main() {
	log.Info().Msg("Starting Server")

	abstractDB, dbPointer := db.CreateSQLiteStorage()

	var server *api.APIServer = api.CreateMainServer(config.Envs.HostAndPort, dbPointer, abstractDB) //Pointer to the API server struct
	if err := server.Run(); err != nil {
		log.Fatal().AnErr("error", err)
	}

}
