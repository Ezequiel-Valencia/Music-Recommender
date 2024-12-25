package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)



type Config struct{
	HostAndPort string
	TimeFormat	string

	DBPort		int
	DBHost		string
	DBName 		string
	DBUser		string
	DBPasswd 	string

	SessionCookieName	string
}

var Envs = initConfig()

func initConfig() Config{
	return Config{
		HostAndPort: getEnv("PUBLIC_HOST", "localhost", true) + ":" + getEnv("PUBLIC_PORT", "8080", true),
		TimeFormat: time.RFC3339,
		DBPort: getEnvInt("DB_PORT", 5432, true),
		DBHost: getEnv("DB_HOST", "localhost", true),
		DBName: getEnv("DB_NAME", "postgres", true),
		DBUser: getEnv("DB_USER", "postgres", true),
		DBPasswd: getEnv("DB_PASSWD", "passwd", true),
		SessionCookieName: "session_token",
	}
}

func getEnvInt(key string, def int, allowForDefault bool) int {
	if value, ok := os.LookupEnv(key); ok {
		i, _ := strconv.Atoi(value)
		return i
	}
	if (allowForDefault){
		return def
	}
	log.Error().Msg(fmt.Sprintf("Unable to retrieve evn key %s", key))
	return -1
}

func getEnv(key string, def string, allowForDefault bool) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	if (allowForDefault){
		return def
	}
	log.Error().Msg(fmt.Sprintf("Unable to retrieve evn key %s", key))
	return ""
}
