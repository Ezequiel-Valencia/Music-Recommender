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
		creation_date, user_role, user_privileges) 
	VALUES($1, $2, $3, $4, $5, $6, $7, $8)`
	nowTime := time.Now().UTC().Format(config.StaticEnvs.TimeFormat)
	_, err := at.db.Exec(executeString, username, email, hashedPassword, subject_identifier,
		creation_source, nowTime, auth_types.VoterRole.String(), auth_types.NoPrivileges.String())
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

func (mdb AuthTable) GetUserStructFromUsername(providedUsername string) auth_types.User {
	var user_id int
	var username, email, creation_source, creation_date, user_role, user_privileges string
	err := mdb.db.QueryRow(`SELECT user_id, username, email, creation_source, creation_date, user_role, user_privileges 
	FROM users WHERE username = $1`, providedUsername).Scan(&user_id,
		&username, &email, &creation_source, &creation_date, &user_role, &user_privileges)
	if err == sql.ErrNoRows || err != nil {
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

func (mdb AuthTable) GetUserStructFromEmail(providedEmail string) auth_types.User {
	var user_id int
	var username, email, creation_source, creation_date, user_role, user_privileges string
	err := mdb.db.QueryRow(`SELECT user_id, username, email, creation_source, creation_date, user_role, user_privileges 
	FROM users WHERE email = $1`, providedEmail).Scan(&user_id,
		&username, &email, &creation_source, &creation_date, &user_role, &user_privileges)
	if err == sql.ErrNoRows || err != nil {
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

func (mdb AuthTable) SetUserRole(username string, userRole auth_types.UserRoles) {
	log.Warn().Msgf("Setting user %s role to %s", username, userRole.String())
	_, err := mdb.db.Exec("UPDATE users SET user_role = $1 WHERE username = $2", userRole.String(), username)
	if err != nil{
		log.Err(err).Msg("Can't set user role")
	}
}

func (mdb AuthTable) SetUserPrivilege(username string, userPrivilege auth_types.UserPrivileges) {
	log.Warn().Msgf("Setting user %s to privilege %s.", username, userPrivilege.String())
	_, err := mdb.db.Exec("UPDATE users SET user_privileges = $1 WHERE username = $2", userPrivilege.String(), username)
	if err != nil{
		log.Err(err).Msg("Can't set user role")
	}
}

