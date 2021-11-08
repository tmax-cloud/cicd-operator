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

package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

type groupVersionResourceClientTestCase struct {
	ns  string
	gvr schema.GroupVersionResource

	expectedApiPath string
}

func TestNewGroupVersionResourceClient(t *testing.T) {
	apiServer := testApiServer(mux.NewRouter())

	cfg := &rest.Config{
		Host: apiServer.URL,
	}

	tc := map[string]groupVersionResourceClientTestCase{
		"pods": {
			ns:              "test-ns",
			gvr:             corev1.SchemeGroupVersion.WithResource("pods"),
			expectedApiPath: "/api/v1",
		},
		"ingresses": {
			ns:              "test-ns2",
			gvr:             networkingv1.SchemeGroupVersion.WithResource("ingresses"),
			expectedApiPath: "/apis/networking.k8s.io/v1",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			cliIn, err := NewGroupVersionResourceClient(cfg, c.ns, c.gvr)
			require.NoError(t, err)

			cli := cliIn.(*gvrClient)

			require.Equal(t, c.ns, cli.ns)
			require.Equal(t, c.gvr.Resource, cli.resource)

			require.Equal(t, c.expectedApiPath, cli.client.Get().URL().Path)
		})
	}
}

type gvrClientRequestTestCase struct {
	ns   string
	name string

	gvr  schema.GroupVersionResource
	objs []runtime.Object

	errorOccurs  bool
	errorMessage string
}

func TestGvrClient_Get(t *testing.T) {
	router := mux.NewRouter()
	router.Path("/api/v1/namespaces/test-ns/pods/pod1").Methods(http.MethodGet).HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod1",
				Namespace: "test-ns",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(pod)
	})
	router.Path("/apis/networking.k8s.io/v1/namespaces/test-ns/ingresses/ing").Methods(http.MethodGet).HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ing := &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ing",
				Namespace: "test-ns",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ing)
	})

	apiServer := testApiServer(router)
	cfg := &rest.Config{
		Host:            apiServer.URL,
		TLSClientConfig: rest.TLSClientConfig{Insecure: true},
	}

	tc := map[string]gvrClientRequestTestCase{
		"pod": {
			ns:          "test-ns",
			name:        "pod1",
			gvr:         corev1.SchemeGroupVersion.WithResource("pods"),
			objs:        []runtime.Object{&corev1.Pod{}},
			errorOccurs: false,
		},
		"podNotExist": {
			ns:           "test-ns",
			name:         "pod-not-exist",
			gvr:          corev1.SchemeGroupVersion.WithResource("pods"),
			objs:         []runtime.Object{&corev1.Pod{}},
			errorOccurs:  true,
			errorMessage: "the server could not find the requested resource (get pods pod-not-exist)",
		},
		"ingress": {
			ns:   "test-ns",
			name: "ing",
			gvr:  networkingv1.SchemeGroupVersion.WithResource("ingresses"),
			objs: []runtime.Object{&networkingv1.Ingress{}},
		},
		"ingressNotExist": {
			ns:           "test-ns",
			name:         "ing-not-exist",
			gvr:          networkingv1.SchemeGroupVersion.WithResource("ingresses"),
			objs:         []runtime.Object{&networkingv1.Ingress{}},
			errorOccurs:  true,
			errorMessage: "the server could not find the requested resource (get ingresses.networking.k8s.io ing-not-exist)",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			cli, err := NewGroupVersionResourceClient(cfg, c.ns, c.gvr)
			require.NoError(t, err)

			err = cli.Get(c.name, &metav1.GetOptions{}, c.objs[0])
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGvrClient_List(t *testing.T) {
	router := mux.NewRouter()
	router.Path("/api/v1/namespaces/test-ns/pods").Methods(http.MethodGet).HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		list := &corev1.PodList{}
		_ = json.NewEncoder(w).Encode(list)
	})
	router.Path("/apis/networking.k8s.io/v1/namespaces/test-ns/ingresses").Methods(http.MethodGet).HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		list := &networkingv1.IngressList{}
		_ = json.NewEncoder(w).Encode(list)
	})

	apiServer := testApiServer(router)
	cfg := &rest.Config{
		Host:            apiServer.URL,
		TLSClientConfig: rest.TLSClientConfig{Insecure: true},
	}

	tc := map[string]gvrClientRequestTestCase{
		"pod": {
			ns:          "test-ns",
			name:        "pod1",
			gvr:         corev1.SchemeGroupVersion.WithResource("pods"),
			objs:        []runtime.Object{&corev1.PodList{}},
			errorOccurs: false,
		},
		"ingress": {
			ns:   "test-ns",
			name: "ing",
			gvr:  networkingv1.SchemeGroupVersion.WithResource("ingresses"),
			objs: []runtime.Object{&networkingv1.IngressList{}},
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			cli, err := NewGroupVersionResourceClient(cfg, c.ns, c.gvr)
			require.NoError(t, err)

			err = cli.List(&metav1.ListOptions{}, c.objs[0])
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGvrClient_Watch(t *testing.T) {
	router := mux.NewRouter()
	router.Path("/api/v1/namespaces/test-ns/pods").Methods(http.MethodGet).HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		flusher := w.(http.Flusher)
		if req.URL.Query().Get("watch") != "true" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		objs := []corev1.Pod{
			{TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Pod"}, ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "test-ns"}},
		}

		fs := req.URL.Query().Get("fieldSelector")
		if fs == "" || !strings.HasPrefix(fs, "metadata.name=") {
			// All
			objs = append(objs, corev1.Pod{TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Pod"}, ObjectMeta: metav1.ObjectMeta{Name: "pod2", Namespace: "test-ns"}})
		}

		w.Header().Set("Content-Type", "application/json")

		for _, o := range objs {
			body, _ := json.Marshal(o)
			_, _ = fmt.Fprintf(w, `{"type": "ADDED", "object": %s}`, string(body))
			flusher.Flush()
		}
	})
	router.Path("/apis/networking.k8s.io/v1/namespaces/test-ns/ingresses").Methods(http.MethodGet).HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		flusher := w.(http.Flusher)
		if req.URL.Query().Get("watch") != "true" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		objs := []networkingv1.Ingress{
			{TypeMeta: metav1.TypeMeta{APIVersion: "networking.k8s.io/v1", Kind: "Ingress"}, ObjectMeta: metav1.ObjectMeta{Name: "ing", Namespace: "test-ns"}},
		}

		fs := req.URL.Query().Get("fieldSelector")
		if fs == "" || !strings.HasPrefix(fs, "metadata.name=") {
			// All
			objs = append(objs, networkingv1.Ingress{TypeMeta: metav1.TypeMeta{APIVersion: "networking.k8s.io/v1", Kind: "Ingress"}, ObjectMeta: metav1.ObjectMeta{Name: "ing2", Namespace: "test-ns"}})
		}

		w.Header().Set("Content-Type", "application/json")

		for _, o := range objs {
			body, _ := json.Marshal(o)
			_, _ = fmt.Fprintf(w, `{"type": "ADDED", "object": %s}`, string(body))
			flusher.Flush()
		}
	})

	apiServer := testApiServer(router)
	cfg := &rest.Config{
		Host:            apiServer.URL,
		TLSClientConfig: rest.TLSClientConfig{Insecure: true},
	}

	tc := map[string]gvrClientRequestTestCase{
		"pod": {
			ns:          "test-ns",
			name:        "pod1",
			gvr:         corev1.SchemeGroupVersion.WithResource("pods"),
			objs:        []runtime.Object{&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "test-ns"}}},
			errorOccurs: false,
		},
		"podAll": {
			ns:  "test-ns",
			gvr: corev1.SchemeGroupVersion.WithResource("pods"),
			objs: []runtime.Object{
				&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "test-ns"}},
				&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod2", Namespace: "test-ns"}},
			},
		},
		"ingress": {
			ns:   "test-ns",
			name: "ing",
			gvr:  networkingv1.SchemeGroupVersion.WithResource("ingresses"),
			objs: []runtime.Object{&networkingv1.Ingress{ObjectMeta: metav1.ObjectMeta{Name: "ing", Namespace: "test-ns"}}},
		},
		"ingressAll": {
			ns:  "test-ns",
			gvr: networkingv1.SchemeGroupVersion.WithResource("ingresses"),
			objs: []runtime.Object{
				&networkingv1.Ingress{ObjectMeta: metav1.ObjectMeta{Name: "ing", Namespace: "test-ns"}},
				&networkingv1.Ingress{ObjectMeta: metav1.ObjectMeta{Name: "ing2", Namespace: "test-ns"}},
			},
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			cli, err := NewGroupVersionResourceClient(cfg, c.ns, c.gvr)
			require.NoError(t, err)

			selector := ""
			if c.name != "" {
				selector = fields.OneTermEqualSelector(metav1.ObjectNameField, c.name).String()
			}
			watcher, err := cli.Watch(&metav1.ListOptions{FieldSelector: selector})
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)

				for _, expectedObj := range c.objs {
					event := <-watcher.ResultChan()
					require.Equal(t, expectedObj, event.Object)
				}
			}
		})
	}
}

func TestGvrClient_Create(t *testing.T) {
	router := mux.NewRouter()
	router.Path("/api/v1/namespaces/test-ns/pods").Methods(http.MethodPost).HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		pod := &corev1.Pod{}
		if err := json.NewDecoder(req.Body).Decode(pod); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if pod.Name != "pod1" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(pod)
	})
	router.Path("/apis/networking.k8s.io/v1/namespaces/test-ns/ingresses").Methods(http.MethodPost).HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ing := &networkingv1.Ingress{}
		if err := json.NewDecoder(req.Body).Decode(ing); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ing.Name != "ing" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ing)
	})

	apiServer := testApiServer(router)
	cfg := &rest.Config{
		Host:            apiServer.URL,
		TLSClientConfig: rest.TLSClientConfig{Insecure: true},
	}

	tc := map[string]gvrClientRequestTestCase{
		"pod": {
			ns:          "test-ns",
			name:        "pod1",
			gvr:         corev1.SchemeGroupVersion.WithResource("pods"),
			objs:        []runtime.Object{&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "test-ns", ResourceVersion: "testversion"}}},
			errorOccurs: false,
		},
		"ingress": {
			ns:   "test-ns",
			name: "ing",
			gvr:  networkingv1.SchemeGroupVersion.WithResource("ingresses"),
			objs: []runtime.Object{&networkingv1.Ingress{ObjectMeta: metav1.ObjectMeta{Name: "ing", Namespace: "test-ns", ResourceVersion: "testversion"}}},
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			cli, err := NewGroupVersionResourceClient(cfg, c.ns, c.gvr)
			require.NoError(t, err)

			err = cli.Create(c.objs[0], &metav1.CreateOptions{})
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGvrClient_Update(t *testing.T) {
	router := mux.NewRouter()
	router.Path("/api/v1/namespaces/test-ns/pods/pod1").Methods(http.MethodPut).HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		pod := &corev1.Pod{}
		_ = json.NewDecoder(req.Body).Decode(pod)
		if pod.ResourceVersion == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(pod)
	})
	router.Path("/apis/networking.k8s.io/v1/namespaces/test-ns/ingresses/ing").Methods(http.MethodPut).HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ing := &networkingv1.Ingress{}
		_ = json.NewDecoder(req.Body).Decode(ing)
		if ing.ResourceVersion == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ing)
	})

	apiServer := testApiServer(router)
	cfg := &rest.Config{
		Host:            apiServer.URL,
		TLSClientConfig: rest.TLSClientConfig{Insecure: true},
	}

	tc := map[string]gvrClientRequestTestCase{
		"pod": {
			ns:          "test-ns",
			name:        "pod1",
			gvr:         corev1.SchemeGroupVersion.WithResource("pods"),
			objs:        []runtime.Object{&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "test-ns", ResourceVersion: "testversion"}}},
			errorOccurs: false,
		},
		"podNotExist": {
			ns:           "test-ns",
			name:         "pod-not-exist",
			gvr:          corev1.SchemeGroupVersion.WithResource("pods"),
			objs:         []runtime.Object{&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod-not-exist", Namespace: "test-ns", ResourceVersion: "testversion"}}},
			errorOccurs:  true,
			errorMessage: "the server could not find the requested resource (put pods pod-not-exist)",
		},
		"podNoResourceVersion": {
			ns:           "test-ns",
			name:         "pod-not-exist",
			gvr:          corev1.SchemeGroupVersion.WithResource("pods"),
			objs:         []runtime.Object{&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "test-ns"}}},
			errorOccurs:  true,
			errorMessage: "the server rejected our request for an unknown reason (put pods pod1)",
		},
		"ingress": {
			ns:   "test-ns",
			name: "ing",
			gvr:  networkingv1.SchemeGroupVersion.WithResource("ingresses"),
			objs: []runtime.Object{&networkingv1.Ingress{ObjectMeta: metav1.ObjectMeta{Name: "ing", Namespace: "test-ns", ResourceVersion: "testversion"}}},
		},
		"ingressNotExist": {
			ns:           "test-ns",
			name:         "ing-not-exist",
			gvr:          networkingv1.SchemeGroupVersion.WithResource("ingresses"),
			objs:         []runtime.Object{&networkingv1.Ingress{ObjectMeta: metav1.ObjectMeta{Name: "ing-not-exist", Namespace: "test-ns", ResourceVersion: "testversion"}}},
			errorOccurs:  true,
			errorMessage: "the server could not find the requested resource (put ingresses.networking.k8s.io ing-not-exist)",
		},
		"ingressNoResourceVersion": {
			ns:           "test-ns",
			name:         "ing-not-exist",
			gvr:          networkingv1.SchemeGroupVersion.WithResource("ingresses"),
			objs:         []runtime.Object{&networkingv1.Ingress{ObjectMeta: metav1.ObjectMeta{Name: "ing", Namespace: "test-ns"}}},
			errorOccurs:  true,
			errorMessage: "the server rejected our request for an unknown reason (put ingresses.networking.k8s.io ing)",
		},
		"notMetaAccessor": {
			ns:           "test-ns",
			name:         "pod1",
			gvr:          corev1.SchemeGroupVersion.WithResource("pods"),
			objs:         []runtime.Object{&testObject{}},
			errorOccurs:  true,
			errorMessage: "object does not implement the Object interfaces",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			cli, err := NewGroupVersionResourceClient(cfg, c.ns, c.gvr)
			require.NoError(t, err)

			err = cli.Update(c.objs[0], &metav1.UpdateOptions{})
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

type testObject struct{}

func (t *testObject) GetObjectKind() schema.ObjectKind { return &metav1.APIGroup{} }
func (t *testObject) DeepCopyObject() runtime.Object   { return t }

func TestGvrClient_Delete(t *testing.T) {
	router := mux.NewRouter()
	router.Path("/api/v1/namespaces/test-ns/pods/pod1").Methods(http.MethodDelete).HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	})
	router.Path("/apis/networking.k8s.io/v1/namespaces/test-ns/ingresses/ing").Methods(http.MethodDelete).HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	})

	apiServer := testApiServer(router)
	cfg := &rest.Config{
		Host:            apiServer.URL,
		TLSClientConfig: rest.TLSClientConfig{Insecure: true},
	}

	tc := map[string]gvrClientRequestTestCase{
		"pod": {
			ns:          "test-ns",
			name:        "pod1",
			gvr:         corev1.SchemeGroupVersion.WithResource("pods"),
			errorOccurs: false,
		},
		"podNotExist": {
			ns:           "test-ns",
			name:         "pod-not-exist",
			gvr:          corev1.SchemeGroupVersion.WithResource("pods"),
			errorOccurs:  true,
			errorMessage: "the server could not find the requested resource (delete pods pod-not-exist)",
		},
		"ingress": {
			ns:   "test-ns",
			name: "ing",
			gvr:  networkingv1.SchemeGroupVersion.WithResource("ingresses"),
		},
		"ingressNotExist": {
			ns:           "test-ns",
			name:         "ing-not-exist",
			gvr:          networkingv1.SchemeGroupVersion.WithResource("ingresses"),
			errorOccurs:  true,
			errorMessage: "the server could not find the requested resource (delete ingresses.networking.k8s.io ing-not-exist)",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			cli, err := NewGroupVersionResourceClient(cfg, c.ns, c.gvr)
			require.NoError(t, err)

			err = cli.Delete(c.name, &metav1.DeleteOptions{})
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func testApiServer(m http.Handler) *httptest.Server {
	return httptest.NewTLSServer(m)
}
