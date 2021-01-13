package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
)

var webhookPath = fmt.Sprintf("/webhook/{%s}/{%s}", paramKeyNamespace, paramKeyConfigName)

type webhookHandler struct {
	k8sClient client.Client
}

func (h *webhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	reqId := utils.RandomString(10)
	log := logger.WithValues("request", reqId)

	vars := mux.Vars(r)

	ns, nsExist := vars[paramKeyNamespace]
	configName, configNameExist := vars[paramKeyConfigName]

	if !nsExist || !configNameExist {
		_ = utils.RespondError(w, http.StatusBadRequest, fmt.Sprintf("req: %s, path is not in form of '%s'", reqId, webhookPath))
		log.Info("Bad request for path", "path", r.RequestURI)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		_ = utils.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("req: %s, cannot read webhook body", reqId))
		log.Info("cannot read webhook body", "error", err.Error())
		return
	}

	config := &cicdv1.IntegrationConfig{}
	if err := h.k8sClient.Get(context.Background(), types.NamespacedName{Name: configName, Namespace: ns}, config); err != nil {
		_ = utils.RespondError(w, http.StatusBadRequest, fmt.Sprintf("req: %s, cannot get IntegrationConfig %s/%s", reqId, ns, configName))
		log.Info("Bad request for path", "path", r.RequestURI, "error", err.Error())
		return
	}

	gitCli, err := utils.GetGitCli(config)
	if err != nil {
		log.Info("Cannot initialize git cli", "error", err.Error())
		_ = utils.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("req: %s, err: %s", reqId, err.Error()))
		return
	}

	// Convert webhook
	wh, err := gitCli.ParseWebhook(config, r.Header, body)
	if err != nil {
		_ = utils.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("req: %s, cannot parse webhook body", reqId))
		log.Info("Cannot parse webhook", "error", err.Error())
		return
	}

	// Call plugin functions
	plugins := GetPlugins(wh.EventType)
	for _, p := range plugins {
		if err := p.Handle(wh, config); err != nil {
			log.Error(err, "")
		}
	}
}
