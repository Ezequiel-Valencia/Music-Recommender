package main

import (
	"database/sql"
	"time"

	"github.com/rs/zerolog/log"
)

func dailyDBTasks(db *sql.DB){
	res, err := db.Exec("DELETE FROM sessions WHERE creation_date < CURRENT_DATE - interval '150 days'")
	resNum, _ := res.RowsAffected()
	if err != nil{
		log.Err(err).Msg("Problem deleting old sessions")
	}
	if resNum > 0{
		log.Info().Msgf("%d number of sessions removed from DB", resNum)
	}
}

func dailyTaskSet(db *sql.DB){
	for {
		log.Info().Msg("Executing Daily Server Maintenance Tasks")
		dailyDBTasks(db)
		time.Sleep(24 * time.Hour)
	}
}
