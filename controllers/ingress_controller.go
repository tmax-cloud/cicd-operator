package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/typed/networking/v1beta1"
	"k8s.io/kubernetes/pkg/apis/core"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	ingName = "cicd-webhook"
)

// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses/status,verbs=get;update;patch

// IngressController is an interface for ingress controller
type IngressController interface {
	Wait() error
	Start()
}

type ingressController struct {
	initiated           bool
	initCh              chan error
	lastResourceVersion string

	ingCli v1beta1.IngressInterface

	log logr.Logger
}

// NewIngressController instantiates ingressController
func NewIngressController() IngressController {
	return &ingressController{
		initiated:           false,
		initCh:              make(chan error),
		lastResourceVersion: "",
		log:                 ctrl.Log.WithName("ingress-controller"),
	}
}

func (i *ingressController) Wait() error {
	if i.initiated {
		return nil
	}
	return <-i.initCh
}

func (i *ingressController) Start() {
	// Initiate client
	ingCli, err := newIngressClient()
	if err != nil {
		i.log.Error(err, "")
		os.Exit(1)
	}
	i.ingCli = ingCli

	// Get first to check the Ingress's existence
	_, err = i.ingCli.Get(context.Background(), ingName, metav1.GetOptions{})
	if err != nil {
		i.log.Error(err, "")
		os.Exit(1)
	}

	for {
		i.watch()
	}
}

func (i *ingressController) watch() {
	watcher, err := i.ingCli.Watch(context.Background(), metav1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(core.ObjectNameField, ingName).String(),
	})
	if err != nil {
		i.log.Error(err, "")
		return
	}

	for result := range watcher.ResultChan() {
		ing, ok := result.Object.(*networkingv1beta1.Ingress)
		if ok && i.lastResourceVersion != ing.ResourceVersion {
			i.lastResourceVersion = ing.ResourceVersion
			if err := i.reconcile(ing); err != nil {
				i.log.Error(err, "")
			}
		}
	}
}

func (i *ingressController) reconcile(ing *networkingv1beta1.Ingress) error {
	i.log.Info(fmt.Sprintf("Reconciling ingress %s", ing.Name))
	// Check if class is set properly
	class := ing.Annotations[networkingv1beta1.AnnotationIngressClass]
	if class != configs.IngressClass {
		i.log.Info("Updating ingress with proper class...")
		ing.Annotations[networkingv1beta1.AnnotationIngressClass] = configs.IngressClass
		if _, err := i.ingCli.Update(context.Background(), ing, metav1.UpdateOptions{}); err != nil {
			return err
		}
		return nil
	}

	// IP is set
	if len(ing.Status.LoadBalancer.Ingress) > 0 && ing.Status.LoadBalancer.Ingress[0].IP != "" {
		ip := ing.Status.LoadBalancer.Ingress[0].IP

		// Check if host is already set
		if len(ing.Spec.Rules) == 0 {
			return fmt.Errorf("rules for ingress are not set")
		}

		// Host is already set
		if ing.Spec.Rules[0].Host != "waiting.for.loadbalancer" {
			if configs.ExternalHostName == "" {
				configs.ExternalHostName = ing.Spec.Rules[0].Host
			}

			i.log.Info("Current external hostname is " + configs.ExternalHostName)
			if !i.initiated {
				i.initiated = true
				i.initCh <- nil
			}
			return nil
		}

		// If not, set it!
		hostname := configs.IngressHost
		if hostname == "" {
			hostname = fmt.Sprintf("cicd-webhook.%s.nip.io", ip)
		}
		ing.Spec.Rules[0].Host = hostname

		if _, err := i.ingCli.Update(context.Background(), ing, metav1.UpdateOptions{}); err != nil {
			return err
		}

		if configs.ExternalHostName == "" {
			configs.ExternalHostName = hostname
		}

		i.log.Info("Current external hostname is " + configs.ExternalHostName)
		if !i.initiated {
			i.initiated = true
			i.initCh <- nil
		}
		return nil
	}

	return nil
}

func newIngressClient() (v1beta1.IngressInterface, error) {
	conf, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	clientSet, err := kubernetes.NewForConfig(conf)
	if err != nil {
		return nil, err
	}

	namespace, err := utils.Namespace()
	if err != nil {
		return nil, err
	}

	return clientSet.NetworkingV1beta1().Ingresses(namespace), nil
}
