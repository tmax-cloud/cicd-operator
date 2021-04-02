package apis

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/tmax-cloud/cicd-operator/internal/apiserver"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	v1 "github.com/tmax-cloud/cicd-operator/pkg/apiserver/apis/v1"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/tmax-cloud/cicd-operator/internal/wrapper"
)

type handler struct {
	v1Handler apiserver.APIHandler
}

// NewHandler instantiates a new apis handler
func NewHandler(parent wrapper.RouterWrapper, cli client.Client, logger logr.Logger) (apiserver.APIHandler, error) {
	handler := &handler{}

	//apis
	apiWrapper := wrapper.New("/apis", nil, handler.apisHandler)
	if err := parent.Add(apiWrapper); err != nil {
		return nil, err
	}

	// /apis/v1
	v1Handler, err := v1.NewHandler(apiWrapper, cli, logger)
	if err != nil {
		return nil, err
	}
	handler.v1Handler = v1Handler

	return handler, nil
}

func (h *handler) apisHandler(w http.ResponseWriter, _ *http.Request) {
	groupVersion := metav1.GroupVersionForDiscovery{
		GroupVersion: fmt.Sprintf("%s/%s", apiserver.APIGroup, v1.APIVersion),
		Version:      v1.APIVersion,
	}

	group := metav1.APIGroup{}
	group.Kind = "APIGroup"
	group.Name = apiserver.APIGroup
	group.PreferredVersion = groupVersion
	group.Versions = append(group.Versions, groupVersion)
	group.ServerAddressByClientCIDRs = append(group.ServerAddressByClientCIDRs, metav1.ServerAddressByClientCIDR{
		ClientCIDR:    "0.0.0.0/0",
		ServerAddress: "",
	})

	apiGroupList := &metav1.APIGroupList{}
	apiGroupList.Kind = "APIGroupList"
	apiGroupList.Groups = append(apiGroupList.Groups, group)

	_ = utils.RespondJSON(w, apiGroupList)
}
