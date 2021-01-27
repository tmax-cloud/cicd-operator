package v1

import (
	"fmt"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	authorization "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/tmax-cloud/cicd-operator/internal/wrapper"
)

// APIGroup and versions
const (
	APIGroup     = "cicdapi.tmax.io"
	APIVersion   = "v1"
	approvalKind = "approvals"
)

var logger = logf.Log.WithName("approve-apis")

var k8sClient client.Client
var k8sCliLock sync.Mutex

var authClient *authorization.AuthorizationV1Client
var authCliLock sync.Mutex

type reqBody struct {
	Reason string `json:"reason"`
}

// AddV1Apis adds v1 api
func AddV1Apis(parent *wrapper.RouterWrapper, cli client.Client) error {
	versionWrapper := wrapper.New(fmt.Sprintf("/%s/%s", APIGroup, APIVersion), nil, versionHandler)
	if err := parent.Add(versionWrapper); err != nil {
		return err
	}

	namespaceWrapper := wrapper.New("/namespaces/{namespace}", nil, nil)
	if err := versionWrapper.Add(namespaceWrapper); err != nil {
		return err
	}

	return addApprovalApis(namespaceWrapper, cli)
}

func versionHandler(w http.ResponseWriter, _ *http.Request) {
	apiResourceList := &metav1.APIResourceList{}
	apiResourceList.Kind = "APIResourceList"
	apiResourceList.GroupVersion = fmt.Sprintf("%s/%s", APIGroup, APIVersion)
	apiResourceList.APIVersion = APIVersion

	apiResourceList.APIResources = []metav1.APIResource{
		{
			Name:       fmt.Sprintf("%s/approve", approvalKind),
			Namespaced: true,
		},
		{
			Name:       fmt.Sprintf("%s/reject", approvalKind),
			Namespaced: true,
		},
	}

	_ = utils.RespondJSON(w, apiResourceList)
}
