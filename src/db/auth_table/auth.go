package auth_table

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"music-recommender/db"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type AuthTable struct {
	db *sql.DB
	AbstractDB *db.AbstractDB
}

func CreateAuthTableDriver(db *sql.DB, abd *db.AbstractDB) *AuthTable{
	return &AuthTable{db: db, AbstractDB: abd}
}

func (mdb AuthTable) CreateUser(username string, hashedPassword string, subject_identifier string, creation_source string) error {
	const executeString = `INSERT INTO users(username, passwd_hash, subject_identifier, creation_source, 
		creation_date, user_role, user_privileges) 
	VALUES(?, ?, ?, ?, ?, ?, ?)`
	_, err := mdb.db.Exec(executeString, username, hashedPassword, subject_identifier,
		creation_source, time.Now().UTC().Format(time.RFC3339), VoterRole.String(), NoPrivileges.String())
	if err != nil {
		return err
	}
	return nil
}

func (mdb AuthTable) DeleteUser() {
}

func (mdb AuthTable) GetUserStructFromUsername(providedUsername string) User {
	var user_id int
	var username, creation_source, creation_date, user_role, user_privileges string
	err := mdb.db.QueryRow(`SELECT user_id, username, creation_source, creation_date, user_role, user_privileges 
	FROM users WHERE username = ?`, providedUsername).Scan(&user_id,
		&username, &creation_source, &creation_date, &user_role, &user_privileges)
	if err == sql.ErrNoRows {
		return User{}
	}
	time, _ := time.Parse(time.RFC3339, creation_date)
	return User{UserId: user_id, Username: username,
		CreationSource: stringToUserCreationSource(creation_source),
		CreationDate:   time,
		UserRole:       stringToUserRoles(user_role),
		UserPrivileges: stringToUserPrivileges(user_privileges),
	}
}

func (mdb AuthTable) ValidUsernameAndCredentials(username string, hashedPassword string) bool {
	dbHashedPassword := ""
	err := bcrypt.CompareHashAndPassword([]byte(dbHashedPassword), []byte(hashedPassword))
	return err == nil
}

func (mdb AuthTable) GenerateAndStoreSessionID(user User) (string, error) {
	newToken := GenerateSecureToken(32)
	const executeString = `INSERT INTO sessions(user_id, session_id) 
	VALUES(?, ?)`
	_, err := mdb.db.Exec(executeString, user.UserId, newToken)
	if err != nil {
		return "", err
	}

	return newToken, nil
}



func (mdb AuthTable) RemoveSessionTokens(username string, sessionid string) error {
	return nil
}

func (mdb AuthTable) CreateNewCurator() error {
	return nil
}

func GenerateSecureToken(length int) string {
	bytesArray := make([]byte, length)
	if _, err := rand.Read(bytesArray); err != nil {
		log.Fatal().Msgf("Failed to generate token: %v", err)
	}
	return base64.URLEncoding.EncodeToString(bytesArray) //Base 64 is a set of characters safe for HTTP traffic
}
