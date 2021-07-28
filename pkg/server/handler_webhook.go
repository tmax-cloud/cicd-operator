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

package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
)

var webhookPath = fmt.Sprintf("/webhook/{%s}/{%s}", paramKeyNamespace, paramKeyConfigName)

type webhookHandler struct {
	k8sClient client.Client
}

func (h *webhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	reqID := utils.RandomString(10)
	log := logger.WithValues("request", reqID)

	vars := mux.Vars(r)

	ns, nsExist := vars[paramKeyNamespace]
	configName, configNameExist := vars[paramKeyConfigName]

	if !nsExist || !configNameExist {
		_ = utils.RespondError(w, http.StatusBadRequest, fmt.Sprintf("req: %s, path is not in form of '%s'", reqID, webhookPath))
		log.Info("Bad request for path", "path", r.RequestURI)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		_ = utils.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("req: %s, cannot read webhook body", reqID))
		log.Info("cannot read webhook body", "error", err.Error())
		return
	}

	config := &cicdv1.IntegrationConfig{}
	if err := h.k8sClient.Get(context.Background(), types.NamespacedName{Name: configName, Namespace: ns}, config); err != nil {
		_ = utils.RespondError(w, http.StatusBadRequest, fmt.Sprintf("req: %s, cannot get IntegrationConfig %s/%s", reqID, ns, configName))
		log.Info("Bad request for path", "path", r.RequestURI, "error", err.Error())
		return
	}

	gitCli, err := utils.GetGitCli(config, h.k8sClient)
	if err != nil {
		log.Info("Cannot initialize git cli", "error", err.Error())
		_ = utils.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("req: %s, err: %s", reqID, err.Error()))
		return
	}

	// Convert webhook
	wh, err := gitCli.ParseWebhook(r.Header, body)
	if err != nil {
		_ = utils.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("req: %s, cannot parse webhook body", reqID))
		log.Info("Cannot parse webhook", "error", err.Error())
		return
	}

	if wh == nil {
		return
	}

	// Call plugin functions
	if err := HandleEvent(wh, config); err != nil {
		log.Error(err, "")
	}
}
