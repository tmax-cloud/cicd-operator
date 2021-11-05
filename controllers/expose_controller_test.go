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

package controllers

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/internal/test"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

func TestNewExposeController(t *testing.T) {
	tc := map[string]struct {
		cfg *rest.Config

		errorOccurs  bool
		errorMessage string
	}{
		"hostErr": {
			cfg:          &rest.Config{Host: "//"},
			errorOccurs:  true,
			errorMessage: "host must be a URL or a host:port pair: \"//\"",
		},
		"normal": {
			cfg: &rest.Config{},
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			ctr, err := NewExposeController(c.cfg)
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, c.cfg, ctr.cfg)
				require.Len(t, ctr.reconcilers, 2)
			}
		})
	}
}

func TestExposeController_Start(t *testing.T) {
	client := &fakeRestClient{
		resources: map[string]runtime.Object{
			"test": &corev1.Pod{},
		},
		fakeWatcher: watch.NewFake(),
	}
	reconciler := &fakeExposeReconciler{
		client:       client,
		resourceName: "test",
		refObj:       &corev1.Pod{},
	}
	logger := &test.FakeLogger{}
	controller := &exposeController{
		reconcilers:         []exposeReconciler{reconciler},
		log:                 logger,
		updateExternalURLCh: make(chan string),
	}

	exit := make(chan struct{})

	go func() {
		controller.updateExternalURLCh <- "test"
		exit <- struct{}{}
	}()

	controller.Start(exit)
}

func TestExposeController_updateExternalURL(t *testing.T) {
	tc := map[string]struct {
		hostname         string
		externalHostName string

		currentHostName string
	}{
		"blank": {},
		"set": {
			hostname:        "test.host.name",
			currentHostName: "test.host.name",
		},
		"externalSet": {
			hostname:         "test.host.name",
			externalHostName: "desired.host.name",
			currentHostName:  "desired.host.name",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			configs.ExternalHostName = c.externalHostName
			configs.CurrentExternalHostName = ""

			logger := &test.FakeLogger{}
			controller := &exposeController{
				log: logger,
			}

			controller.updateExternalURL(c.hostname)

			require.Equal(t, c.currentHostName, configs.CurrentExternalHostName)
		})
	}
}

func TestExposeController_startResourceWatcher(t *testing.T) {
	exitCh := make(chan struct{})

	logger := &test.FakeLogger{}
	controller := &exposeController{
		log: logger,
	}
	client := &fakeRestClient{
		resources: map[string]runtime.Object{
			"test": &corev1.Pod{},
		},
		fakeWatcher: watch.NewFake(),
	}
	reconciler := &fakeExposeReconciler{
		client:       client,
		resourceName: "test",
		refObj:       &corev1.Pod{},
	}

	go func() {
		client.fakeWatcher.Add(&corev1.Pod{})
		exitCh <- struct{}{}
	}()

	controller.startResourceWatcher(reconciler, exitCh)
}

func TestExposeController_loopWatch(t *testing.T) {
	tc := map[string]struct {
		goFunc    func(client *fakeRestClient, rscUpdateCh chan runtime.Object, exitCh chan struct{})
		noWatcher bool

		expectedReturn bool
	}{
		"invalidClient": {
			goFunc: func(client *fakeRestClient, rscUpdateCh chan runtime.Object, exitCh chan struct{}) {
				exitCh <- struct{}{}
			},
			noWatcher:      true,
			expectedReturn: false,
		},
		"exitNil": {
			goFunc: func(client *fakeRestClient, rscUpdateCh chan runtime.Object, exitCh chan struct{}) {
				client.fakeWatcher.Add(nil)
			},
			expectedReturn: false,
		},
		"exitTrue": {
			goFunc: func(client *fakeRestClient, rscUpdateCh chan runtime.Object, exitCh chan struct{}) {
				client.fakeWatcher.Add(&corev1.Pod{})
				<-rscUpdateCh
				exitCh <- struct{}{}
			},
			expectedReturn: true,
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			exitCh := make(chan struct{})
			rscUpdateCh := make(chan runtime.Object)

			logger := &test.FakeLogger{}
			controller := &exposeController{
				log: logger,
			}
			client := &fakeRestClient{
				resources: map[string]runtime.Object{
					"test": &corev1.Pod{},
				},
				fakeWatcher: watch.NewFake(),
			}

			if c.noWatcher {
				client.fakeWatcher = nil
			}

			go c.goFunc(client, rscUpdateCh, exitCh)
			if c.expectedReturn {
				require.True(t, controller.loopWatch(client, "", rscUpdateCh, exitCh))
			} else {
				require.False(t, controller.loopWatch(client, "", rscUpdateCh, exitCh))
			}
		})
	}
}

type fakeExposeReconciler struct {
	client       utils.RestClient
	resourceName string
	refObj       runtime.Object

	savedObjects []runtime.Object
}

func (f *fakeExposeReconciler) reconcile(object runtime.Object, _ exposeType) error {
	if object == nil {
		return fmt.Errorf("object cannot be nil")
	}
	f.savedObjects = append(f.savedObjects, object)
	return nil
}
func (f *fakeExposeReconciler) getClient() utils.RestClient  { return f.client }
func (f *fakeExposeReconciler) getResourceName() string      { return f.resourceName }
func (f *fakeExposeReconciler) getRefObject() runtime.Object { return f.refObj }

func TestExposeController_watchResource(t *testing.T) {
	tc := map[string]struct {
		signaller  func(chan runtime.Object, chan struct{})
		exposeMode string

		errors  []error
		objects []runtime.Object
	}{
		"pod": {
			exposeMode: "Ingress",
			signaller: func(objects chan runtime.Object, c chan struct{}) {
				objects <- &corev1.Pod{ObjectMeta: metav1.ObjectMeta{ResourceVersion: "version"}}
			},
			objects: []runtime.Object{&corev1.Pod{ObjectMeta: metav1.ObjectMeta{ResourceVersion: "version"}}},
		},
		"podNoResourceVersion": {
			exposeMode: "Ingress",
			signaller: func(objects chan runtime.Object, c chan struct{}) {
				objects <- &corev1.Pod{}
			},
		},
		"nil": {
			exposeMode: "Ingress",
			signaller: func(objects chan runtime.Object, c chan struct{}) {
				objects <- nil
			},
			errors: []error{fmt.Errorf("object does not implement the Object interfaces")},
		},
		"cfg": {
			exposeMode: "Ingress",
			signaller: func(objects chan runtime.Object, c chan struct{}) {
				c <- struct{}{}
			},
		},
		"blankExposeMode": {
			exposeMode: "",
			signaller: func(objects chan runtime.Object, c chan struct{}) {
				c <- struct{}{}
			},
		},
		"invalidExposeMode": {
			exposeMode: "NodePort",
			signaller: func(objects chan runtime.Object, c chan struct{}) {
				objects <- &corev1.Pod{ObjectMeta: metav1.ObjectMeta{ResourceVersion: "version"}}
			},
			errors: []error{fmt.Errorf("exposeMode NodePort is not valid")},
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			configs.ExposeMode = c.exposeMode
			rscUpdateCh := make(chan runtime.Object)
			cfgUpdateCh := make(chan struct{})
			done := make(chan struct{})

			go func() {
				c.signaller(rscUpdateCh, cfgUpdateCh)
				done <- struct{}{}
			}()

			logger := &test.FakeLogger{}
			controller := &exposeController{
				log: logger,
			}
			reconciler := &fakeExposeReconciler{}
			controller.watchResource(rscUpdateCh, cfgUpdateCh, done, reconciler)

			require.Len(t, logger.Infos, 0)
			require.Len(t, logger.Errors, len(logger.ErrorMsgs))

			require.Equal(t, c.errors, logger.Errors)
			require.Equal(t, c.objects, reconciler.savedObjects)
		})
	}
}

func TestExposeController_validateExposeMode(t *testing.T) {
	logger := &test.FakeLogger{}
	controller := &exposeController{
		log: logger,
	}

	tc := map[string]struct {
		mode         exposeType
		errorOccurs  bool
		errorMessage string
	}{
		"Ingress": {
			mode: exposeTypeIngress,
		},
		"LoadBalancer": {
			mode: exposeTypeLoadBalancer,
		},
		"ClusterIP": {
			mode: exposeTypeClusterIP,
		},
		"invalid": {
			mode:         exposeType("hi"),
			errorOccurs:  true,
			errorMessage: "exposeMode hi is not valid",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			err := controller.validateExposeMode(c.mode)
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_exposeIngressReconciler_getClient(t *testing.T) {
	reconciler := &exposeIngressReconciler{
		client: &fakeRestClient{
			resources:   map[string]runtime.Object{},
			fakeWatcher: watch.NewFake(),
		},
	}
	require.Equal(t, reconciler.client, reconciler.getClient())
}

func Test_exposeIngressReconciler_getResourceName(t *testing.T) {
	reconciler := &exposeIngressReconciler{
		resourceName: "hi",
	}
	require.Equal(t, reconciler.resourceName, reconciler.getResourceName())
}

func Test_exposeIngressReconciler_getRefObject(t *testing.T) {
	reconciler := &exposeIngressReconciler{
		obj: &corev1.Service{Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeClusterIP}},
	}
	require.Equal(t, reconciler.obj, reconciler.getRefObject())
}

func Test_exposeIngressReconciler_reconcile(t *testing.T) {
	strNginx := "nginx"

	tc := map[string]struct {
		exposeMode   string
		ingressHost  string
		ingressClass string
		obj          runtime.Object

		errorOccurs  bool
		errorMessage string
		configHost   bool
		expectedHost string
		expectedObj  runtime.Object
	}{
		"nilObj": {
			ingressClass: strNginx,
			obj:          nil,
			errorOccurs:  true,
			errorMessage: "obj is not an Ingress",
		},
		"notIngress": {
			ingressClass: strNginx,
			obj:          &corev1.Pod{},
			errorOccurs:  true,
			errorMessage: "obj is not an Ingress",
		},
		"noRules": {
			ingressClass: strNginx,
			obj:          &networkingv1.Ingress{Spec: networkingv1.IngressSpec{Rules: []networkingv1.IngressRule{}}},
			errorOccurs:  true,
			errorMessage: "rules for ingress are not set",
		},
		"noUpdate": {
			exposeMode:   "LoadBalancer",
			ingressClass: strNginx,
			obj:          &networkingv1.Ingress{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{}}, Spec: networkingv1.IngressSpec{Rules: []networkingv1.IngressRule{{}}}},
			expectedObj:  &networkingv1.Ingress{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{"kubernetes.io/ingress.class": "nginx"}}, Spec: networkingv1.IngressSpec{IngressClassName: &strNginx, Rules: []networkingv1.IngressRule{{}}}},
		},
		"emptyClass": {
			exposeMode:  "LoadBalancer",
			obj:         &networkingv1.Ingress{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{}}, Spec: networkingv1.IngressSpec{Rules: []networkingv1.IngressRule{{}}}},
			expectedObj: &networkingv1.Ingress{ObjectMeta: metav1.ObjectMeta{}, Spec: networkingv1.IngressSpec{Rules: []networkingv1.IngressRule{{}}}},
		},
		"setHost": {
			exposeMode:   "Ingress",
			ingressHost:  "host.ingress.com",
			ingressClass: strNginx,
			obj:          &networkingv1.Ingress{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{}}, Spec: networkingv1.IngressSpec{Rules: []networkingv1.IngressRule{{}}}},
			expectedObj:  &networkingv1.Ingress{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{"kubernetes.io/ingress.class": "nginx"}}, Spec: networkingv1.IngressSpec{IngressClassName: &strNginx, Rules: []networkingv1.IngressRule{{Host: "host.ingress.com"}}}},
			configHost:   true,
			expectedHost: "host.ingress.com",
		},
		"setDefaultHost": {
			exposeMode:   "Ingress",
			ingressClass: strNginx,
			obj:          &networkingv1.Ingress{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{}}, Spec: networkingv1.IngressSpec{Rules: []networkingv1.IngressRule{{}}}, Status: networkingv1.IngressStatus{LoadBalancer: corev1.LoadBalancerStatus{Ingress: []corev1.LoadBalancerIngress{{IP: "172.22.11.11"}}}}},
			expectedObj:  &networkingv1.Ingress{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{"kubernetes.io/ingress.class": "nginx"}}, Spec: networkingv1.IngressSpec{IngressClassName: &strNginx, Rules: []networkingv1.IngressRule{{Host: "cicd-webhook.172.22.11.11.nip.io"}}}, Status: networkingv1.IngressStatus{LoadBalancer: corev1.LoadBalancerStatus{Ingress: []corev1.LoadBalancerIngress{{IP: "172.22.11.11"}}}}},
			configHost:   true,
			expectedHost: "cicd-webhook.172.22.11.11.nip.io",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			configs.IngressHost = c.ingressHost
			configs.IngressClass = c.ingressClass

			reconciler := &exposeIngressReconciler{
				log: ctrl.Log.WithName(""),
				client: &fakeRestClient{
					resources:   map[string]runtime.Object{},
					fakeWatcher: watch.NewFake(),
				},
				hostUpdateCh: make(chan string),
			}

			go func() {
				w, _ := reconciler.client.Watch(nil)
				for range w.ResultChan() {
				}
			}()

			proxyCh := make(chan string)
			go func() {
				proxyCh <- <-reconciler.hostUpdateCh
			}()

			if c.obj != nil {
				require.NoError(t, reconciler.client.Create(c.obj, nil))
			}

			err := reconciler.reconcile(c.obj, exposeType(c.exposeMode))
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
				if c.configHost {
					require.Equal(t, c.expectedHost, <-proxyCh)
				}
				objMeta, err := apimeta.Accessor(c.obj)
				require.NoError(t, err)
				into := c.expectedObj.DeepCopyObject()
				require.NoError(t, reconciler.client.Get(objMeta.GetName(), nil, into))
				require.Equal(t, c.expectedObj, into)
			}
		})
	}
}

func Test_exposeServiceReconciler_getClient(t *testing.T) {
	reconciler := &exposeServiceReconciler{
		client: &fakeRestClient{
			resources:   map[string]runtime.Object{},
			fakeWatcher: watch.NewFake(),
		},
	}
	require.Equal(t, reconciler.client, reconciler.getClient())
}

func Test_exposeServiceReconciler_getResourceName(t *testing.T) {
	reconciler := &exposeServiceReconciler{
		resourceName: "hi",
	}
	require.Equal(t, reconciler.resourceName, reconciler.getResourceName())
}

func Test_exposeServiceReconciler_getRefObject(t *testing.T) {
	reconciler := &exposeServiceReconciler{
		obj: &corev1.Service{Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeClusterIP}},
	}
	require.Equal(t, reconciler.obj, reconciler.getRefObject())
}

func Test_exposeServiceReconciler_reconcile(t *testing.T) {
	tc := map[string]struct {
		exposeMode string
		obj        runtime.Object

		errorOccurs  bool
		errorMessage string
		configHost   bool
		expectedHost string
		expectedObj  runtime.Object
	}{
		"nilObj": {
			obj:          nil,
			errorOccurs:  true,
			errorMessage: "svc is not a Service",
		},
		"notService": {
			obj:          &corev1.Pod{},
			errorOccurs:  true,
			errorMessage: "svc is not a Service",
		},
		"noop": {
			exposeMode:  "Ingress",
			obj:         &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "svc1"}, Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeClusterIP, ClusterIP: "10.0.0.3"}},
			expectedObj: &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "svc1"}, Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeClusterIP, ClusterIP: "10.0.0.3"}},
		},
		"toClusterIP": {
			exposeMode:   "ClusterIP",
			obj:          &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "svc1"}, Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeLoadBalancer, ClusterIP: "10.0.0.3", Ports: []corev1.ServicePort{{NodePort: 32008}}}},
			expectedObj:  &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "svc1"}, Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeClusterIP, ClusterIP: "10.0.0.3", Ports: []corev1.ServicePort{{}}}},
			configHost:   true,
			expectedHost: "10.0.0.3:24335",
		},
		"toLoadBalancer": {
			exposeMode:   "LoadBalancer",
			obj:          &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "svc1"}, Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeClusterIP, ClusterIP: "10.0.0.3"}, Status: corev1.ServiceStatus{LoadBalancer: corev1.LoadBalancerStatus{Ingress: []corev1.LoadBalancerIngress{{IP: "172.22.11.11"}}}}},
			expectedObj:  &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "svc1"}, Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeLoadBalancer, ClusterIP: "10.0.0.3"}, Status: corev1.ServiceStatus{LoadBalancer: corev1.LoadBalancerStatus{Ingress: []corev1.LoadBalancerIngress{{IP: "172.22.11.11"}}}}},
			configHost:   true,
			expectedHost: "172.22.11.11:24335",
		},
		"toLoadBalancerError": {
			exposeMode:   "LoadBalancer",
			obj:          &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "svc1"}, Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeClusterIP, ClusterIP: "10.0.0.3"}},
			expectedObj:  &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "svc1"}, Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeLoadBalancer, ClusterIP: "10.0.0.3"}},
			errorOccurs:  true,
			errorMessage: "ExposeMode is set as LoadBalancer but IP is not set to the service",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			reconciler := &exposeServiceReconciler{
				log: ctrl.Log.WithName(""),
				client: &fakeRestClient{
					resources:   map[string]runtime.Object{},
					fakeWatcher: watch.NewFake(),
				},
				hostUpdateCh: make(chan string),
			}

			go func() {
				w, _ := reconciler.client.Watch(nil)
				for range w.ResultChan() {
				}
			}()

			proxyCh := make(chan string)
			go func() {
				proxyCh <- <-reconciler.hostUpdateCh
			}()

			if c.obj != nil {
				require.NoError(t, reconciler.client.Create(c.obj, nil))
			}

			err := reconciler.reconcile(c.obj, exposeType(c.exposeMode))
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
				if c.configHost {
					require.Equal(t, c.expectedHost, <-proxyCh)
				}
				objMeta, err := apimeta.Accessor(c.obj)
				require.NoError(t, err)
				into := c.expectedObj.DeepCopyObject()
				require.NoError(t, reconciler.client.Get(objMeta.GetName(), nil, into))
				require.Equal(t, c.expectedObj, into)
			}
		})
	}
}

func Test_exposeServiceReconciler_configureService(t *testing.T) {
	s := &exposeServiceReconciler{}

	tc := map[string]struct {
		mode        exposeType
		svc         corev1.Service
		expectedSvc corev1.Service
	}{
		"noop": {
			mode:        exposeTypeIngress,
			svc:         corev1.Service{Spec: corev1.ServiceSpec{Type: "ClusterIP", ClusterIP: "10.0.0.3", Ports: []corev1.ServicePort{{}}}},
			expectedSvc: corev1.Service{Spec: corev1.ServiceSpec{Type: "ClusterIP", ClusterIP: "10.0.0.3", Ports: []corev1.ServicePort{{}}}},
		},
		"nodePortToClusterIP": {
			mode:        exposeTypeClusterIP,
			svc:         corev1.Service{Spec: corev1.ServiceSpec{Type: "NodePort", ClusterIP: "10.0.0.3", Ports: []corev1.ServicePort{{Port: 80, NodePort: 32008}}}},
			expectedSvc: corev1.Service{Spec: corev1.ServiceSpec{Type: "ClusterIP", ClusterIP: "10.0.0.3", Ports: []corev1.ServicePort{{Port: 80}}}},
		},
		"lbToClusterIP": {
			mode:        exposeTypeClusterIP,
			svc:         corev1.Service{Spec: corev1.ServiceSpec{Type: "LoadBalancer", ClusterIP: "10.0.0.3", Ports: []corev1.ServicePort{{Port: 80, NodePort: 32008}}}},
			expectedSvc: corev1.Service{Spec: corev1.ServiceSpec{Type: "ClusterIP", ClusterIP: "10.0.0.3", Ports: []corev1.ServicePort{{Port: 80}}}},
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			s.configureService(c.mode, &c.svc)
			require.Equal(t, c.expectedSvc, c.svc)
		})
	}
}

func Test_exposeServiceReconciler_generateExternalHost(t *testing.T) {
	s := &exposeServiceReconciler{}

	svc := &corev1.Service{
		Spec:   corev1.ServiceSpec{ClusterIP: "10.0.0.3", Ports: []corev1.ServicePort{{}}},
		Status: corev1.ServiceStatus{LoadBalancer: corev1.LoadBalancerStatus{Ingress: []corev1.LoadBalancerIngress{{IP: "172.22.11.11"}}}},
	}
	t.Run("ingress", func(t *testing.T) {
		host, err := s.generateExternalHost(exposeTypeIngress, svc)
		require.NoError(t, err)
		require.Equal(t, "", host)
	})
}

func Test_exposeServiceReconciler_findWebhookPort(t *testing.T) {
	s := &exposeServiceReconciler{}

	tc := map[string]struct {
		svc          *corev1.Service
		expectedPort int32
	}{
		"default": {
			svc:          &corev1.Service{Spec: corev1.ServiceSpec{Ports: []corev1.ServicePort{{}}}},
			expectedPort: 24335,
		},
		"diff": {
			svc:          &corev1.Service{Spec: corev1.ServiceSpec{Ports: []corev1.ServicePort{{Name: "webhook", Port: int32(80)}}}},
			expectedPort: 80,
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, c.expectedPort, s.findWebhookPort(c.svc))

		})
	}
}

type fakeRestClient struct {
	resources   map[string]runtime.Object
	fakeWatcher *watch.FakeWatcher
}

func (f *fakeRestClient) Get(name string, _ *metav1.GetOptions, into runtime.Object) error {
	rsc, exist := f.resources[name]
	if !exist {
		return fmt.Errorf("%s does not exist", name)
	}

	body, err := json.Marshal(rsc)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, into)
}

func (f *fakeRestClient) List(_ *metav1.ListOptions, _ runtime.Object) error {
	return nil
}

func (f *fakeRestClient) Watch(_ *metav1.ListOptions) (watch.Interface, error) {
	if f.fakeWatcher == nil {
		return nil, fmt.Errorf("no watcher")
	}
	return f.fakeWatcher, nil
}

func (f *fakeRestClient) Create(obj runtime.Object, _ *metav1.CreateOptions) error {
	objMeta, err := apimeta.Accessor(obj)
	if err != nil {
		return err
	}

	if _, exist := f.resources[objMeta.GetName()]; exist {
		return fmt.Errorf("%s already exist", objMeta.GetName())
	}
	f.resources[objMeta.GetName()] = obj
	f.fakeWatcher.Add(obj)

	return nil
}

func (f *fakeRestClient) Update(obj runtime.Object, _ *metav1.UpdateOptions) error {
	objMeta, err := apimeta.Accessor(obj)
	if err != nil {
		return err
	}

	if objMeta.GetResourceVersion() == "" {
		return fmt.Errorf("resourceVersion must be set for update")
	}

	if _, exist := f.resources[objMeta.GetName()]; !exist {
		return fmt.Errorf("%s does not exist", objMeta.GetName())
	}
	f.resources[objMeta.GetName()] = obj
	f.fakeWatcher.Modify(obj)

	return nil
}

func (f *fakeRestClient) Delete(name string, _ *metav1.DeleteOptions) error {
	obj, exist := f.resources[name]
	if !exist {
		return fmt.Errorf("%s does not exist", name)
	}

	delete(f.resources, name)
	f.fakeWatcher.Delete(obj)

	return nil
}
