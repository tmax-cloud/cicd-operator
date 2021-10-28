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

package v1

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

	t.Run("versionErr", func(t *testing.T) {
		p := wrapper.New("/", nil, nil)
		_, err := NewHandler(p, nil, nil, nil)
		require.Error(t, err)
		require.Equal(t, "parent does not have a router", err.Error())
	})
}

func Test_handler_versionHandler(t *testing.T) {
	handler := &handler{}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.versionHandler(w, req)

	require.Equal(t, 200, w.Result().StatusCode)
	b, err := ioutil.ReadAll(w.Result().Body)
	require.NoError(t, err)
	require.Equal(t, "{\"kind\":\"APIResourceList\",\"apiVersion\":\"v1\",\"groupVersion\":\"cicdapi.tmax.io/v1\",\"resources\":[{\"name\":\"approvals/approve\",\"singularName\":\"\",\"namespaced\":true,\"kind\":\"\",\"verbs\":null},{\"name\":\"approvals/reject\",\"singularName\":\"\",\"namespaced\":true,\"kind\":\"\",\"verbs\":null},{\"name\":\"integrationconfigs/runpre\",\"singularName\":\"\",\"namespaced\":true,\"kind\":\"\",\"verbs\":null},{\"name\":\"integrationconfigs/runpost\",\"singularName\":\"\",\"namespaced\":true,\"kind\":\"\",\"verbs\":null},{\"name\":\"integrationconfigs/webhookurl\",\"singularName\":\"\",\"namespaced\":true,\"kind\":\"\",\"verbs\":null}]}", string(b))
}
