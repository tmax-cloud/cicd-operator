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

package approvals

import (
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/apiserver"
	authorization "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/tmax-cloud/cicd-operator/internal/wrapper"
)

const (
	// APIVersion for the api
	APIVersion = "v1"

	approvalNameParamKey = "approvalName"
)

type handler struct {
	k8sClient client.Client
	log       logr.Logger

	authorizer apiserver.Authorizer
}

// NewHandler instantiates a new approvals api handler
func NewHandler(parent wrapper.RouterWrapper, cli client.Client, authCli authorization.AuthorizationV1Interface, logger logr.Logger) (apiserver.APIHandler, error) {
	handler := &handler{k8sClient: cli, log: logger}

	// Authorizer
	handler.authorizer = apiserver.NewAuthorizer(authCli, apiserver.APIGroup, APIVersion, "update")

	// /approvals/<approval>
	approvalWrapper := wrapper.New(fmt.Sprintf("/%s/{%s}", cicdv1.ApprovalKind, approvalNameParamKey), nil, nil)
	if err := parent.Add(approvalWrapper); err != nil {
		return nil, err
	}
	approvalWrapper.Router().Use(handler.authorizer.Authorize)

	// /approvals/<approval>/approve
	approveWrapper := wrapper.New("/"+cicdv1.ApprovalAPIApprove, []string{http.MethodPut}, handler.approveHandler)
	if err := approvalWrapper.Add(approveWrapper); err != nil {
		return nil, err
	}

	// /approvals/<approval>/reject
	rejectWrapper := wrapper.New("/"+cicdv1.ApprovalAPIReject, []string{http.MethodPut}, handler.rejectHandler)
	if err := approvalWrapper.Add(rejectWrapper); err != nil {
		return nil, err
	}

	return handler, nil
}
