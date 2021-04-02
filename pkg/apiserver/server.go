package apiserver

import (
	"context"
	"fmt"
	"github.com/tmax-cloud/cicd-operator/internal/apiserver"
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

var log = logf.Log.WithName("api-server")

// Server is an interface of server
type Server interface {
	Start()
}

// server is an api server
type server struct {
	wrapper wrapper.RouterWrapper
	client  client.Client

	apisHandler apiserver.APIHandler
}

// +kubebuilder:rbac:groups=apiregistration.k8s.io,resources=apiservices,resourceNames=v1.cicdapi.tmax.io,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,namespace=kube-system,resourceNames=extension-apiserver-authentication,verbs=get;list;watch

// New is a constructor of server
func New(scheme *runtime.Scheme) *server {
	srv := &server{}
	srv.wrapper = wrapper.New("/", nil, srv.rootHandler)
	srv.wrapper.SetRouter(mux.NewRouter())
	srv.wrapper.Router().HandleFunc("/", srv.rootHandler)

	// Set client
	cli, err := utils.Client(scheme)
	if err != nil {
		log.Error(err, "cannot get client")
		os.Exit(1)
	}
	srv.client = cli

	// Set apisHandler
	apisHandler, err := apis.NewHandler(srv.wrapper, srv.client, log)
	if err != nil {
		log.Error(err, "cannot add apis")
		os.Exit(1)
	}
	srv.apisHandler = apisHandler

	if err := createCert(context.Background(), srv.client); err != nil {
		log.Error(err, "cannot create cert")
		os.Exit(1)
	}

	return srv
}

// Start starts the server
func (s *server) Start() {
	addr := "0.0.0.0:34335"
	log.Info(fmt.Sprintf("API aggregation server is running on %s", addr))

	cfg, err := tlsConfig(context.Background(), s.client)
	if err != nil {
		log.Error(err, "cannot get tls config")
		os.Exit(1)
	}

	httpServer := &http.Server{Addr: addr, Handler: s.wrapper.Router(), TLSConfig: cfg}
	if err := httpServer.ListenAndServeTLS(path.Join(certDir, "tls.crt"), path.Join(certDir, "tls.key")); err != nil {
		log.Error(err, "cannot launch server")
		os.Exit(1)
	}
}

func (s *server) rootHandler(w http.ResponseWriter, _ *http.Request) {
	paths := metav1.RootPaths{}

	addPath(&paths.Paths, s.wrapper)

	_ = utils.RespondJSON(w, paths)
}

func addPath(paths *[]string, w wrapper.RouterWrapper) {
	if w.Handler() != nil {
		*paths = append(*paths, w.FullPath())
	}

	for _, c := range w.Children() {
		addPath(paths, c)
	}
}
