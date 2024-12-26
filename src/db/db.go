package db

import (
	"database/sql"
	"fmt"
	"music-recommender/config"
	"net/http"
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
		config.Envs.DBHost, config.Envs.DBPort, config.Envs.DBUser, config.Envs.DBPasswd, config.Envs.DBName)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		if testMode{return nil, nil, err}
		log.Fatal().Msg(err.Error())
	}
	err = createTables(db, testMode)
	if err != nil{return nil, nil, err}
	return &AbstractDB{db}, db, nil
}

func (abd AbstractDB) GetUserFromSessionID(r *http.Request) (User, error) {
	sessionCookie, err := r.Cookie(config.Envs.SessionCookieName)
	if err != nil {
		return User{}, err
	}

	var userID int
	err = abd.db.QueryRow("SELECT user_id FROM sessions WHERE session_id = $1", sessionCookie.Value).Scan(&userID)
	if err != nil {
		log.Err(err).Msg("Can't retrieve user ID from session.")
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
	time, _ := time.Parse(config.Envs.TimeFormat, creation_date)
	return User{UserId: user_id, Username: username,
		CreationSource: StringToUserCreationSource(creation_source),
		CreationDate:   time,
		UserRole:       StringToUserRoles(user_role),
		UserPrivileges: StringToUserPrivileges(user_privileges),
	}, nil
}
