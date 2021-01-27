package utils

import (
	"encoding/json"
	"net/http"
)

// RespondJSON responds with arbitrary data objects
func RespondJSON(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")

	j, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = w.Write(j)
	if err != nil {
		return err
	}

	return nil
}

// ErrorResponse is a common struct for responding error for HTTP requests
type ErrorResponse struct {
	Message string `json:"message"`
}

// RespondError responds to a HTTP request with body of ErrorResponse
func RespondError(w http.ResponseWriter, code int, msg string) error {
	w.WriteHeader(code)
	return RespondJSON(w, ErrorResponse{Message: msg})
}
