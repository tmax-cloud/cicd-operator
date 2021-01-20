package controllers

import (
	"context"
	"fmt"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/typed/networking/v1beta1"
	"k8s.io/kubernetes/pkg/apis/core"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses/status,verbs=get;update;patch

func WaitIngressReady() error {
	log := ctrl.Log.WithName("ingress-controller")

	ingCli, err := newIngressClient()
	if err != nil {
		return err
	}

	watcher, err := ingCli.Watch(context.Background(), metav1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(core.ObjectNameField, "cicd-webhook").String(),
	})
	if err != nil {
		return err
	}

	for result := range watcher.ResultChan() {
		ing, ok := result.Object.(*networkingv1beta1.Ingress)
		if !ok {
			return fmt.Errorf("watch result is not an ingress")
		}

		log.Info(ing.Name)

		// Check if class is set properly
		class := ing.Annotations[networkingv1beta1.AnnotationIngressClass]
		if class != configs.IngressClass {
			log.Info("Updating ingress with proper class...")
			ing.Annotations[networkingv1beta1.AnnotationIngressClass] = configs.IngressClass
			if _, err := ingCli.Update(context.Background(), ing, metav1.UpdateOptions{}); err != nil {
				return err
			}
			continue
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
				return nil
			}

			// If not, set it!
			hostname := configs.ExternalHostName
			if configs.ExternalHostName == "" {
				hostname = fmt.Sprintf("cicd-webhook.%s.nip.io", ip)
			}
			ing.Spec.Rules[0].Host = hostname

			if _, err := ingCli.Update(context.Background(), ing, metav1.UpdateOptions{}); err != nil {
				return err
			}

			if configs.ExternalHostName == "" {
				configs.ExternalHostName = hostname
			}

			return nil
		}
	}

	return fmt.Errorf("cannot wait ingress ready")
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
