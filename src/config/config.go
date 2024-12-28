package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/rs/zerolog/log"
)

type Config struct {
	HostAndPort string

	DBPort   int
	DBHost   string
	DBName   string
	DBUser   string
	DBPasswd string
}

type StaticConfig struct {
	TimeFormat  string
	SessionCookieName string
	CSRFCookieName	string
	CSRFHeaderName string
	APIPrefix	string
}

var DynamicEnvs = initConfig()

// Cookie Token should be 64 bytes, https://github.com/gorilla/securecookie
var SecureCookie *securecookie.SecureCookie;

var StaticEnvs = StaticConfig{
	TimeFormat: time.RFC3339,
	SessionCookieName: "session_token",
	APIPrefix: "/api/v1",
	CSRFCookieName: "csrf_token",
	CSRFHeaderName: "X-CSRF-Token",
}

func initConfig() Config {
	// Signed Cookies, and CSRF Setup
	cookie_token := getEnv("COOKIE_SIGNING_KEY", "insecure", false)
	SecureCookie = securecookie.New([]byte(cookie_token), nil)

	// Dynamic Config Setup
	return Config{
		HostAndPort:       getEnv("PUBLIC_HOST", "localhost", true) + ":" + getEnv("PUBLIC_PORT", "8080", true),
		DBPort:            getEnvInt("DB_PORT", 5432, true),
		DBHost:            getEnv("DB_HOST", "localhost", true),
		DBName:            getEnv("DB_NAME", "postgres", true),
		DBUser:            getEnv("DB_USER", "postgres", true),
		DBPasswd:          getEnv("DB_PASSWD", "passwd", true),
	}
}

func getEnvInt(key string, def int, allowForDefault bool) int {
	if value, ok := os.LookupEnv(key); ok {
		i, _ := strconv.Atoi(value)
		return i
	}
	if allowForDefault {
		return def
	}
	log.Error().Msg(fmt.Sprintf("Unable to retrieve evn key %s", key))
	return def
}

func getEnv(key string, def string, allowForDefault bool) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	if allowForDefault {
		return def
	}
	log.Error().Msg(fmt.Sprintf("Unable to retrieve evn key %s", key))
	return def
}
