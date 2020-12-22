/*


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
	"github.com/go-logr/logr"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tmax-cloud/cicd-operator/pkg/pipelinemanager"
	"github.com/tmax-cloud/cicd-operator/pkg/scheduler"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
)

// IntegrationJobReconciler reconciles a IntegrationJob object
type IntegrationJobReconciler struct {
	client.Client
	Log       logr.Logger
	Scheme    *runtime.Scheme
	Scheduler *scheduler.Scheduler
}

// +kubebuilder:rbac:groups=cicd.tmax.io,resources=integrationjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cicd.tmax.io,resources=integrationjobs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=tekton.dev,resources=pipelineruns,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tekton.dev,resources=pipelineruns/status,verbs=get

func (r *IntegrationJobReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("integrationjob", req.NamespacedName)

	// Get IntegrationJob
	instance := &cicdv1.IntegrationJob{}
	if err := r.Client.Get(ctx, req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		log.Error(err, "")
		return reconcile.Result{}, err
	}
	original := instance.DeepCopy()

	// Notify state change to scheduler
	defer r.Scheduler.Notify(instance)

	// Get parent IntegrationConfig
	config := &cicdv1.IntegrationConfig{}
	if err := r.Client.Get(ctx, types.NamespacedName{Name: instance.Spec.ConfigRef.Name, Namespace: instance.Namespace}, config); err != nil {
		log.Error(err, "")
		return ctrl.Result{}, err
	}

	// Get PipelineRun
	pr := &tektonv1beta1.PipelineRun{}
	if err := r.Client.Get(ctx, types.NamespacedName{Name: pipelinemanager.Name(instance), Namespace: instance.Namespace}, pr); err != nil {
		if !errors.IsNotFound(err) {
			log.Error(err, "")
			return ctrl.Result{}, err
		} else {
			pr = nil
		}
	}

	// Set default values for IntegrationJob.status
	if err := instance.Status.SetDefaults(); err != nil {
		log.Error(err, "")
		return ctrl.Result{}, err
	}

	// Check PipelineRun's status and update IntegrationJob's status
	if err := pipelinemanager.ReflectStatus(pr, instance, config, r.Client); err != nil {
		log.Error(err, "")
		return ctrl.Result{}, err
	}

	// Update IntegrationJob
	p := client.MergeFrom(original)
	if err := r.Client.Status().Patch(context.Background(), instance, p); err != nil {
		log.Error(err, "")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *IntegrationJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cicdv1.IntegrationJob{}).
		Owns(&tektonv1beta1.PipelineRun{}).
		Complete(r)
}
