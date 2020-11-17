package webhook

// Webhook implements webhook server (i.e., event listener server) for git events

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	port = 24335
)

var log = logf.Log.WithName("webhook-server")

type Server struct {
	k8sClient client.Client
	router    *mux.Router
}

func New(c client.Client) *Server {
	r := mux.NewRouter()

	// Add webhook handler
	r.Methods(http.MethodPost).Subrouter().Handle(webhookPath, &webhookHandler{k8sClient: c})

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
