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

package apis

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"github.com/tmax-cloud/cicd-operator/internal/wrapper"
)

func TestNewHandler(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		p := wrapper.New("/", nil, nil)
		p.SetRouter(mux.NewRouter())
		_, err := NewHandler(p, nil, nil, nil)
		require.NoError(t, err)
	})

	t.Run("apisErr", func(t *testing.T) {
		p := wrapper.New("/", nil, nil)
		_, err := NewHandler(p, nil, nil, nil)
		require.Error(t, err)
		require.Equal(t, "parent does not have a router", err.Error())
	})
}

func Test_handler_apisHandler(t *testing.T) {
	handler := &handler{}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.apisHandler(w, req)

	require.Equal(t, 200, w.Result().StatusCode)
	b, err := ioutil.ReadAll(w.Result().Body)
	require.NoError(t, err)
	require.Equal(t, "{\"kind\":\"APIGroupList\",\"groups\":[{\"kind\":\"APIGroup\",\"name\":\"cicdapi.tmax.io\",\"versions\":[{\"groupVersion\":\"cicdapi.tmax.io/v1\",\"version\":\"v1\"}],\"preferredVersion\":{\"groupVersion\":\"cicdapi.tmax.io/v1\",\"version\":\"v1\"},\"serverAddressByClientCIDRs\":[{\"clientCIDR\":\"0.0.0.0/0\",\"serverAddress\":\"\"}]}]}", string(b))
}
