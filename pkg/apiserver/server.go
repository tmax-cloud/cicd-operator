package apiserver

import (
	"context"
	"fmt"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
	"os"
	"path"

	"github.com/gorilla/mux"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/tmax-cloud/cicd-operator/internal/wrapper"
	"github.com/tmax-cloud/cicd-operator/pkg/apiserver/apis"
)

const (
	Port = 34335
)

var log = logf.Log.WithName("approve-server")

type Server struct {
	Wrapper *wrapper.RouterWrapper
	Client  client.Client
}

// +kubebuilder:rbac:groups=apiregistration.k8s.io,resources=apiservices,resourceNames=v1.cicdapi.tmax.io,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,namespace=kube-system,resourceNames=extension-apiserver-authentication,verbs=get;list;watch

func New(scheme *runtime.Scheme) *Server {
	server := &Server{}
	server.Wrapper = wrapper.New("/", nil, server.rootHandler)
	server.Wrapper.Router = mux.NewRouter()
	server.Wrapper.Router.HandleFunc("/", server.rootHandler)

	cli, err := utils.Client(scheme)
	if err != nil {
		log.Error(err, "cannot get client")
		os.Exit(1)
	}

	server.Client = cli

	if err := apis.AddApis(server.Wrapper, cli); err != nil {
		log.Error(err, "cannot add apis")
		os.Exit(1)
	}

	if err := createCert(context.Background(), server.Client); err != nil {
		log.Error(err, "cannot create cert")
		os.Exit(1)
	}

	return server
}

func (s *Server) Start() {
	addr := "0.0.0.0:34335"
	log.Info(fmt.Sprintf("API aggregation server is running on %s", addr))

	cfg, err := tlsConfig(context.Background(), s.Client)
	if err != nil {
		log.Error(err, "cannot get tls config")
		os.Exit(1)
	}

	httpServer := &http.Server{Addr: addr, Handler: s.Wrapper.Router, TLSConfig: cfg}
	if err := httpServer.ListenAndServeTLS(path.Join(CertDir, "tls.crt"), path.Join(CertDir, "tls.key")); err != nil {
		log.Error(err, "cannot launch server")
		os.Exit(1)
	}
}

func (s *Server) rootHandler(w http.ResponseWriter, _ *http.Request) {
	paths := metav1.RootPaths{}

	addPath(&paths.Paths, s.Wrapper)

	_ = utils.RespondJSON(w, paths)
}

func addPath(paths *[]string, w *wrapper.RouterWrapper) {
	if w.Handler != nil {
		*paths = append(*paths, w.FullPath())
	}

	for _, c := range w.Children {
		addPath(paths, c)
	}
}
