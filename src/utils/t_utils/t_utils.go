package t_utils

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"log"
	"math/rand"
	"music-recommender/config"
	"music-recommender/db"
	"music-recommender/db/music_table"
	"music-recommender/types/communication_types"
	"music-recommender/types/internal_types/auth_types"
	"net/url"
	"strconv"
	"time"

	"github.com/ory/dockertest"
	"golang.org/x/crypto/bcrypt"
)

var dbPointer *sql.DB = nil
var adb *db.AbstractDB = nil
var pool *dockertest.Pool = nil
var resource *dockertest.Resource = nil

// Creates a Postgres DB using docker. And if the container is already up returns pointers to the DB.
func GetTestDB() (*db.AbstractDB, *sql.DB) {
	if dbPointer != nil && adb != nil && pool != nil && resource != nil {
		return adb, dbPointer
	}
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	var err error
	pool, err = dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	// uses pool to try to connect to Docker
	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err = pool.Run("postgres", "", []string{"POSTGRES_PASSWORD=passwd"})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	config.DynamicEnvs.DBPort, _ = strconv.Atoi(resource.GetPort("5432/tcp"))
	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		adb, dbPointer, err = db.CreateDB(true)
		if err != nil {
			return err
		}
		return dbPointer.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}
	return adb, dbPointer
}

// Completely destroys everything regarding the containers and DB
func TearDownTestDB() {
	if dbPointer != nil && adb != nil && pool != nil && resource != nil {
		if err := dbPointer.Close(); err != nil {
			log.Fatal(err)
		}
		adb = nil
		dbPointer = nil
		if err := pool.Purge(resource); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
		}
		pool = nil
		resource = nil
	}
}

func ResetTestDB() {
	_, err := dbPointer.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
	if err != nil {
		log.Fatalf("Could not reset test DB: %s", err)
	}
	if err := db.CreateTablesAndFunctions(dbPointer, true); err != nil {
		log.Fatal(err)
	}
}

// If not Alpha-Numeric compliant, UTF-32 characters are generated.
func GenerateRandomRuneString(lenOfRunes int, alphaNumericCompliant bool) string {
	const alphaNumerics = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	b := make([]rune, lenOfRunes)

	for i := range b {
		var randRune rune
		if alphaNumericCompliant {
			randRune = rune(alphaNumerics[rand.Intn(len(alphaNumerics))])
		} else {
			randRune = rune(rand.Intn(0x10FFF + 1))
		}
		b[i] = randRune
	}
	return string(b)
}

// Inserts both a fake user, and user privileges associated with them.
func CreateFakeUser(db *sql.DB, user *auth_types.User, nonHashedPasswd string) {
	const executeString = `INSERT INTO users(username, email, password_hash, subject_identifier, creation_source, 
		creation_date) 
	VALUES($1, $2, $3, $4, $5, $6) RETURNING user_id`

	var userID int
	bytes, _ := bcrypt.GenerateFromPassword([]byte(nonHashedPasswd), 14)
	hashedPassword := string(bytes)
	db.QueryRow(executeString, user.Username, user.Email, 
		hashedPassword, "", user.CreationSource, 
		user.CreationDate.UTC().Format(config.StaticEnvs.TimeFormat)).Scan(&userID)
	_, err := db.Exec(`INSERT INTO userPrivileges(user_id, moderator, music_submission)
	VALUES($1, $2, $3)`, userID, user.UserPrivileges.String(), user.UserRole.String())
	if err != nil {
		log.Fatal(err)
	}
}

func CreateHTTPBodyURLEncoded(body string) io.Reader {
	b64 := url.PathEscape(body)
	return bytes.NewBufferString(b64)
}

func FillDBWithFakeSongsAndDescription(dbPointer *sql.DB, adb *db.AbstractDB, user *auth_types.User, fakeDescription string) {
	musicDriver := music_table.CreateMusicTableDriver(dbPointer, adb)
	for i := range 10 {
		submitSong := communication_types.SubmitSong{Name: fmt.Sprintf("Song %d", i),
			Artist: fmt.Sprintf("Artist %d", i), SongURL: fmt.Sprintf("https://youtu.be/MPANooz_b9Q%d", i),}
		musicDriver.InsertNewSong(&submitSong, *user)
	}
	_, err := dbPointer.Exec(`INSERT INTO submissionDescriptions(description) VALUES($1)`, fakeDescription)
	if err != nil{
		log.Print(err)
	}
}

var TestUserBob auth_types.User = auth_types.User{Username: "Bob", UserId: 1,
	Email: "bob@gmail.com", CreationSource: auth_types.LocalUserCreationSource,
	CreationDate: time.Now(), UserRole: auth_types.VoterRole, UserPrivileges: auth_types.NoPrivileges}

var TestUserCuratorModerator auth_types.User = auth_types.User{Username: "Admin", UserId: 2,
	Email: "admin@gmail.com", CreationSource: auth_types.LocalUserCreationSource,
	CreationDate: time.Now(), UserRole: auth_types.CuratorRole, UserPrivileges: auth_types.ModeratorPrivileges}

var TestUserOwner auth_types.User = auth_types.User{Username: "Owner", UserId: 3,
	Email: "owner@gmail.com", CreationSource: auth_types.LocalUserCreationSource,
	CreationDate: time.Now(), UserRole: auth_types.UnlimitedRole, UserPrivileges: auth_types.OwnerPrivileges}
