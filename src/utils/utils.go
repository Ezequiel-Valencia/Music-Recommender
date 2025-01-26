package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)


func WriteJSON(w http.ResponseWriter, output any, status int) error{
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(output)
}

type MalformedRequest struct {
    Status int
    Msg    string
}

func (mr *MalformedRequest) Error() string {
    return mr.Msg
}

func DecodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
    ct := r.Header.Get("Content-Type")
    if ct != "" {
        mediaType := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
        if mediaType != "application/json" {
            msg := "Content-Type header is not application/json"
            return &MalformedRequest{Status: http.StatusUnsupportedMediaType, Msg: msg}
        }
    }

    r.Body = http.MaxBytesReader(w, r.Body, 1048576)

    dec := json.NewDecoder(r.Body)
    dec.DisallowUnknownFields()
    var malformedRequest MalformedRequest

    err := dec.Decode(&dst)
    if err != nil {
        var syntaxError *json.SyntaxError
        var unmarshalTypeError *json.UnmarshalTypeError
        

        switch {
        case errors.As(err, &syntaxError):
            msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
            malformedRequest = MalformedRequest{Status: http.StatusBadRequest, Msg: msg}

        case errors.Is(err, io.ErrUnexpectedEOF):
            msg := "Request body contains badly-formed JSON"
            malformedRequest = MalformedRequest{Status: http.StatusBadRequest, Msg: msg}

        case errors.As(err, &unmarshalTypeError):
            msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
            malformedRequest = MalformedRequest{Status: http.StatusBadRequest, Msg: msg}

        case strings.HasPrefix(err.Error(), "json: unknown field "):
            fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
            msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
            malformedRequest = MalformedRequest{Status: http.StatusBadRequest, Msg: msg}

        case errors.Is(err, io.EOF):
            msg := "Request body must not be empty"
            malformedRequest = MalformedRequest{Status: http.StatusBadRequest, Msg: msg}

        case err.Error() == "http: request body too large":
            msg := "Request body must not be larger than 1MB"
            malformedRequest = MalformedRequest{Status: http.StatusRequestEntityTooLarge, Msg: msg}

        default:
            return err
        }
        http.Error(w, malformedRequest.Msg, malformedRequest.Status)
        return &malformedRequest
    }

	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
        msg := "Request body must only contain a single JSON object"
        malformedRequest = MalformedRequest{Status: http.StatusBadRequest, Msg: msg}
        http.Error(w, malformedRequest.Msg, malformedRequest.Status)
        return &malformedRequest
    }

    return nil
}


func IsStringAlphaNumeric(text string) bool{
    return regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString(text)
}

func IsStringAlphaNumericWithPunctuation(text string) bool{
    return regexp.MustCompile(`^[a-zA-Z0-9 ()!.]*$`).MatchString(text)
}

// The extra chars allowed by this function is as follows
// (,)!@#$%^&*.:-;_
func IsStringANWithExtraChars(text string) bool {
    return regexp.MustCompile(`^[a-zA-Z0-9()!@#$%^&*.,:;\-_]*$`).MatchString(text)

}

func IsProperYouTubeLink(link string) bool{
    _, err := GetResourceFromYouTubeLink(link)
    return err == nil
}

func GetResourceFromYouTubeLink(link string) (string, error){
    cutUp := strings.Split(link, "/")
    fmt.Println(cutUp)
    pathAndQuery := strings.Split(cutUp[3], "?")
    path := pathAndQuery[0]
    properPath := regexp.MustCompile(`^[a-zA-Z0-9\-]*$`).MatchString(path)
    if (!properPath){
        return "", errors.New("not a proper path")
    }
    return path, nil
}

func executeAtXMath(hour int, now time.Time) time.Time{
    if (hour < 0 || hour > 23){
        log.Fatalf("Hour set for execution was invalid: %d", hour)
    }
    var day int
    if hour < now.Hour(){
        day = now.AddDate(0, 0, 1).Day()
    } else{
        day = now.Day()
    }
    newDate := time.Date(now.Year(), now.Month(), day, hour, 0, 0, 0, time.Local)
    if (newDate.Before(now)){
        log.Fatal("New sleep date is before the present time.")
    }
    return newDate
}

/* Within a 24 hour time system, sleep until that set hour
   Starts at 0:00 (12am) and ends at 23:00 (11pm)
*/
func SleepUntilXHour(hour int){
    now := time.Now()
    time.Sleep(executeAtXMath(hour, now).Sub(now))
}

