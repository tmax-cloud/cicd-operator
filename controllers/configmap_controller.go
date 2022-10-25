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
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

// ConfigReconciler reconciles a Approval object
type ConfigReconciler interface {
	Start()
	Add(name string, handler configs.Handler)
}

type configReconciler struct {
	client typedcorev1.ConfigMapInterface
	log    logr.Logger
	config *rest.Config

	lastResourceVersions map[string]string
	handlers             map[string]configs.Handler
}

// NewConfigReconciler creates a new configReconciler
func NewConfigReconciler(config *rest.Config) (ConfigReconciler, error) {
	r := &configReconciler{
		log:                  ctrl.Log.WithName("controllers").WithName("ConfigController"),
		config:               config,
		lastResourceVersions: map[string]string{},
		handlers:             map[string]configs.Handler{},
	}

	var err error
	r.client, err = newConfigMapClient(r.config)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// Start starts the config map reconciler
func (r *configReconciler) Start() {
	for {
		// Watch & reconcile ConfigMaps
		watcher, err := r.client.Watch(context.Background(), metav1.ListOptions{})
		if err != nil {
			r.log.Error(err, "")
			continue
		}

		for ev := range watcher.ResultChan() {
			cm, ok := ev.Object.(*corev1.ConfigMap)
			if ok {
				if err := r.Reconcile(cm); err != nil {
					r.log.Error(err, "")
				}
			}
		}
	}
}

// Add adds config map handler
func (r *configReconciler) Add(name string, handler configs.Handler) {
	r.handlers[name] = handler
}

// Reconcile reconciles ConfigMap
func (r *configReconciler) Reconcile(cm *corev1.ConfigMap) error {
	if cm == nil {
		return nil
	}

	// Check resource versions
	if cm.ResourceVersion == r.lastResourceVersions[cm.Name] {
		return nil
	}
	r.lastResourceVersions[cm.Name] = cm.ResourceVersion

	handler, exist := r.handlers[cm.Name]
	if !exist {
		return nil
	}
	r.log.Info(fmt.Sprintf("Configmap (%s) is changed", cm.Name))
	if err := handler(cm); err != nil {
		return err
	}
	return nil
}

func newConfigMapClient(conf *rest.Config) (typedcorev1.ConfigMapInterface, error) {
	clientSet, err := kubernetes.NewForConfig(conf)
	if err != nil {
		return nil, err
	}

	return clientSet.CoreV1().ConfigMaps(utils.Namespace()), nil
}
