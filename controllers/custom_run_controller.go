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

	"github.com/go-logr/logr"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// KindHandler is an interface of custom task handlers
type KindHandler interface {
	Handle(run *tektonv1alpha1.Run) (ctrl.Result, error)
}

// CustomRunReconciler reconciles a Run object, which refers to approval/email/...
type CustomRunReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	KindHandlerMap  map[string]KindHandler
	HandlerChildren map[string][]client.Object
}

// AddKindHandler registers custom task handlers
func (r *CustomRunReconciler) AddKindHandler(kind string, handler KindHandler, children ...client.Object) {
	r.KindHandlerMap[kind] = handler
	r.HandlerChildren[kind] = append(r.HandlerChildren[kind], children...)
}

// +kubebuilder:rbac:groups=tekton.dev,resources=runs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tekton.dev,resources=runs/status,verbs=get;update;patch

// Reconcile reconciles Run object
func (r *CustomRunReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("customRuns", req.NamespacedName)

	// Get Run
	run := &tektonv1alpha1.Run{}
	if err := r.Client.Get(ctx, req.NamespacedName, run); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "")
		return ctrl.Result{}, err
	}

	// Filter only cicd.tmax.io/v1
	if run.Spec.Ref == nil || run.Spec.Ref.APIVersion != cicdv1.CustomTaskAPIVersion {
		return ctrl.Result{}, nil
	}

	// Call custom handlers
	handler, exist := r.KindHandlerMap[string(run.Spec.Ref.Kind)]
	if exist {
		return handler.Handle(run)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets CustomRunReconciler to the manager
func (r *CustomRunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	builder := ctrl.NewControllerManagedBy(mgr).
		For(&tektonv1alpha1.Run{})

	// Set owner ref. handler for children
	for _, objects := range r.HandlerChildren {
		for _, o := range objects {
			builder.Owns(o)
		}
	}
	return builder.Complete(r)
}
