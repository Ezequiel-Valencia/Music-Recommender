package auth_table

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"music-recommender/db"

	"github.com/rs/zerolog/log"
)

type AuthTable struct {
	db         *sql.DB
	AbstractDB *db.AbstractDB
}

func CreateAuthTableDriver(db *sql.DB, abd *db.AbstractDB) *AuthTable {
	return &AuthTable{db: db, AbstractDB: abd}
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

func (at AuthTable) UpdatePassword(user db.User, hashedPassword string) {
	_, err := at.db.Exec("UPDATE users SET password_hash = $1 WHERE user_id = $2", hashedPassword, user.UserId)
	if err != nil {
		log.Err(err).Msg("DB Error: Update password")
	}
}

func GenerateSecureToken(length int) string {
	bytesArray := make([]byte, length)
	if _, err := rand.Read(bytesArray); err != nil {
		log.Fatal().Msgf("Failed to generate token: %v", err)
	}
	return base64.URLEncoding.EncodeToString(bytesArray) //Base 64 is a set of characters safe for HTTP traffic
}

