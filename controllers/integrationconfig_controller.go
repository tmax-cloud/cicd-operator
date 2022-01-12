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
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/go-logr/logr"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/periodictrigger"
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

var periodicTriggers = map[string]*periodictrigger.PeriodicTrigger{}

// +kubebuilder:rbac:groups=cicd.tmax.io,resources=integrationconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cicd.tmax.io,resources=integrationconfigs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=secrets;serviceaccounts,verbs=get;list;watch;create;update;patch;delete

// Reconcile reconciles IntegrationConfig
func (r *IntegrationConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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
	if cond := meta.FindStatusCondition(instance.Status.Conditions, cicdv1.IntegrationConfigConditionReady); cond == nil {
		meta.SetStatusCondition(&instance.Status.Conditions, metav1.Condition{
			Type:    cicdv1.IntegrationConfigConditionReady,
			Status:  metav1.ConditionFalse,
			Reason:  "NotReady",
			Message: "Not ready",
		})
	}

	r.bumpV050(instance)

	specChanged := false
	defer func(specChanged *bool) {
		p := client.MergeFrom(original)

		if *specChanged {
			if err := r.Client.Patch(ctx, instance, p); err != nil {
				log.Error(err, "")
			}
		} else {
			if err := r.Client.Status().Patch(ctx, instance, p); err != nil {
				log.Error(err, "")
			}
		}
	}(&specChanged)

	if specChanged = r.handleFinalizer(instance); specChanged {
		return ctrl.Result{}, nil
	}

	// Set secret
	r.setSecretString(instance)

	// Set webhook registered
	var re reconcile.Result
	if resetTime := r.setWebhookRegisteredCond(instance); resetTime > 0 {
		// Get time remaining from reset time and set to run reconcile at that time.
		re = ctrl.Result{RequeueAfter: time.Duration(git.GetGapTime(resetTime)) * time.Second, Requeue: true}
	} else {
		re = ctrl.Result{}
	}

	// Set ready
	r.setReadyCond(instance)

	if instance.Spec.Jobs.Periodic != nil {
		r.setPeriodicTrigger(instance)
	}

	// Service account for running PipelineRuns
	if err := r.createServiceAccount(instance); err != nil {
		log.Error(err, "")
		cond := meta.FindStatusCondition(instance.Status.Conditions, cicdv1.IntegrationConfigConditionReady)
		cond.Status = metav1.ConditionFalse
		cond.Reason = "CannotCreateAccount"
		cond.Message = err.Error()
		return ctrl.Result{}, nil
	}

	// Git credential secret - referred by tekton
	if err := r.createGitSecret(instance); err != nil {
		log.Error(err, "")
		cond := meta.FindStatusCondition(instance.Status.Conditions, cicdv1.IntegrationConfigConditionReady)
		cond.Status = metav1.ConditionFalse
		cond.Reason = "CannotCreateSecret"
		cond.Message = err.Error()
		return ctrl.Result{}, nil
	}

	return re, nil
}

// SetupWithManager sets IntegrationConfigReconciler to the manager
func (r *IntegrationConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cicdv1.IntegrationConfig{}).
		Complete(r)
}

// Update to v0.5.0 - reason, message became required
func (r *IntegrationConfigReconciler) bumpV050(instance *cicdv1.IntegrationConfig) {
	// Bump ready cond
	if condReady := meta.FindStatusCondition(instance.Status.Conditions, cicdv1.IntegrationConfigConditionReady); condReady != nil {
		upgradeV050Condition(condReady, "Ready", "NotReady")
	}

	// Bump webhook-registered cond
	if condReg := meta.FindStatusCondition(instance.Status.Conditions, cicdv1.IntegrationConfigConditionWebhookRegistered); condReg != nil {
		upgradeV050Condition(condReg, "Registered", "NotRegistered")
	}
}

// handleFinalizer handles finalizer (add or remove) and returns whether to exit or not (for spec update)
func (r *IntegrationConfigReconciler) handleFinalizer(instance *cicdv1.IntegrationConfig) bool {
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
		return true
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

		// Stop periodic trigger and reomve the periodic trigger.
		// Check if periodicTrigger exists
		nameAndNamespace := instance.Name + instance.Namespace
		_, exists := periodicTriggers[nameAndNamespace]
		if exists {
			periodicTriggers[nameAndNamespace].Stop()
			delete(periodicTriggers, nameAndNamespace)
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
		return true
	}

	return false
}

// Set status.secrets, return if it's changed or not
func (r *IntegrationConfigReconciler) setSecretString(instance *cicdv1.IntegrationConfig) {
	if instance.Status.Secrets == "" {
		instance.Status.Secrets = utils.RandomString(20)
	}
}

// Set webhook-registered condition, return if it's changed or not
func (r *IntegrationConfigReconciler) setWebhookRegisteredCond(instance *cicdv1.IntegrationConfig) int {
	webhookRegistered := meta.FindStatusCondition(instance.Status.Conditions, cicdv1.IntegrationConfigConditionWebhookRegistered)
	if webhookRegistered == nil {
		meta.SetStatusCondition(&instance.Status.Conditions, metav1.Condition{
			Type:   cicdv1.IntegrationConfigConditionWebhookRegistered,
			Status: metav1.ConditionFalse,
			Reason: "NotRegistered",
		})
		webhookRegistered = meta.FindStatusCondition(instance.Status.Conditions, cicdv1.IntegrationConfigConditionWebhookRegistered)
	}

	// If token is empty, skip to register
	if instance.Spec.Git.Token == nil {
		webhookRegistered.Reason = cicdv1.IntegrationConfigConditionReasonNoGitToken
		webhookRegistered.Message = "Skipped to register webhook"
		return 0
	}

	// Register only if the condition is false
	if webhookRegistered.Status == metav1.ConditionFalse {
		webhookRegistered.Status = metav1.ConditionFalse
		webhookRegistered.Reason = "NotRegistered"
		webhookRegistered.Message = "Webhook is not registered"

		gitCli, err := utils.GetGitCli(instance, r.Client)
		if err != nil {
			webhookRegistered.Reason = "gitCliErr"
			webhookRegistered.Message = err.Error()
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
				if err = gitCli.RegisterWebhook(addr); err != nil {
					webhookRegistered.Reason = "webhookRegisterFailed"
					webhookRegistered.Message = err.Error()
				} else {
					webhookRegistered.Status = metav1.ConditionTrue
					webhookRegistered.Reason = "Registered"
					webhookRegistered.Message = "Webhook is registered"
				}
			}
			if err != nil {
				return git.CheckRateLimitGetResetTime(err)
			}
		}
	}
	return 0
}

// Set ready condition, return if it's changed or not
func (r *IntegrationConfigReconciler) setReadyCond(instance *cicdv1.IntegrationConfig) {
	cond := meta.FindStatusCondition(instance.Status.Conditions, cicdv1.IntegrationConfigConditionReady)
	// For now, only checked is if webhook-registered is true & secrets are set
	webhookRegistered := meta.FindStatusCondition(instance.Status.Conditions, cicdv1.IntegrationConfigConditionWebhookRegistered)
	if instance.Status.Secrets != "" && webhookRegistered != nil && (webhookRegistered.Status == metav1.ConditionTrue || webhookRegistered.Reason == cicdv1.IntegrationConfigConditionReasonNoGitToken) {
		cond.Status = metav1.ConditionTrue
		cond.Reason = "Ready"
		cond.Message = "Ready"
	}
}

func (r *IntegrationConfigReconciler) setPeriodicTrigger(instance *cicdv1.IntegrationConfig) {
	// Check if periodicTrigger exists
	nameAndNamespace := instance.Name + instance.Namespace
	_, exists := periodicTriggers[nameAndNamespace]
	if !exists {
		// create periodic trigger
		periodicTriggers[nameAndNamespace] = periodictrigger.New(r.Client, instance, context.Background())
		err := periodicTriggers[nameAndNamespace].Start()

		if err != nil {
			r.Log.Error(err, "Starting periodicTrigger failed")
		}
	}
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

	return utils.CreateOrPatchObject(secret, original, instance, r.Client, r.Scheme)
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
		return r.Client.Patch(context.Background(), sa, p)
	}

	return r.Client.Create(context.Background(), sa)
}

func upgradeV050Condition(cond *metav1.Condition, trueMsg, falseMsg string) {
	var msg string
	switch cond.Status {
	case metav1.ConditionTrue:
		msg = trueMsg
	case metav1.ConditionFalse:
		msg = falseMsg
	default:
		msg = "Unknown"
	}

	if cond.Reason == "" {
		cond.Reason = msg
	}
	if cond.Message == "" {
		cond.Message = msg
	}
}
