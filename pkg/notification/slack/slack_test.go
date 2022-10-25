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

package slack

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

const (
	testMessage = "MessageMessageMessage"
)

func TestSendMessage(t *testing.T) {
	errChan := make(chan error)

	srv := newTestServer()
	go func() {
		if err := SendMessage(srv.URL, testMessage); err != nil {
			errChan <- err
		}
		errChan <- nil
	}()

	for err := range errChan {
		if err != nil {
			t.Fatal(err)
		} else {
			return
		}
	}
}

func newTestServer() *httptest.Server {
	router := mux.NewRouter()

	router.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			_ = req.Body.Close()
		}()
		userReq := Message{}
		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&userReq); err != nil {
			_ = utils.RespondError(w, http.StatusBadRequest, fmt.Sprintf("body is not in json form or is malformed, err : %s", err.Error()))
			return
		}

		desired := Message{
			Text: messageTitle,
			Blocks: []MessageBlock{{
				Type: "section",
				Text: BlockText{
					Type: "mrkdwn",
					Text: testMessage,
				},
			}},
		}

		if !reflect.DeepEqual(userReq, desired) {
			_ = utils.RespondError(w, http.StatusBadRequest, fmt.Sprintf("user request (%+v) != desired (%+v)", userReq, desired))
			return
		}
	})

	return httptest.NewServer(router)
}
