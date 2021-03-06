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
