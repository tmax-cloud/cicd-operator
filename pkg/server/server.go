package server

// Server implements webhook server (i.e., event listener server) for git events and report server for jobs

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net/http"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	port = 24335

	paramKeyNamespace  = "namespace"
	paramKeyConfigName = "configName"
	paramKeyJobName    = "jobName"
	paramKeyJobJobName = "jobJobName"
)

var logger = logf.Log.WithName("server")

// Server is a HTTP server for git webhook API and report page
type Server struct {
	k8sClient client.Client
	router    *mux.Router
}

// New is a constructor of a Server
func New(c client.Client, cfg *rest.Config) *Server {
	r := mux.NewRouter()

	// ClientSet
	clientSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		logger.Error(err, "")
		os.Exit(1)
	}

	// Add webhook handler
	r.Methods(http.MethodPost).Subrouter().Handle(webhookPath, &webhookHandler{k8sClient: c})

	// Add report handler
	r.Methods(http.MethodGet).Subrouter().Handle(reportPath, &reportHandler{k8sClient: c, clientSet: clientSet})

	return &Server{
		k8sClient: c,
		router:    r,
	}
}

// Start starts the Server
func (s *Server) Start() {
	httpAddr := fmt.Sprintf("0.0.0.0:%d", port)

	logger.Info(fmt.Sprintf("Server is running on %s", httpAddr))
	if err := http.ListenAndServe(httpAddr, s.router); err != nil {
		logger.Error(err, "cannot launch http server")
		os.Exit(1)
	}
}

func logAndRespond(w http.ResponseWriter, log logr.Logger, code int, respMsg, logMsg string) {
	_ = utils.RespondError(w, code, respMsg)
	log.Info(logMsg)
}
