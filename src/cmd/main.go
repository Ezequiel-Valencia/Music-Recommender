package main

import (
	"database/sql"
	api "music-recommender/api"
	"music-recommender/db"
	"music-recommender/db/ranking_table"
	"music-recommender/types/internal_types/auth_types"
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
		utils.SleepUntilXHour(0)
		log.Info().Msg("Executing Daily Server Maintenance Tasks")
		dbCleanUp(db)
		if isThereMoreThanOneOwner(db) {
			log.Fatal().Msg("There seems to be more than one owner. Terminating server.")
		}
		countVoteAndPlaceNewSongs(db)
	}
}

func dbCleanUp(db *sql.DB) {
	res, err := db.Exec("DELETE FROM sessions WHERE creation_date < CURRENT_DATE - interval '150 days'")
	resNum, _ := res.RowsAffected()
	if err != nil {
		log.Err(err).Msg("Problem deleting old sessions")
	}
	if resNum > 0 {
		log.Info().Msgf("%d number of sessions removed from DB", resNum)
	}
}

func isThereMoreThanOneOwner(dbPointer *sql.DB) bool {
	res, err := dbPointer.Exec("SELECT * FROM userPrivileges WHERE moderator = $1 OR music_submission = $2", auth_types.OwnerPrivileges.String(), auth_types.UnlimitedRole.String())
	resNum, _ := res.RowsAffected()
	if err != nil {
		log.Err(err).Msg("Error occurred checking if there is more than one Owner")
	} else if resNum > 1 {
		log.Error().Msg("Daily checker found more than one owner.")
	}

	return resNum > 1 || err != nil
}

func countVoteAndPlaceNewSongs(dbPointer *sql.DB){
	// Get the rankings
	todaysRankingDriver := ranking_table.CreateTodaysRankingDriver(dbPointer)
	if (!todaysRankingDriver.AnySongsToBeRanked()){
		log.Info().Msg("No new songs to select from.")
		return
	}
	rankedSongs, topSong, err := todaysRankingDriver.CalculateTodaysRank()
	if (err != nil){
		return
	}

	if (topSong != -1){
		// Insert them into ranked table
		ranking_table.CreateRankedDriver(dbPointer).InsertAlreadyRankedSongs(topSong, rankedSongs)
		todaysRankingDriver.CleanTodaysRanking()
	} else {
		log.Warn().Msg("No songs where available to calculate rank.")
	}
	
	// Insert new songs into todaysRanking, inefficient and random for now
	todaysRankingDriver.SelectNewSongs()
}


