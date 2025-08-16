package db

import (
	"database/sql"
	"errors"
	"fmt"
	"music-recommender/config"
	"music-recommender/types/internal_types/auth_types"
	"time"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3" // Used to import the side effects of a package. Allows for SQlit3 Driver to be known
	"github.com/rs/zerolog/log"
)

type AbstractDB struct {
	db *sql.DB
}

func CreateDB(testMode bool) (*AbstractDB, *sql.DB, error) {
	var disableSSL string = " sslmode=disable"
	if (config.DynamicEnvs.DBSsl){
		disableSSL = ""
	}
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s%s",
		config.DynamicEnvs.DBHost, config.DynamicEnvs.DBPort, config.DynamicEnvs.DBUser, config.DynamicEnvs.DBPasswd, config.DynamicEnvs.DBName,
	disableSSL)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		if testMode {
			return nil, nil, err
		}
		log.Fatal().Msg(err.Error())
	}
	err = db.Ping()
	if err != nil {
		if testMode {
			return nil, nil, err
		}
		log.Fatal().Err(err).Msg("Can't connect to db.")
	}
	err = CreateTablesAndFunctions(db, testMode)
	initializeOrGetServerState(db)
	if err != nil {
		return nil, nil, err
	}
	return &AbstractDB{db}, db, nil
}

// The session and CSRF tokens should be decoded
func (abd AbstractDB) GetUserFromSessionID(sessionCookie string, csrfToken string, requiresCSRF bool) (auth_types.User, error) {
	if sessionCookie == "" {
		log.Error().Msg("Invalid Signature for Cookie")
		return auth_types.User{}, errors.New("invalid session cookie")
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
		return auth_types.User{}, err
	}

	var user_id int
	var username, email, creation_source, creation_date, user_role, user_privileges string
	err = abd.db.QueryRow(`SELECT users.user_id, users.username, users.email, 
	users.creation_source, users.creation_date, priv.music_submission, priv.moderator 
	FROM users 
	INNER JOIN userPrivileges priv ON priv.user_id = users.user_id
	WHERE users.user_id = $1 AND priv.user_id = $1`, userID).Scan(&user_id,
		&username, &email, &creation_source, &creation_date, &user_role, &user_privileges)
	if err == sql.ErrNoRows || err != nil {
		return auth_types.User{}, err
	}
	time, _ := time.Parse(config.StaticEnvs.TimeFormat, creation_date)
	return auth_types.User{UserId: user_id, Username: username, Email: email,
		CreationSource: auth_types.StringToUserCreationSource(creation_source),
		CreationDate:   time,
		UserRole:       auth_types.StringToUserRoles(user_role),
		UserPrivileges: auth_types.StringToUserPrivileges(user_privileges),
	}, nil
}
