package db

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type UserRoles int
const (
	VoterRole UserRoles = iota // 0
	OneSubmissionRole // 1
	CuratorRole // 5
	TrustedCuratorRole // 15
	UnlimitedRole // unlimited
)
var userRolesMap = map[int]string {
	VoterRole.EnumIndex() : "Voter",
	OneSubmissionRole.EnumIndex() : "OneSubmission",
	CuratorRole.EnumIndex() : "Curator",
	TrustedCuratorRole.EnumIndex() : "TrustedCurator",
	UnlimitedRole.EnumIndex() : "Unlimited",
}

func (ur UserRoles) String() string {
	return userRolesMap[ur.EnumIndex()]
}
func (ur UserRoles) EnumIndex() int {
	return int(ur)
}
func stringToUserRoles(s string) UserRoles{
	switch s{
	case userRolesMap[VoterRole.EnumIndex()]:
		return VoterRole
	default:
		return VoterRole
	}
}

type UserPrivileges int
const (
	NoPrivileges UserPrivileges = iota
	ModeratorPrivileges
	AdminPrivileges
	OwnerPrivileges
)

var userPrivilegesMap = map[int]string{
	NoPrivileges.EnumIndex() : "None",
	ModeratorPrivileges.EnumIndex() : "Moderator",
	AdminPrivileges.EnumIndex() : "Admin",
	OwnerPrivileges.EnumIndex() : "Owner",
}

func stringToUserPrivileges(s string) UserPrivileges{
	switch s{
	case userPrivilegesMap[NoPrivileges.EnumIndex()]:
		return NoPrivileges
	default:
		return NoPrivileges
	}
}

func (up UserPrivileges) String() string {
	return userPrivilegesMap[up.EnumIndex()]
}
func (up UserPrivileges) EnumIndex() int {
	return int(up)
}

type UserCreationSource int
var userCreationSourceMap = map[int] string{
	LocalUserCreationSource.EnumIndex() : "Local",
}
const (
	LocalUserCreationSource UserCreationSource = iota
)
func stringToUserCreationSource(s string) UserCreationSource{
	switch s{
	default:
		return LocalUserCreationSource
	}
}

func (up UserCreationSource) String() string {
	return userCreationSourceMap[up.EnumIndex()]
}
func (up UserCreationSource) EnumIndex() int {
	return int(up)
}







// https://builtin.com/software-engineering-perspectives/golang-enum
type UserRolesStruct struct {
	Owner string
}

func (mdb MusicDB) CreateUser(username string, hashedPassword string, subject_identifier string, creation_source string) error {
	const executeString = `INSERT INTO users(username, passwd_hash, subject_identifier, creation_source, 
		creation_date, user_role, user_privileges) 
	VALUES(?, ?, ?, ?, ?, ?, ?)`
	_, err := mdb.db.Exec(executeString, username, hashedPassword, subject_identifier,
		creation_source, time.Now().UTC().Format(time.RFC3339), VoterRole.String(), NoPrivileges.String())
	handleMusicDBError(err)
	if err != nil{return err}
	return nil
}

func (mdb MusicDB) DeleteUser() {
}

func (mdb MusicDB) GetUserStructFromUsername(providedUsername string) User {
	var user_id int
	var username, creation_source, creation_date, user_role, user_privileges string
	err := mdb.db.QueryRow(`SELECT user_id, username, creation_source, creation_date, user_role, user_privileges 
	FROM users WHERE username = ?`, providedUsername).Scan(&user_id, 
		&username, &creation_source,  &creation_date, &user_role, &user_privileges)
	if err == sql.ErrNoRows{
		return User{}
	}
	time, _ := time.Parse(time.RFC3339, creation_date)
	return User{UserId: user_id, Username: username, 
		CreationSource: stringToUserCreationSource(creation_source),
		CreationDate: time,
		UserRole: stringToUserRoles(user_role),
		UserPrivileges: stringToUserPrivileges(user_privileges),
	}
}

func (mdb MusicDB) ValidUsernameAndCredentials(username string, hashedPassword string) bool {
	dbHashedPassword := ""
	err := bcrypt.CompareHashAndPassword([]byte(dbHashedPassword), []byte(hashedPassword))
	return err == nil
}

func (mdb MusicDB) GenerateAndStoreSessionID(user User) (string, error) {
	newToken := GenerateSecureToken(32)
	const executeString = `INSERT INTO sessions(user_id, session_id) 
	VALUES(?, ?)`
	_, err := mdb.db.Exec(executeString, user.UserId, newToken)
	handleMusicDBError(err)
	if (err != nil){return "", err}

	return newToken, nil
}

func (mdb MusicDB) SessionTokenIsValid(username string) (bool, error) {
	return false, nil
}

func (mdb MusicDB) RemoveSessionTokens(username string, sessionid string) error {
	return nil
}

func GenerateSecureToken(length int) string {
	bytesArray := make([]byte, length)
	if _, err := rand.Read(bytesArray); err != nil {
		log.Fatal().Msgf("Failed to generate token: %v", err)
	}
	return base64.URLEncoding.EncodeToString(bytesArray) //Base 64 is a set of characters safe for HTTP traffic
}
