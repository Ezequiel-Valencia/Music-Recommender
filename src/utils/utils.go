package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)


func WriteJSON(w http.ResponseWriter, output any, status int) error{
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(output)
}

func WriteError(){
	
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


