package t_utils

import (
	"database/sql"
	"log"
	"math/rand"
	"music-recommender/config"
	"music-recommender/db"
	"strconv"

	"github.com/ory/dockertest"
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

func TearDownTestDB() {
	if (dbPointer != nil && adb != nil && pool != nil && resource != nil){
		dbPointer.Close()
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
	db.CreateTables(dbPointer, true)
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
