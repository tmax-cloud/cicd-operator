package approvals

import (
	"fmt"
	"github.com/go-logr/logr"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/apiserver"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"net/http"
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
func NewHandler(parent wrapper.RouterWrapper, cli client.Client, logger logr.Logger) (apiserver.APIHandler, error) {
	handler := &handler{k8sClient: cli, log: logger}

	// Authorizer
	authClient, err := utils.AuthClient()
	if err != nil {
		return nil, err
	}
	handler.authorizer = apiserver.NewAuthorizer(authClient, apiserver.APIGroup, APIVersion, "update")

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
