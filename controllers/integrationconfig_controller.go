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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/go-logr/logr"
	"github.com/operator-framework/operator-lib/status"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
)

const (
	finalizer         = "cicd.tmax.io/finalizer"
	gitSecretHostKey  = "tekton.dev/git-0"
	gitSecretUserName = "tmax-cicd-bot"
)

// IntegrationConfigReconciler reconciles a IntegrationConfig object
type IntegrationConfigReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=cicd.tmax.io,resources=integrationconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cicd.tmax.io,resources=integrationconfigs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=secrets;serviceaccounts,verbs=get;list;watch;create;update;patch;delete

// Reconcile reconciles IntegrationConfig
func (r *IntegrationConfigReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("integrationconfig", req.NamespacedName)

	instance := &cicdv1.IntegrationConfig{}
	if err := r.Client.Get(ctx, req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "")
		return ctrl.Result{}, err
	}
	original := instance.DeepCopy()

	// New Condition default
	cond := instance.Status.Conditions.GetCondition(cicdv1.IntegrationConfigConditionReady)
	if cond == nil {
		cond = &status.Condition{
			Type:   cicdv1.IntegrationConfigConditionReady,
			Status: corev1.ConditionFalse,
		}
	}

	defer func() {
		instance.Status.Conditions.SetCondition(*cond)
		p := client.MergeFrom(original)
		if err := r.Client.Status().Patch(ctx, instance, p); err != nil {
			log.Error(err, "")
		}
	}()

	exit, err := r.handleFinalizer(instance, original)
	if err != nil {
		log.Error(err, "")
		cond.Reason = "CannotHandleFinalizer"
		cond.Message = err.Error()
		return ctrl.Result{}, nil
	}
	if exit {
		return ctrl.Result{}, nil
	}

	// Set secret
	secretChanged := r.setSecretString(instance)

	// Set webhook registered
	webhookConditionChanged := r.setWebhookRegisteredCond(instance)

	// Set ready
	readyConditionChanged := r.setReadyCond(instance)

	// Git credential secret - referred by tekton
	if err := r.createGitSecret(instance); err != nil {
		log.Error(err, "")
		cond.Reason = "CannotCreateSecret"
		cond.Message = err.Error()
		return ctrl.Result{}, nil
	}

	// Service account for running PipelineRuns
	if err := r.createServiceAccount(instance); err != nil {
		log.Error(err, "")
		cond.Reason = "CannotCreateAccount"
		cond.Message = err.Error()
		return ctrl.Result{}, nil
	}

	// If conditions changed, update status
	if secretChanged || webhookConditionChanged || readyConditionChanged {
		p := client.MergeFrom(original)
		if err := r.Client.Status().Patch(ctx, instance, p); err != nil {
			log.Error(err, "")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets IntegrationConfigReconciler to the manager
func (r *IntegrationConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cicdv1.IntegrationConfig{}).
		Complete(r)
}

func (r *IntegrationConfigReconciler) handleFinalizer(instance, original *cicdv1.IntegrationConfig) (bool, error) {
	// Check first if finalizer is already set
	found := false
	idx := -1
	for i, f := range instance.Finalizers {
		if f == finalizer {
			found = true
			idx = i
			break
		}
	}
	if !found {
		instance.Finalizers = append(instance.Finalizers, finalizer)
		p := client.MergeFrom(original)
		if err := r.Client.Patch(context.Background(), instance, p); err != nil {
			return false, err
		}
		return true, nil
	}

	// Deletion check-up
	if instance.DeletionTimestamp != nil && idx >= 0 {
		// Delete webhook only if it has git token
		if instance.Spec.Git.Token != nil {
			gitCli, err := utils.GetGitCli(instance, r.Client)
			if err != nil {
				r.Log.Error(err, "")
			} else {
				hookList, err := gitCli.ListWebhook()
				if err != nil {
					r.Log.Error(err, "")
				}
				for _, h := range hookList {
					if h.URL == instance.GetWebhookServerAddress() {
						r.Log.Info("Deleting webhook " + h.URL)
						if err := gitCli.DeleteWebhook(h.ID); err != nil {
							r.Log.Error(err, "")
						}
					}
				}
			}
		}

		// Delete finalizer
		if len(instance.Finalizers) == 1 {
			instance.Finalizers = nil
		} else {
			last := len(instance.Finalizers) - 1
			instance.Finalizers[idx] = instance.Finalizers[last]
			instance.Finalizers[last] = ""
			instance.Finalizers = instance.Finalizers[:last]
		}

		p := client.MergeFrom(original)
		if err := r.Client.Patch(context.Background(), instance, p); err != nil {
			return false, err
		}

		return true, nil
	}

	return false, nil
}

// Set status.secrets, return if it's changed or not
func (r *IntegrationConfigReconciler) setSecretString(instance *cicdv1.IntegrationConfig) bool {
	secretChanged := false
	if instance.Status.Secrets == "" {
		instance.Status.Secrets = utils.RandomString(20)
		secretChanged = true
	}

	return secretChanged
}

// Set webhook-registered condition, return if it's changed or not
func (r *IntegrationConfigReconciler) setWebhookRegisteredCond(instance *cicdv1.IntegrationConfig) bool {
	webhookConditionChanged := false
	webhookRegistered := instance.Status.Conditions.GetCondition(cicdv1.IntegrationConfigConditionWebhookRegistered)
	if webhookRegistered == nil {
		webhookRegistered = &status.Condition{
			Type:   cicdv1.IntegrationConfigConditionWebhookRegistered,
			Status: corev1.ConditionFalse,
		}
	}

	// If token is empty, skip to register
	if instance.Spec.Git.Token == nil {
		webhookRegistered.Reason = cicdv1.IntegrationConfigConditionReasonNoGitToken
		webhookRegistered.Message = "Skipped to register webhook"
		webhookConditionChanged = instance.Status.Conditions.SetCondition(*webhookRegistered)
		return webhookConditionChanged
	}

	// Register only if the condition is false
	if webhookRegistered.IsFalse() {
		webhookRegistered.Status = corev1.ConditionFalse
		webhookRegistered.Reason = ""
		webhookRegistered.Message = ""

		gitCli, err := utils.GetGitCli(instance, r.Client)
		if err != nil {
			webhookRegistered.Reason = "invalidGitType"
			webhookRegistered.Message = fmt.Sprintf("git type %s is not supported", instance.Spec.Git.Type)
		} else {
			addr := instance.GetWebhookServerAddress()
			isUnique := true
			r.Log.Info("Registering webhook " + addr)
			entries, err := gitCli.ListWebhook()
			if err != nil {
				webhookRegistered.Reason = "webhookRegisterFailed"
				webhookRegistered.Message = err.Error()
			}
			for _, e := range entries {
				if addr == e.URL {
					webhookRegistered.Reason = "webhookRegisterFailed"
					webhookRegistered.Message = "same webhook has already registered"
					isUnique = false
					break
				}
			}
			if isUnique {
				if err := gitCli.RegisterWebhook(addr); err != nil {
					webhookRegistered.Reason = "webhookRegisterFailed"
					webhookRegistered.Message = err.Error()
				} else {
					webhookRegistered.Status = corev1.ConditionTrue
					webhookRegistered.Reason = ""
					webhookRegistered.Message = ""
				}
			}
		}
		webhookConditionChanged = instance.Status.Conditions.SetCondition(*webhookRegistered)
	}

	return webhookConditionChanged
}

// Set ready condition, return if it's changed or not
func (r *IntegrationConfigReconciler) setReadyCond(instance *cicdv1.IntegrationConfig) bool {
	ready := instance.Status.Conditions.GetCondition(cicdv1.IntegrationConfigConditionReady)
	if ready == nil {
		ready = &status.Condition{
			Type:   cicdv1.IntegrationConfigConditionReady,
			Status: corev1.ConditionFalse,
		}
	}

	// For now, only checked is if webhook-registered is true & secrets are set
	webhookRegistered := instance.Status.Conditions.GetCondition(cicdv1.IntegrationConfigConditionWebhookRegistered)
	if instance.Status.Secrets != "" && webhookRegistered != nil && (webhookRegistered.Status == corev1.ConditionTrue || webhookRegistered.Reason == cicdv1.IntegrationConfigConditionReasonNoGitToken) {
		ready.Status = corev1.ConditionTrue
	}
	readyConditionChanged := instance.Status.Conditions.SetCondition(*ready)

	return readyConditionChanged
}

// Create git credential secret
// The secret is parsed by tekton controller
// (ref: https://github.com/tektoncd/pipeline/blob/master/docs/auth.md#configuring-basic-auth-authentication-for-git)
func (r *IntegrationConfigReconciler) createGitSecret(instance *cicdv1.IntegrationConfig) error {
	// Get and check if values are set right
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cicdv1.GetSecretName(instance.Name),
			Namespace: instance.Namespace,
		},
	}
	if err := r.Client.Get(context.Background(), types.NamespacedName{Name: cicdv1.GetSecretName(instance.Name), Namespace: instance.Namespace}, secret); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
	}
	original := secret.DeepCopy()

	// Update content of the git secret
	needPatch, err := r.updateGitSecret(instance, secret)
	if err != nil {
		return err
	}

	// Skip create/update if everything is set right
	if !needPatch {
		return nil
	}

	if err := utils.CreateOrPatchObject(secret, original, instance, r.Client, r.Scheme); err != nil {
		return err
	}

	return nil
}

// Update content of git secret (only content. not applying to etcd)
func (r *IntegrationConfigReconciler) updateGitSecret(instance *cicdv1.IntegrationConfig, secret *corev1.Secret) (bool, error) {
	needPatch := false

	// check and set annotation
	gitHostVal, err := instance.Spec.Git.GetGitHost()
	if err != nil {
		return false, err
	}
	if secret.Annotations == nil {
		needPatch = true
		secret.Annotations = map[string]string{}
	} else if gitHostVal != secret.Annotations[gitSecretHostKey] {
		needPatch = true
	}
	secret.Annotations[gitSecretHostKey] = gitHostVal

	// check and set type
	if secret.Type != corev1.SecretTypeBasicAuth {
		needPatch = true
	}
	secret.Type = corev1.SecretTypeBasicAuth

	// check and set token
	token, err := instance.GetToken(r.Client)
	if err != nil {
		return false, err
	}
	if secret.Data == nil {
		needPatch = true
		secret.Data = map[string][]byte{}
	} else if string(secret.Data[corev1.BasicAuthUsernameKey]) != gitSecretUserName || string(secret.Data[corev1.BasicAuthPasswordKey]) != token {
		needPatch = true
	}
	secret.Data[corev1.BasicAuthUsernameKey] = []byte(gitSecretUserName)
	secret.Data[corev1.BasicAuthPasswordKey] = []byte(token)

	return needPatch, nil
}

// Create service account for pipeline run
func (r *IntegrationConfigReconciler) createServiceAccount(instance *cicdv1.IntegrationConfig) error {
	// Get and check if values are set right
	sa := &corev1.ServiceAccount{}
	if err := r.Client.Get(context.Background(), types.NamespacedName{Name: cicdv1.GetServiceAccountName(instance.Name), Namespace: instance.Namespace}, sa); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
	}
	original := sa.DeepCopy()

	sa.Name = cicdv1.GetServiceAccountName(instance.Name)
	sa.Namespace = instance.Namespace

	desiredSecrets := []corev1.LocalObjectReference{{Name: cicdv1.GetSecretName(instance.Name)}}
	desiredSecrets = append(desiredSecrets, instance.Spec.Secrets...)

	changed := false
	for _, s := range desiredSecrets {
		if s.Name == "" {
			continue
		}

		// Check if secret is set for the service account
		found := false
		for _, cur := range sa.Secrets {
			if cur.Name == s.Name {
				found = true
				break
			}
		}
		if found {
			continue
		}
		// If not set, set one
		changed = true
		sa.Secrets = append(sa.Secrets, corev1.ObjectReference{Name: s.Name})
	}

	if !changed {
		return nil
	}

	if err := controllerutil.SetControllerReference(instance, sa, r.Scheme); err != nil {
		return err
	}

	// Update if resourceVersion exists, but create if not
	if sa.ResourceVersion != "" {
		p := client.MergeFrom(original)
		if err := r.Client.Patch(context.Background(), sa, p); err != nil {
			return err
		}
	} else {
		if err := r.Client.Create(context.Background(), sa); err != nil {
			return err
		}
	}

	return nil
}
