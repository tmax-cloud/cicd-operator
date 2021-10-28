/*
 Copyright 2021 The CI/CD Operator Authors

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package apiserver

import (
	"context"
	"fmt"
	"net/http"
	"path"

	"github.com/tmax-cloud/cicd-operator/internal/apiserver"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	authorization "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"

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
	authCli authorization.AuthorizationV1Interface
	cache   cache.Cache

	apisHandler apiserver.APIHandler
}

// +kubebuilder:rbac:groups=apiregistration.k8s.io,resources=apiservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,namespace=kube-system,resourceNames=extension-apiserver-authentication,verbs=get;list;watch

// New is a constructor of server
func New(cli client.Client, cfg *rest.Config, cache cache.Cache) (Server, error) {
	var err error
	srv := &server{}
	srv.wrapper = wrapper.New("/", nil, srv.rootHandler)
	srv.wrapper.SetRouter(mux.NewRouter())
	srv.wrapper.Router().HandleFunc("/", srv.rootHandler)
	srv.client = cli
	srv.cache = cache
	srv.authCli, err = authorization.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	// Set apisHandler
	apisHandler, err := apis.NewHandler(srv.wrapper, srv.client, srv.authCli, log)
	if err != nil {
		return nil, err
	}
	srv.apisHandler = apisHandler

	return srv, nil
}

// Start starts the server
func (s *server) Start() {
	// Wait for the cache init
	if cacheInit := s.cache.WaitForCacheSync(context.Background()); !cacheInit {
		panic(fmt.Errorf("cannot wait for cache init"))
	}

	// Create cert
	if err := createCert(context.Background(), s.client); err != nil {
		panic(err)
	}

	addr := "0.0.0.0:34335"
	log.Info(fmt.Sprintf("API aggregation server is running on %s", addr))

	cfg, err := tlsConfig(context.Background(), s.client)
	if err != nil {
		panic(err)
	}

	httpServer := &http.Server{Addr: addr, Handler: s.wrapper.Router(), TLSConfig: cfg}
	if err := httpServer.ListenAndServeTLS(path.Join(certDir, "tls.crt"), path.Join(certDir, "tls.key")); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

func (s *server) rootHandler(w http.ResponseWriter, _ *http.Request) {
	paths := metav1.RootPaths{}

	addPath(&paths.Paths, s.wrapper)

	_ = utils.RespondJSON(w, paths)
}

// addPath adds all the leaf API endpoints
func addPath(paths *[]string, w wrapper.RouterWrapper) {
	if w.Handler() != nil {
		*paths = append(*paths, w.FullPath())
	}

	for _, c := range w.Children() {
		addPath(paths, c)
	}
}
