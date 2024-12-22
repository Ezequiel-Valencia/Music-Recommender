package auth

import (
	"net/http"

	"music-recommender/db"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)




func RequireAuth(handlerFunc http.HandlerFunc, mdb *db.MusicDB) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		isSessionValid, err := mdb.SessionTokenIsValid("username")
		if (err != nil || !isSessionValid) {
			if (err != nil){
				log.Err(err).Msg("User is not authenticated!")
			} else{
				log.Error().Msg("User session is not valid!")
			}
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		handlerFunc(w, r)
	}
}


func hashPassword(password string)(string, error){
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}



