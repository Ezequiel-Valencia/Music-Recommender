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

func (at AuthTable) CreateUser(username string, hashedPassword string,
	subject_identifier string, creation_source string) error {

	const executeString = `INSERT INTO users(username, password_hash, subject_identifier, creation_source, 
		creation_date, user_role, user_privileges) 
	VALUES($1, $2, $3, $4, $5, $6, $7)`
	_, err := at.db.Exec(executeString, username, hashedPassword, subject_identifier,
		creation_source, time.Now().UTC().Format(config.StaticEnvs.TimeFormat), db.VoterRole.String(), db.NoPrivileges.String())
	if err != nil {
		log.Err(err).Msg("DB Error")
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
	var username, creation_source, creation_date, user_role, user_privileges string
	err := mdb.db.QueryRow(`SELECT user_id, username, creation_source, creation_date, user_role, user_privileges 
	FROM users WHERE username = $1`, providedUsername).Scan(&user_id,
		&username, &creation_source, &creation_date, &user_role, &user_privileges)
	if err == sql.ErrNoRows || err != nil {
		return db.User{}
	}
	time, _ := time.Parse(config.StaticEnvs.TimeFormat, creation_date)
	return db.User{UserId: user_id, Username: username,
		CreationSource: db.StringToUserCreationSource(creation_source),
		CreationDate:   time,
		UserRole:       db.StringToUserRoles(user_role),
		UserPrivileges: db.StringToUserPrivileges(user_privileges),
	}
}

func (at AuthTable) CorrectUsernameAndPassword(username string, password string) bool {
	var dbHashedPassword string
	err := at.db.QueryRow("SELECT password_hash FROM users WHERE username = $1", username).Scan(&dbHashedPassword)
	if err != nil {
		return false
	}
	err = bcrypt.CompareHashAndPassword([]byte(dbHashedPassword), []byte(password))
	return err == nil
}

func (mdb AuthTable) GenerateAndStoreSessionID(user db.User) (string, error) {
	newToken := GenerateSecureToken(50)
	const executeString = `INSERT INTO sessions(user_id, session_id) 
	VALUES($1, $2)`
	_, err := mdb.db.Exec(executeString, user.UserId, newToken)
	if err != nil {
		log.Err(err).Msg("DB Error")
		return "", err
	}

	return newToken, nil
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
