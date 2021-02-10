package apis

import (
	"fmt"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/tmax-cloud/cicd-operator/internal/wrapper"
)

// APIGroup and versions
const (
	APIGroup   = "cicdapi.tmax.io"
	APIVersion = "v1"
)

// AddAPIFuncs adds add functions
var AddAPIFuncs []func(wrapper.RouterWrapper, client.Client) error

// AddApis registers apis api for service discovery
func AddApis(parent wrapper.RouterWrapper, cli client.Client) error {
	apiWrapper := wrapper.New("/apis", nil, apisHandler)
	if err := parent.Add(apiWrapper); err != nil {
		return err
	}

	for _, f := range AddAPIFuncs {
		if err := f(apiWrapper, cli); err != nil {
			return err
		}
	}

	return nil
}

func apisHandler(w http.ResponseWriter, _ *http.Request) {
	groupVersion := metav1.GroupVersionForDiscovery{
		GroupVersion: fmt.Sprintf("%s/%s", APIGroup, APIVersion),
		Version:      APIVersion,
	}

	group := metav1.APIGroup{}
	group.Kind = "APIGroup"
	group.Name = APIGroup
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
