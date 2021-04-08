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
