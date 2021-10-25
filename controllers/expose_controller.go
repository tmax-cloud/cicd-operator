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
	"fmt"

	"github.com/go-logr/logr"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

type exposeType string

const (
	ingName = "cicd-webhook"
	svcName = "cicd-webhook"

	exposeTypeIngress      = exposeType("Ingress")
	exposeTypeLoadBalancer = exposeType(corev1.ServiceTypeLoadBalancer)
	exposeTypeClusterIP    = exposeType(corev1.ServiceTypeClusterIP)

	resourceIngress = "ingresses"
	resourceService = "services"
)

// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

// ExposeController controls webhook exposure configuration, using ingress and service
type ExposeController interface {
	Start()
}

type exposeController struct {
	cfg *rest.Config
	log logr.Logger

	reconcilers         []exposeReconciler
	updateExternalURLCh chan string
}

// NewExposeController creates a new ExposeController
func NewExposeController(cfg *rest.Config) (*exposeController, error) {
	namespace := utils.Namespace()

	controller := &exposeController{
		cfg:                 cfg,
		log:                 ctrl.Log.WithName("exposeController"),
		updateExternalURLCh: make(chan string),
	}

	// Add reconcilers
	// Service reconciler
	svcClient, err := utils.NewGroupVersionResourceClient(cfg, namespace, corev1.SchemeGroupVersion.WithResource(resourceService))
	if err != nil {
		return nil, err
	}
	controller.reconcilers = append(controller.reconcilers, &exposeServiceReconciler{
		client:       svcClient,
		log:          controller.log.WithName(resourceService),
		hostUpdateCh: controller.updateExternalURLCh,
		resourceName: svcName,
		obj:          &corev1.Service{},
	})

	// Ingress reconciler
	ingClient, err := utils.NewGroupVersionResourceClient(cfg, namespace, networkingv1.SchemeGroupVersion.WithResource(resourceIngress))
	if err != nil {
		return nil, err
	}
	controller.reconcilers = append(controller.reconcilers, &exposeIngressReconciler{
		client:       ingClient,
		log:          controller.log.WithName(resourceIngress),
		hostUpdateCh: controller.updateExternalURLCh,
		resourceName: ingName,
		obj:          &networkingv1.Ingress{},
	})

	return controller, nil
}

// Start starts to watch resources
func (e *exposeController) Start(exit chan struct{}) {
	// Start reconcilers
	for _, rec := range e.reconcilers {
		go e.startResourceWatcher(rec, exit)
	}

	// Listen to externalURL update requests
	go func() {
		for host := range e.updateExternalURLCh {
			e.updateExternalURL(host)
		}
	}()
}

func (e *exposeController) updateExternalURL(hostname string) {
	if hostname == "" {
		return
	}

	// Respect external hostname configuration
	if configs.ExternalHostName != "" {
		configs.CurrentExternalHostName = configs.ExternalHostName
	} else {
		configs.CurrentExternalHostName = hostname
	}

	e.log.Info(fmt.Sprintf("Current external hostname is %s", configs.CurrentExternalHostName))
}

type exposeReconciler interface {
	reconcile(object runtime.Object, mode exposeType) error

	getClient() utils.RestClient
	getResourceName() string
	getRefObject() runtime.Object
}

func (e *exposeController) startResourceWatcher(reconciler exposeReconciler, exit chan struct{}) {
	// Make & register config update channel
	cfgUpdateCh := make(chan struct{}, 1)
	configs.RegisterControllerConfigUpdateChan(cfgUpdateCh)

	gvrCli := reconciler.getClient()
	name := reconciler.getResourceName()
	obj := reconciler.getRefObject()

	// Initial Get - check existence of the resource
	utilruntime.Must(gvrCli.Get(name, &metav1.GetOptions{}, obj))

	// Watch now
	rscUpdateCh := make(chan runtime.Object)
	go func() {
		e.watchResource(rscUpdateCh, cfgUpdateCh, nil, reconciler)
	}()

	// Infinite watch (watch ends if there is no event on the resource in a specific amount of time)
	for {
		if doExit := e.loopWatch(gvrCli, name, rscUpdateCh, exit); doExit {
			return
		}
	}
}

func (e *exposeController) loopWatch(gvrCli utils.RestClient, name string, rscUpdateCh chan runtime.Object, exitCh chan struct{}) bool {
	watcher, err := gvrCli.Watch(&metav1.ListOptions{FieldSelector: fields.OneTermEqualSelector(metav1.ObjectNameField, name).String()})
	if err != nil {
		e.log.Error(err, "")
		return false
	}
	for {
		select {
		case result := <-watcher.ResultChan():
			if result.Object == nil {
				return false
			}
			rscUpdateCh <- result.Object
		case <-exitCh:
			return true
		}
	}
}

func (e *exposeController) watchResource(rscUpdateCh chan runtime.Object, cfgUpdateCh, done chan struct{}, reconciler exposeReconciler) {
	lastResourceVersion := ""
	var lastResource runtime.Object
	for {
		select {
		case obj := <-rscUpdateCh:
			meta, err := apimeta.Accessor(obj)
			if err != nil {
				e.log.Error(err, "")
				continue
			}
			if lastResourceVersion == meta.GetResourceVersion() {
				continue
			}
			lastResourceVersion = meta.GetResourceVersion()
			lastResource = obj
		case <-cfgUpdateCh:
			// Do nothing! lastResource is handed to the reconciler
		case <-done:
			return
		}

		if lastResource == nil {
			continue
		}

		// Get expose mode
		exposeMode := exposeType(configs.ExposeMode)
		if exposeMode == "" {
			exposeMode = exposeTypeIngress
		}

		if err := e.validateExposeMode(exposeMode); err != nil {
			e.log.Error(err, "")
			continue
		}

		// Reconcile svc & ingress
		if err := reconciler.reconcile(lastResource, exposeMode); err != nil {
			e.log.Error(err, "")
		}
	}
}

func (e *exposeController) validateExposeMode(mode exposeType) error {
	switch mode {
	case exposeTypeClusterIP, exposeTypeLoadBalancer, exposeTypeIngress:
		// These are valid values
		return nil
	default:
		return fmt.Errorf("exposeMode %s is not valid", mode)
	}
}

type exposeIngressReconciler struct {
	log    logr.Logger
	client utils.RestClient

	hostUpdateCh chan string

	resourceName string
	obj          runtime.Object
}

func (i *exposeIngressReconciler) getClient() utils.RestClient {
	return i.client
}

func (i *exposeIngressReconciler) getResourceName() string {
	return i.resourceName
}

func (i *exposeIngressReconciler) getRefObject() runtime.Object {
	return i.obj
}

func (i *exposeIngressReconciler) reconcile(obj runtime.Object, exposeMode exposeType) error {
	ing, ok := obj.(*networkingv1.Ingress)
	if !ok {
		return fmt.Errorf("obj is not an Ingress")
	}

	i.log.Info(fmt.Sprintf("Reconciling ingress %s, desired mode: %s", ing.Name, exposeMode))
	if len(ing.Spec.Rules) == 0 {
		return fmt.Errorf("rules for ingress are not set")
	}

	defer func() {
		if exposeMode == exposeTypeIngress {
			i.hostUpdateCh <- ing.Spec.Rules[0].Host
		}
		if err := i.client.Update(ing, &metav1.UpdateOptions{}); err != nil {
			i.log.Error(err, "")
		}
	}()

	// Check if class is set properly
	if configs.IngressClass != "" {
		ing.Annotations["kubernetes.io/ingress.class"] = configs.IngressClass
		ing.Spec.IngressClassName = &configs.IngressClass
	}

	// Check if desired host is set properly
	if configs.IngressHost != "" {
		ing.Spec.Rules[0].Host = configs.IngressHost
		return nil
	}

	// Default ingress host (*.nip.io) only if IP is set
	if len(ing.Status.LoadBalancer.Ingress) > 0 && ing.Status.LoadBalancer.Ingress[0].IP != "" {
		ing.Spec.Rules[0].Host = fmt.Sprintf("cicd-webhook.%s.nip.io", ing.Status.LoadBalancer.Ingress[0].IP)
	}

	return nil
}

type exposeServiceReconciler struct {
	log    logr.Logger
	client utils.RestClient

	hostUpdateCh chan string

	resourceName string
	obj          runtime.Object
}

func (s *exposeServiceReconciler) getClient() utils.RestClient {
	return s.client
}

func (s *exposeServiceReconciler) getResourceName() string {
	return s.resourceName
}

func (s *exposeServiceReconciler) getRefObject() runtime.Object {
	return s.obj
}

func (s *exposeServiceReconciler) reconcile(obj runtime.Object, exposeMode exposeType) error {
	svc, ok := obj.(*corev1.Service)
	if !ok {
		return fmt.Errorf("svc is not a Service")
	}

	s.log.Info(fmt.Sprintf("Reconciling service %s, desired mode: %s", svc.Name, exposeMode))

	defer func(svc *corev1.Service) {
		if err := s.client.Update(svc, &metav1.UpdateOptions{}); err != nil {
			s.log.Error(err, "")
		}
	}(svc)

	// Configure Service properly
	s.configureService(exposeMode, svc)

	if exposeMode == exposeTypeIngress {
		// Do nothing! expose host is configured by ingress controller
		return nil
	}

	// Get external host
	externalHost, err := s.generateExternalHost(exposeMode, svc)
	if err != nil {
		return err
	}
	s.hostUpdateCh <- externalHost

	return nil
}

func (s *exposeServiceReconciler) configureService(mode exposeType, svc *corev1.Service) {
	// Ser Service type if not configured properly
	desiredType := corev1.ServiceType(mode)
	if mode == exposeTypeIngress {
		desiredType = corev1.ServiceTypeClusterIP
	}
	svc.Spec.Type = desiredType

	// If ClusterIP, remove node port
	if desiredType == corev1.ServiceTypeClusterIP {
		for i := range svc.Spec.Ports {
			svc.Spec.Ports[i].NodePort = 0
		}
	}
}

func (s *exposeServiceReconciler) generateExternalHost(mode exposeType, svc *corev1.Service) (string, error) {
	externalHost := ""
	webhookPort := s.findWebhookPort(svc)
	switch mode {
	case exposeTypeClusterIP:
		externalHost = fmt.Sprintf("%s:%d", svc.Spec.ClusterIP, webhookPort)
	case exposeTypeLoadBalancer:
		if len(svc.Status.LoadBalancer.Ingress) == 0 {
			return "", fmt.Errorf("ExposeMode is set as LoadBalancer but IP is not set to the service")
		}
		externalHost = fmt.Sprintf("%s:%d", svc.Status.LoadBalancer.Ingress[0].IP, webhookPort)
	}
	return externalHost, nil
}

func (s *exposeServiceReconciler) findWebhookPort(svc *corev1.Service) int32 {
	webhookPort := int32(24335)
	for _, p := range svc.Spec.Ports {
		if p.Name == "webhook" {
			webhookPort = p.Port
			break
		}
	}
	return webhookPort
}
