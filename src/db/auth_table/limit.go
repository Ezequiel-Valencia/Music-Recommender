package auth_table

import (
	"music-recommender/config"
	"music-recommender/types/internal_types/auth_types"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

func (at AuthTable) ReachedMaxNumberOfSessionsForUser(user auth_types.User) bool {
	res, err := at.db.Exec("SELECT session_id FROM sessions WHERE user_id = $1", user.UserId)
	numRes, _ := res.RowsAffected()
	if err != nil || numRes >= 10 {
		return true
	}
	return false
}

func (at AuthTable) IsTheUsernameAndEmailUnique(username string, email string) bool {
	res, err := at.db.Exec("SELECT * FROM users WHERE username = $1 OR email = $2", username, email)
	numRes, _ := res.RowsAffected()
	if err != nil || numRes != 0 {
		return false
	}
	return true
}

func (at AuthTable) SetAbilityForUserCreation(allowCreation bool) {
	log.Warn().Msgf("Changing state of user creation to %t", allowCreation)
	config.DynamicEnvs.AllowUserCreation = allowCreation
	res, _ := at.db.Exec("SELECT * FROM server_state")
	resNum, _ := res.RowsAffected()
	if resNum > 1 {
		log.Fatal().Msg("There is more than one row for server state.")
	} else if resNum == 1 {
		at.db.Exec("UPDATE server_state SET allow_user_creation = $1, update_date = $2", allowCreation, time.Now().UTC().Format(config.StaticEnvs.TimeFormat))
	}
}

func (at AuthTable) CorrectEmailAndPassword(email string, password string) bool {
	var dbHashedPassword string
	err := at.db.QueryRow("SELECT password_hash FROM users WHERE email = $1", email).Scan(&dbHashedPassword)
	if err != nil {
		return false
	}
	err = bcrypt.CompareHashAndPassword([]byte(dbHashedPassword), []byte(password))
	return err == nil
}
