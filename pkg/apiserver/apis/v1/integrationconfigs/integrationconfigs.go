package integrationconfigs

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/tmax-cloud/cicd-operator/internal/apiserver"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"github.com/tmax-cloud/cicd-operator/internal/wrapper"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// APIVersion of the api
	APIVersion = "v1"

	// ICKind is a kind string of IntegrationConfigs
	ICKind     = "integrationconfigs"
	icParamKey = "icName"
)

type handler struct {
	k8sClient client.Client
	log       logr.Logger

	authorizer apiserver.Authorizer
}

// NewHandler instantiates a new integration configs api handler
func NewHandler(parent wrapper.RouterWrapper, cli client.Client, logger logr.Logger) (apiserver.APIHandler, error) {
	handler := &handler{k8sClient: cli, log: logger}

	// Authorizer
	authClient, err := utils.AuthClient()
	if err != nil {
		return nil, err
	}
	handler.authorizer = apiserver.NewAuthorizer(authClient, apiserver.APIGroup, APIVersion, "create")

	// /integrationconfigs/<integrationconfig>
	icWrapper := wrapper.New(fmt.Sprintf("/%s/{%s}", ICKind, icParamKey), nil, nil)
	if err := parent.Add(icWrapper); err != nil {
		return nil, err
	}
	icWrapper.Router().Use(handler.authorizer.Authorize)

	// /integrationconfigs/<integrationconfig>/runpre
	runPreWrapper := wrapper.New("/runpre", []string{http.MethodPost}, handler.runPreHandler)
	if err := icWrapper.Add(runPreWrapper); err != nil {
		return nil, err
	}

	// /integrationconfigs/<integrationconfig>/runpost
	runPostWrapper := wrapper.New("/runpost", []string{http.MethodPost}, handler.runPostHandler)
	if err := icWrapper.Add(runPostWrapper); err != nil {
		return nil, err
	}

	return handler, nil
}
