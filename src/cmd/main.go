package main

import (
	"fmt"
	api "music-recommender/api"
	"music-recommender/db"

	"github.com/rs/zerolog/log"
)

func main() {
	fmt.Println("Hello World!")
	log.Info().Msg("Logging Info")

	dbPointer := db.CreateSQLiteStorage()

	var server *api.APIServer = api.CreateMainServer(":8080", dbPointer) //Pointer to the API server struct
	if err := server.Run(); err != nil {
		log.Fatal().AnErr("error", err)
	}

}
