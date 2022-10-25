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

package integrationconfigs

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/apiserver"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
)

func (h *handler) webhookURLHandler(w http.ResponseWriter, req *http.Request) {
	reqID := utils.RandomString(10)
	log := h.log.WithValues("request", reqID)

	// Get ns/resource name
	vars := mux.Vars(req)

	ns, nsExist := vars[apiserver.NamespaceParamKey]
	resName, nameExist := vars[icParamKey]
	if !nsExist || !nameExist {
		log.Info("url is malformed")
		_ = utils.RespondError(w, http.StatusBadRequest, "url is malformed")
		return
	}

	// Get IntegrationConfig
	ic := &cicdv1.IntegrationConfig{}
	if err := h.k8sClient.Get(context.Background(), types.NamespacedName{Name: resName, Namespace: ns}, ic); err != nil {
		log.Info(err.Error())
		_ = utils.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("req: %s, cannot get IntegrationConfig %s/%s", reqID, ns, resName))
		return
	}

	_ = utils.RespondJSON(w, cicdv1.IntegrationConfigAPIReqWebhookURL{
		URL:    ic.GetWebhookServerAddress(),
		Secret: ic.Status.Secrets,
	})
}
