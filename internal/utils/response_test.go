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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRespondJSON(t *testing.T) {
	tc := map[string]struct {
		w    http.ResponseWriter
		data interface{}

		errorOccurs  bool
		errorMessage string
	}{
		"normal": {
			w:    httptest.NewRecorder(),
			data: struct{}{},
		},
		"marshalErr": {
			w:            httptest.NewRecorder(),
			data:         make(chan struct{}),
			errorOccurs:  true,
			errorMessage: "json: unsupported type: chan struct {}",
		},
		"writeErr": {
			w:            &errorWriter{},
			data:         struct{}{},
			errorOccurs:  true,
			errorMessage: "test error",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			err := RespondJSON(c.w, c.data)
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, "application/json", c.w.Header().Get("Content-Type"))
			}
		})
	}
}

type errorWriter struct{}

func (e *errorWriter) Header() http.Header       { return map[string][]string{} }
func (e *errorWriter) Write([]byte) (int, error) { return 0, fmt.Errorf("test error") }
func (e *errorWriter) WriteHeader(_ int)         {}

func TestRespondError(t *testing.T) {
	w := httptest.NewRecorder()
	require.NoError(t, RespondError(w, 302, "test str"))
	require.Equal(t, 302, w.Result().StatusCode)
	b, err := ioutil.ReadAll(w.Result().Body)
	require.NoError(t, err)
	require.Equal(t, "{\"message\":\"test str\"}", string(b))
}
