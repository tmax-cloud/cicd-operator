package v1

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/tmax-cloud/cicd-operator/internal/apiserver"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"github.com/tmax-cloud/cicd-operator/internal/wrapper"
	"github.com/tmax-cloud/cicd-operator/pkg/apiserver/apis/v1/approvals"
	"github.com/tmax-cloud/cicd-operator/pkg/apiserver/apis/v1/integrationconfigs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
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
func NewHandler(parent wrapper.RouterWrapper, cli client.Client, logger logr.Logger) (apiserver.APIHandler, error) {
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
	approvalsHandler, err := approvals.NewHandler(namespaceWrapper, cli, logger)
	if err != nil {
		return nil, err
	}
	handler.approvalsHandler = approvalsHandler

	// /v1/namespaces/<namespace>/integrationconfigs
	icHandler, err := integrationconfigs.NewHandler(namespaceWrapper, cli, logger)
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
			Name:       fmt.Sprintf("%s/approve", approvals.ApprovalKind),
			Namespaced: true,
		},
		{
			Name:       fmt.Sprintf("%s/reject", approvals.ApprovalKind),
			Namespaced: true,
		},
		{
			Name:       fmt.Sprintf("%s/runpre", integrationconfigs.ICKind),
			Namespaced: true,
		},
		{
			Name:       fmt.Sprintf("%s/runpost", integrationconfigs.ICKind),
			Namespaced: true,
		},
	}

	_ = utils.RespondJSON(w, apiResourceList)
}
