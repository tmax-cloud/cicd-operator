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
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/apiserver"
	"github.com/tmax-cloud/cicd-operator/internal/wrapper"
	authorization "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// APIVersion of the api
	APIVersion = "v1"

	icParamKey = "icName"
)

type handler struct {
	k8sClient client.Client
	log       logr.Logger

	authorizer apiserver.Authorizer
}

// NewHandler instantiates a new integration configs api handler
func NewHandler(parent wrapper.RouterWrapper, cli client.Client, authCli *authorization.AuthorizationV1Client, logger logr.Logger) (apiserver.APIHandler, error) {
	handler := &handler{k8sClient: cli, log: logger}

	// Authorizer
	handler.authorizer = apiserver.NewAuthorizer(authCli, apiserver.APIGroup, APIVersion, "create")

	// /integrationconfigs/<integrationconfig>
	icWrapper := wrapper.New(fmt.Sprintf("/%s/{%s}", cicdv1.IntegrationConfigKind, icParamKey), nil, nil)
	if err := parent.Add(icWrapper); err != nil {
		return nil, err
	}
	icWrapper.Router().Use(handler.authorizer.Authorize)

	// /integrationconfigs/<integrationconfig>/runpre
	runPreWrapper := wrapper.New("/"+cicdv1.IntegrationConfigAPIRunPre, []string{http.MethodPost}, handler.runPreHandler)
	if err := icWrapper.Add(runPreWrapper); err != nil {
		return nil, err
	}

	// /integrationconfigs/<integrationconfig>/runpost
	runPostWrapper := wrapper.New("/"+cicdv1.IntegrationConfigAPIRunPost, []string{http.MethodPost}, handler.runPostHandler)
	if err := icWrapper.Add(runPostWrapper); err != nil {
		return nil, err
	}

	// /integrationconfigs/<integrationconfig>/webhookurl
	webhookURLWrapper := wrapper.New("/"+cicdv1.IntegrationConfigAPIWebhookURL, []string{http.MethodGet}, handler.webhookURLHandler)
	if err := icWrapper.Add(webhookURLWrapper); err != nil {
		return nil, err
	}

	return handler, nil
}
