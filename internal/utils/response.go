/*
 Copyright 2021 The CI/CD Operator Authors

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

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
