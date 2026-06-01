package auth_table

import (
	"database/sql"
	"music-recommender/config"
	"music-recommender/types/internal_types/auth_types"
	"time"

	"github.com/rs/zerolog/log"
)

func (at AuthTable) CreateUser(username string, email string, hashedPassword string,
	subject_identifier string, creation_source string) error {

	const executeString = `INSERT INTO users(username, email, password_hash, subject_identifier, creation_source, 
		creation_date) 
	VALUES($1, $2, $3, $4, $5, $6) RETURNING user_id`
	var userID int
	nowTime := time.Now().UTC().Format(config.StaticEnvs.TimeFormat)
	if err := at.db.QueryRow(executeString, username, email, hashedPassword, subject_identifier,
		creation_source, nowTime).Scan(&userID); err != nil {
		log.Err(err).Msgf("DB Error: Can't get user ID for %s", username)
	}
	_, err := at.db.Exec(`INSERT INTO userPrivileges (user_id, moderator, music_submission)
	VALUES($1, $2, $3)`, userID, auth_types.VoterRole.String(), auth_types.NoPrivileges.String())

	if err != nil {
		log.Err(err).Msgf("DB Error: Can't create user %s", username)
	}

	return nil
}

func (at AuthTable) DeleteUser(username string, userID int) error {
	_, err := at.db.Exec("DELETE FROM users WHERE username = $1 AND user_id = $2", username, userID)
	return err
}

func (mdb AuthTable) GetUserStructFromUsername(providedUsername string) auth_types.User {
	return mdb.gettingUser("", providedUsername)
}

func (mdb AuthTable) GetUserStructFromEmail(providedEmail string) auth_types.User {
	return mdb.gettingUser(providedEmail, "")
}

func (mdb AuthTable) gettingUser(providedEmail string, providedUsername string) auth_types.User {
	var user_id int
	var username, email, creation_source, creation_date, user_role, user_privileges string
	var providedString string
	var queryString string

	if providedEmail == "" {
		providedString = providedUsername
		queryString = `SELECT users.user_id, users.username, users.email, 
		users.creation_source, users.creation_date, 
		priv.music_submission, priv.moderator
		FROM users 
		INNER JOIN userPrivileges priv ON users.user_id = priv.user_id
		WHERE users.username = $1`
	} else {
		providedString = providedEmail
		queryString = `SELECT users.user_id, users.username, users.email, 
		users.creation_source, users.creation_date, 
		priv.music_submission, priv.moderator
		FROM users 
		INNER JOIN userPrivileges priv ON users.user_id = priv.user_id
		WHERE users.email = $1`
	}

	err := mdb.db.QueryRow(queryString, providedString).
		Scan(&user_id, &username, &email, &creation_source, &creation_date, &user_role, &user_privileges)
	if err == sql.ErrNoRows || err != nil {
		log.Err(err).Msgf("Unable to get user from email/username: %s", providedString)
		return auth_types.User{}
	}

	time, _ := time.Parse(config.StaticEnvs.TimeFormat, creation_date)
	return auth_types.User{UserId: user_id, Username: username, Email: email,
		CreationSource: auth_types.StringToUserCreationSource(creation_source),
		CreationDate:   time,
		UserRole:       auth_types.StringToUserRoles(user_role),
		UserPrivileges: auth_types.StringToUserPrivileges(user_privileges),
	}
}

func (mdb AuthTable) SetUserRole(user auth_types.User, userRole auth_types.UserRoles) {
	log.Warn().Msgf("Setting user %s role to %s", user.Username, userRole.String())
	_, err := mdb.db.Exec("UPDATE userPrivileges SET music_submission = $1 WHERE user_id = $2", userRole.String(), user.UserId)
	if err != nil {
		log.Err(err).Msg("Can't set user role")
	}
}

func (mdb AuthTable) SetUserPrivilege(user auth_types.User, userPrivilege auth_types.UserPrivileges) {
	log.Warn().Msgf("Setting user %s to privilege %s.", user.Username, userPrivilege.String())
	_, err := mdb.db.Exec("UPDATE userPrivileges SET moderator = $1 WHERE user_id = $2", userPrivilege.String(), user.UserId)
	if err != nil {
		log.Err(err).Msg("Can't set user role")
	}
}
