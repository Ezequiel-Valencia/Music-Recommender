package main

import (
	"database/sql"
	"music-recommender/db"
	"time"

	"github.com/rs/zerolog/log"
)

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

func isThereMoreThanOneOwner(dbPointer *sql.DB) bool{
	res, err := dbPointer.Exec("SELECT * FROM users WHERE user_privileges = $1 OR user_role = $2", db.OwnerPrivileges, db.UnlimitedRole)
	resNum, _ := res.RowsAffected()
	if err != nil{
		log.Err(err).Msg("Error occurred checking if there is more than one Owner")
	} else if resNum > 1{
		log.Error().Msg("Daily checker found more than one owner.")
	}
	
	return resNum > 1 || err != nil
}

func dailyTaskSet(db *sql.DB) {
	for {
		log.Info().Msg("Executing Daily Server Maintenance Tasks")
		dbCleanUp(db)
		if (isThereMoreThanOneOwner(db)){log.Fatal().Msg("There seems to be more than one owner. Terminating server.")}
		time.Sleep(24 * time.Hour)
	}
}
