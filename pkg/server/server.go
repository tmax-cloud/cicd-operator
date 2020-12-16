package server

// Server implements webhook server (i.e., event listener server) for git events and report server for jobs

import (
	"fmt"
	"github.com/gorilla/mux"
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

var log = logf.Log.WithName("server")

type Server struct {
	k8sClient client.Client
	config    *rest.Config
	router    *mux.Router
}

func New(c client.Client, cfg *rest.Config) *Server {
	r := mux.NewRouter()

	// ClientSet
	clientSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Error(err, "")
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

func (s *Server) Start() {
	httpAddr := fmt.Sprintf("0.0.0.0:%d", port)

	log.Info(fmt.Sprintf("Server is running on %s", httpAddr))
	if err := http.ListenAndServe(httpAddr, s.router); err != nil {
		log.Error(err, "cannot launch http server")
		os.Exit(1)
	}
}
