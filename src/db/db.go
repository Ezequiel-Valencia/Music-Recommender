package db

import (
	"database/sql"
	"errors"
	"fmt"
	"music-recommender/config"
	"time"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3" // Used to import the side effects of a package. Allows for SQlit3 Driver to be known
	"github.com/rs/zerolog/log"
)

type AbstractDB struct {
	db *sql.DB
}

func CreateDB(testMode bool) (*AbstractDB, *sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config.DynamicEnvs.DBHost, config.DynamicEnvs.DBPort, config.DynamicEnvs.DBUser, config.DynamicEnvs.DBPasswd, config.DynamicEnvs.DBName)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		if testMode {
			return nil, nil, err
		}
		log.Fatal().Msg(err.Error())
	}
	err = db.Ping()
	if (err != nil){
		if testMode{
			return nil, nil, err
		}
		log.Fatal().Err(err).Msg("Can't connect to db.")
	}
	err = CreateTables(db, testMode)
	initializeOrGetServerState(db)
	if err != nil {
		return nil, nil, err
	}
	return &AbstractDB{db}, db, nil
}

// The session and CSRF tokens should be decoded
func (abd AbstractDB) GetUserFromSessionID(sessionCookie string, csrfToken string, requiresCSRF bool) (User, error) {
	if sessionCookie == "" {
		log.Error().Msg("Invalid Signature for Cookie")
		return User{}, errors.New("invalid session cookie")
	}
	var userID int
	var err error
	if requiresCSRF {
		err = abd.db.QueryRow("SELECT user_id FROM sessions WHERE session_id = $1 AND csrf_token = $2", sessionCookie, csrfToken).Scan(&userID)
	} else {
		err = abd.db.QueryRow("SELECT user_id FROM sessions WHERE session_id = $1", sessionCookie).Scan(&userID)
	}
	if err != nil || err == sql.ErrNoRows {
		log.Error().Msg("Can't retrieve user ID from session.")
		return User{}, err
	}

	var user_id int
	var username, creation_source, creation_date, user_role, user_privileges string
	err = abd.db.QueryRow(`SELECT user_id, username, creation_source, creation_date, user_role, user_privileges 
	FROM users WHERE user_id = $1`, userID).Scan(&user_id,
		&username, &creation_source, &creation_date, &user_role, &user_privileges)
	if err == sql.ErrNoRows || err != nil {
		return User{}, err
	}
	time, _ := time.Parse(config.StaticEnvs.TimeFormat, creation_date)
	return User{UserId: user_id, Username: username,
		CreationSource: StringToUserCreationSource(creation_source),
		CreationDate:   time,
		UserRole:       StringToUserRoles(user_role),
		UserPrivileges: StringToUserPrivileges(user_privileges),
	}, nil
}
