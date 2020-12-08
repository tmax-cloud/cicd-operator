package controllers

import (
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

func WaitIngressReady() error {
	log := ctrl.Log.WithName("ingress-controller")

	ingCli, err := newClient()
	if err != nil {
		return err
	}

	watcher, err := ingCli.Watch(metav1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(core.ObjectNameField, "cicd-webhook").String(),
	})

	for result := range watcher.ResultChan() {
		ing, ok := result.Object.(*networkingv1beta1.Ingress)
		if !ok {
			return fmt.Errorf("watch result is not an ingress")
		}

		log.Info(ing.Name)

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
			// TODO make nip.io configurable
			hostname := fmt.Sprintf("cicd-webhook.%s.nip.io", ip)
			ing.Spec.Rules[0].Host = hostname

			if _, err := ingCli.Update(ing); err != nil {
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

func newClient() (v1beta1.IngressInterface, error) {
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
