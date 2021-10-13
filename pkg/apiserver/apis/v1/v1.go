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
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/apiserver"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"github.com/tmax-cloud/cicd-operator/internal/wrapper"
	"github.com/tmax-cloud/cicd-operator/pkg/apiserver/apis/v1/approvals"
	"github.com/tmax-cloud/cicd-operator/pkg/apiserver/apis/v1/integrationconfigs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	authorization "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// APIVersion of the apis
	APIVersion = "v1"
)

type handler struct {
	approvalsHandler apiserver.APIHandler
	icHandler        apiserver.APIHandler
}

// NewHandler instantiates a new v1 api handler
func NewHandler(parent wrapper.RouterWrapper, cli client.Client, authCli *authorization.AuthorizationV1Client, logger logr.Logger) (apiserver.APIHandler, error) {
	handler := &handler{}

	// /v1
	versionWrapper := wrapper.New(fmt.Sprintf("/%s/%s", apiserver.APIGroup, APIVersion), nil, handler.versionHandler)
	if err := parent.Add(versionWrapper); err != nil {
		return nil, err
	}

	// /v1/namespaces/<namespace>
	namespaceWrapper := wrapper.New("/namespaces/{namespace}", nil, nil)
	if err := versionWrapper.Add(namespaceWrapper); err != nil {
		return nil, err
	}

	// /v1/namespaces/<namespace>/approvals
	approvalsHandler, err := approvals.NewHandler(namespaceWrapper, cli, authCli, logger)
	if err != nil {
		return nil, err
	}
	handler.approvalsHandler = approvalsHandler

	// /v1/namespaces/<namespace>/integrationconfigs
	icHandler, err := integrationconfigs.NewHandler(namespaceWrapper, cli, authCli, logger)
	if err != nil {
		return nil, err
	}
	handler.icHandler = icHandler

	return handler, nil
}

func (h *handler) versionHandler(w http.ResponseWriter, _ *http.Request) {
	apiResourceList := &metav1.APIResourceList{}
	apiResourceList.Kind = "APIResourceList"
	apiResourceList.GroupVersion = fmt.Sprintf("%s/%s", apiserver.APIGroup, APIVersion)
	apiResourceList.APIVersion = APIVersion

	apiResourceList.APIResources = []metav1.APIResource{
		{
			Name:       fmt.Sprintf("%s/%s", cicdv1.ApprovalKind, cicdv1.ApprovalAPIApprove),
			Namespaced: true,
		},
		{
			Name:       fmt.Sprintf("%s/%s", cicdv1.ApprovalKind, cicdv1.ApprovalAPIReject),
			Namespaced: true,
		},
		{
			Name:       fmt.Sprintf("%s/%s", cicdv1.IntegrationConfigKind, cicdv1.IntegrationConfigAPIRunPre),
			Namespaced: true,
		},
		{
			Name:       fmt.Sprintf("%s/%s", cicdv1.IntegrationConfigKind, cicdv1.IntegrationConfigAPIRunPost),
			Namespaced: true,
		},
		{
			Name:       fmt.Sprintf("%s/%s", cicdv1.IntegrationConfigKind, cicdv1.IntegrationConfigAPIWebhookURL),
			Namespaced: true,
		},
	}

	_ = utils.RespondJSON(w, apiResourceList)
}
