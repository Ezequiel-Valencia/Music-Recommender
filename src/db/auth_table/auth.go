package auth_table

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"music-recommender/config"
	"music-recommender/db"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type AuthTable struct {
	db         *sql.DB
	AbstractDB *db.AbstractDB
}

func CreateAuthTableDriver(db *sql.DB, abd *db.AbstractDB) *AuthTable {
	return &AuthTable{db: db, AbstractDB: abd}
}

func (at AuthTable) CreateUser(username string, email string, hashedPassword string,
	subject_identifier string, creation_source string) error {

	const executeString = `INSERT INTO users(username, email, password_hash, subject_identifier, creation_source, 
		creation_date, user_role, user_privileges) 
	VALUES($1, $2, $3, $4, $5, $6, $7, $8)`
	nowTime := time.Now().UTC().Format(config.StaticEnvs.TimeFormat)
	_, err := at.db.Exec(executeString, username, email, hashedPassword, subject_identifier,
		creation_source, nowTime, db.VoterRole.String(), db.NoPrivileges.String())
	if err != nil {
		log.Err(err).Msg("DB Error: Create User")
		return err
	}
	return nil
}

func (at AuthTable) DeleteUser(username string, userID int) error {
	_, err := at.db.Exec("DELETE FROM users WHERE username = $1 AND user_id = $2", username, userID)
	return err
}

func (mdb AuthTable) GetUserStructFromUsername(providedUsername string) db.User {
	var user_id int
	var username, email, creation_source, creation_date, user_role, user_privileges string
	err := mdb.db.QueryRow(`SELECT user_id, username, email, creation_source, creation_date, user_role, user_privileges 
	FROM users WHERE username = $1`, providedUsername).Scan(&user_id,
		&username, &email, &creation_source, &creation_date, &user_role, &user_privileges)
	if err == sql.ErrNoRows || err != nil {
		return db.User{}
	}
	time, _ := time.Parse(config.StaticEnvs.TimeFormat, creation_date)
	return db.User{UserId: user_id, Username: username, Email: email,
		CreationSource: db.StringToUserCreationSource(creation_source),
		CreationDate:   time,
		UserRole:       db.StringToUserRoles(user_role),
		UserPrivileges: db.StringToUserPrivileges(user_privileges),
	}
}

func (mdb AuthTable) GetUserStructFromEmail(providedEmail string) db.User {
	var user_id int
	var username, email, creation_source, creation_date, user_role, user_privileges string
	err := mdb.db.QueryRow(`SELECT user_id, username, email, creation_source, creation_date, user_role, user_privileges 
	FROM users WHERE email = $1`, providedEmail).Scan(&user_id,
		&username, &email, &creation_source, &creation_date, &user_role, &user_privileges)
	if err == sql.ErrNoRows || err != nil {
		return db.User{}
	}
	time, _ := time.Parse(config.StaticEnvs.TimeFormat, creation_date)
	return db.User{UserId: user_id, Username: username, Email: email,
		CreationSource: db.StringToUserCreationSource(creation_source),
		CreationDate:   time,
		UserRole:       db.StringToUserRoles(user_role),
		UserPrivileges: db.StringToUserPrivileges(user_privileges),
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

func (mdb AuthTable) GenerateAndStoreSessionID(user db.User, timeCreated string) (string, string, error) {
	newToken := GenerateSecureToken(50)
	csrfToken := GenerateSecureToken(50)
	const executeString = `INSERT INTO sessions(user_id, session_id, csrf_token, creation_date) 
	VALUES($1, $2, $3, $4)`
	_, err := mdb.db.Exec(executeString, user.UserId, newToken, csrfToken, timeCreated)
	if err != nil {
		log.Err(err).Msg("DB Error: Generate and store session")
		return "", "", err
	}

	return newToken, csrfToken, nil
}

func (at AuthTable) RemoveSessionTokens(user db.User, sessionid string) error {
	res, err := at.db.Exec("DELETE FROM sessions WHERE user_id = $1 AND session_id = $2", user.UserId, sessionid)
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		errString := fmt.Sprintf("deleted 0 sessions for user %s", user.Username)
		log.Error().Msg(errString)
		return errors.New(errString)
	}
	return err
}

func GenerateSecureToken(length int) string {
	bytesArray := make([]byte, length)
	if _, err := rand.Read(bytesArray); err != nil {
		log.Fatal().Msgf("Failed to generate token: %v", err)
	}
	return base64.URLEncoding.EncodeToString(bytesArray) //Base 64 is a set of characters safe for HTTP traffic
}

func (at AuthTable) UpdatePassword(user db.User, hashedPassword string) {
	_, err := at.db.Exec("UPDATE users SET password_hash = $1 WHERE user_id = $2", hashedPassword, user.UserId)
	if err != nil {
		log.Err(err).Msg("DB Error: Update password")
	}
}

func (at AuthTable) ReachedMaxNumberOfSessionsForUser(user db.User) bool {
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
