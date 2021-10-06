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
	"os"

	"github.com/go-logr/logr"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

// ConfigReconciler reconciles a Approval object
type ConfigReconciler struct {
	client typedcorev1.ConfigMapInterface
	Log    logr.Logger
	Config *rest.Config

	lastResourceVersions map[string]string
	Handlers             map[string]configs.Handler
}

// Start starts the config map reconciler
func (r *ConfigReconciler) Start() {
	var err error
	r.client, err = newConfigMapClient(r.Config)
	if err != nil {
		r.Log.Error(err, "")
		os.Exit(1)
	}

	//Set last resource versions for each config map
	r.lastResourceVersions = map[string]string{}
	for {
		r.watch()
	}
}

// Add adds config map handler
func (r *ConfigReconciler) Add(name string, handler configs.Handler) {
	r.Handlers[name] = handler
}

func (r *ConfigReconciler) watch() {
	log := r.Log.WithName("config-controller")

	watcher, err := r.client.Watch(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Error(err, "")
		return
	}

	for ev := range watcher.ResultChan() {
		cm, ok := ev.Object.(*corev1.ConfigMap)
		if ok {
			if err := r.Reconcile(cm); err != nil {
				log.Error(err, "")
			}
		}
	}
}

// Reconcile reconciles ConfigMap
func (r *ConfigReconciler) Reconcile(cm *corev1.ConfigMap) error {
	if cm == nil {
		return nil
	}

	// Check resource versions
	if cm.ResourceVersion == r.lastResourceVersions[cm.Name] {
		return nil
	}
	r.lastResourceVersions[cm.Name] = cm.ResourceVersion

	handler, exist := r.Handlers[cm.Name]
	if !exist {
		return nil
	}
	r.Log.Info(fmt.Sprintf("Configmap (%s) is changed", cm.Name))
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

	namespace, err := utils.Namespace()
	if err != nil {
		return nil, err
	}

	return clientSet.CoreV1().ConfigMaps(namespace), nil
}
