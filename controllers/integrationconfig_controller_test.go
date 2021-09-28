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
	"testing"

	"github.com/operator-framework/operator-lib/status"
	"github.com/stretchr/testify/require"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	gitfake "github.com/tmax-cloud/cicd-operator/pkg/git/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestIntegrationConfigReconciler_Reconcile(t *testing.T) {
	s := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(s))
	utilruntime.Must(cicdv1.AddToScheme(s))

	sNoCicd := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(sNoCicd))

	sNoCore := runtime.NewScheme()
	utilruntime.Must(cicdv1.AddToScheme(sNoCore))

	tc := map[string]struct {
		ic                    *cicdv1.IntegrationConfig
		notApplied            bool
		preRegisteredWebhooks []string
		scheme                *runtime.Scheme

		errorOccurs            bool
		errorMessage           string
		expectedWebhooks       []string
		expectedFinalizers     []string
		expectedReadyStatus    corev1.ConditionStatus
		expectedReadyReason    status.ConditionReason
		expectedReadyMessage   string
		expectedWebhookStatus  corev1.ConditionStatus
		expectedWebhookReason  status.ConditionReason
		expectedWebhookMessage string
	}{
		"create": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ic",
					Namespace: "test-ns",
				},
			},
			scheme:               s,
			expectedFinalizers:   []string{finalizer},
			expectedReadyStatus:  corev1.ConditionFalse,
			expectedReadyReason:  "",
			expectedReadyMessage: "",
		},
		"hasFinalizer": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-ic",
					Namespace:  "test-ns",
					Finalizers: []string{finalizer},
				},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:       cicdv1.GitTypeFake,
						Repository: "test-repo",
						Token:      &cicdv1.GitToken{Value: "test-tkn"},
					},
				},
			},
			scheme:                 s,
			expectedWebhooks:       []string{"http://cicd-webhook.com/webhook/test-ns/test-ic"},
			expectedFinalizers:     []string{finalizer},
			expectedWebhookStatus:  corev1.ConditionTrue,
			expectedWebhookReason:  "",
			expectedWebhookMessage: "",
			expectedReadyStatus:    corev1.ConditionTrue,
			expectedReadyReason:    "",
			expectedReadyMessage:   "",
		},
		"notFound": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-ic",
					Namespace:  "test-ns",
					Finalizers: []string{finalizer},
				},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:       cicdv1.GitTypeFake,
						Repository: "test-repo",
						Token:      &cicdv1.GitToken{Value: "test-tkn"},
					},
				},
			},
			scheme:     s,
			notApplied: true,
		},
		"getError": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-ic",
					Namespace:  "test-ns",
					Finalizers: []string{finalizer},
				},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:       cicdv1.GitTypeFake,
						Repository: "test-repo",
						Token:      &cicdv1.GitToken{Value: "test-tkn"},
					},
				},
			},
			notApplied:   true,
			scheme:       sNoCicd,
			errorOccurs:  true,
			errorMessage: "no kind is registered for the type v1.IntegrationConfig in scheme \"pkg/runtime/scheme.go:101\"",
		},
		"ready": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-ic",
					Namespace:  "test-ns",
					Finalizers: []string{finalizer},
				},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:       cicdv1.GitTypeFake,
						Repository: "test-repo",
						Token:      &cicdv1.GitToken{Value: "test-tkn"},
					},
				},
				Status: cicdv1.IntegrationConfigStatus{
					Secrets: "test-secret",
					Conditions: status.Conditions{
						{Type: cicdv1.IntegrationConfigConditionReady, Status: corev1.ConditionTrue, Reason: "", Message: ""},
						{Type: cicdv1.IntegrationConfigConditionWebhookRegistered, Status: corev1.ConditionTrue, Reason: "", Message: ""},
					},
				},
			},
			scheme:                 s,
			expectedFinalizers:     []string{finalizer},
			expectedWebhookStatus:  corev1.ConditionTrue,
			expectedWebhookReason:  "",
			expectedWebhookMessage: "",
			expectedReadyStatus:    corev1.ConditionTrue,
			expectedReadyReason:    "",
			expectedReadyMessage:   "",
		},
		"createGitSecretErr": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-ic",
					Namespace:  "test-ns",
					Finalizers: []string{finalizer},
				},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:       cicdv1.GitTypeFake,
						Repository: "test-repo",
						Token: &cicdv1.GitToken{ValueFrom: &cicdv1.GitTokenFrom{
							SecretKeyRef: corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "no-exist"}, Key: "no-key"}},
						},
					},
				},
				Status: cicdv1.IntegrationConfigStatus{
					Secrets: "test-secret",
					Conditions: status.Conditions{
						{Type: cicdv1.IntegrationConfigConditionReady, Status: corev1.ConditionTrue, Reason: "", Message: ""},
						{Type: cicdv1.IntegrationConfigConditionWebhookRegistered, Status: corev1.ConditionTrue, Reason: "", Message: ""},
					},
				},
			},
			scheme:                 s,
			expectedFinalizers:     []string{finalizer},
			expectedWebhookStatus:  corev1.ConditionTrue,
			expectedWebhookReason:  "",
			expectedWebhookMessage: "",
			expectedReadyStatus:    corev1.ConditionFalse,
			expectedReadyReason:    "CannotCreateSecret",
			expectedReadyMessage:   "secrets \"no-exist\" not found",
		},
		"createServiceAccountErr": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-ic",
					Namespace:  "test-ns",
					Finalizers: []string{finalizer},
				},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:       cicdv1.GitTypeFake,
						Repository: "test-repo",
						Token:      &cicdv1.GitToken{Value: "test-tkn"},
					},
				},
			},
			scheme:                 sNoCore,
			expectedWebhooks:       []string{"http://cicd-webhook.com/webhook/test-ns/test-ic"},
			expectedFinalizers:     []string{finalizer},
			expectedWebhookStatus:  corev1.ConditionTrue,
			expectedWebhookReason:  "",
			expectedWebhookMessage: "",
			expectedReadyStatus:    corev1.ConditionFalse,
			expectedReadyReason:    "CannotCreateAccount",
			expectedReadyMessage:   "no kind is registered for the type v1.ServiceAccount in scheme \"pkg/runtime/scheme.go:101\"",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			configs.CurrentExternalHostName = "cicd-webhook.com"
			fakeCli := fake.NewFakeClientWithScheme(c.scheme)
			if !c.notApplied {
				require.NoError(t, fakeCli.Create(context.Background(), c.ic))
			}

			gitfake.Repos = map[string]*gitfake.Repo{
				"test-repo": {
					Webhooks: map[int]*git.WebhookEntry{},
				},
			}
			for i, w := range c.preRegisteredWebhooks {
				gitfake.Repos["test-repo"].Webhooks[i] = &git.WebhookEntry{ID: i, URL: w}
			}

			reconciler := &IntegrationConfigReconciler{Log: &fakeLogger{}, Scheme: c.scheme, Client: fakeCli}

			_, err := reconciler.Reconcile(ctrl.Request{NamespacedName: types.NamespacedName{Name: c.ic.Name, Namespace: c.ic.Namespace}})
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
				if c.notApplied {
					return
				}

				result := &cicdv1.IntegrationConfig{}
				require.NoError(t, fakeCli.Get(context.Background(), types.NamespacedName{Name: c.ic.Name, Namespace: c.ic.Namespace}, result))

				require.Len(t, gitfake.Repos["test-repo"].Webhooks, len(c.expectedWebhooks))
				for _, w := range c.expectedWebhooks {
					found := false
					for _, ww := range gitfake.Repos["test-repo"].Webhooks {
						if w == ww.URL {
							found = true
							break
						}
					}
					require.True(t, found)
				}

				require.Equal(t, c.expectedFinalizers, result.Finalizers)

				webhookCond := result.Status.Conditions.GetCondition(cicdv1.IntegrationConfigConditionWebhookRegistered)
				if c.expectedWebhookStatus == "" {
					require.Nil(t, webhookCond)
				} else {
					require.NotNil(t, webhookCond)
					require.Equal(t, c.expectedWebhookStatus, webhookCond.Status)
					require.Equal(t, c.expectedWebhookReason, webhookCond.Reason)
					require.Equal(t, c.expectedWebhookMessage, webhookCond.Message)
				}

				readyCond := result.Status.Conditions.GetCondition(cicdv1.IntegrationConfigConditionReady)
				if c.expectedReadyStatus == "" {
					require.Nil(t, readyCond)
				} else {
					require.NotNil(t, readyCond)
					require.Equal(t, c.expectedReadyStatus, readyCond.Status)
					require.Equal(t, c.expectedReadyReason, readyCond.Reason)
					require.Equal(t, c.expectedReadyMessage, readyCond.Message)
				}
			}
		})
	}
}

func TestIntegrationConfigReconciler_SetupWithManager(t *testing.T) {
	s := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(s))
	utilruntime.Must(cicdv1.AddToScheme(s))

	mgr := &fakeManager{scheme: s}

	reconciler := &IntegrationConfigReconciler{}
	require.NoError(t, reconciler.SetupWithManager(mgr))
}

func TestIntegrationConfigReconciler_handleFinalizer(t *testing.T) {
	s := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(s))
	utilruntime.Must(cicdv1.AddToScheme(s))

	nowTime := metav1.Now()

	tc := map[string]struct {
		ic                    *cicdv1.IntegrationConfig
		notApplied            bool
		preRegisteredWebhooks []string

		doExit             bool
		errorOccurs        bool
		errorMessage       string
		expectedWebhooks   []string
		expectedFinalizers []string
	}{
		"finalizerNotFound": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ic",
					Namespace: "test-ns",
				},
			},
			doExit:             true,
			expectedFinalizers: []string{finalizer},
		},
		"noDelete": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-ic",
					Namespace:  "test-ns",
					Finalizers: []string{finalizer},
				},
			},
			preRegisteredWebhooks: []string{"http://cicd-webhook.com/webhook/test-ns/test-ic"},
			doExit:                false,
			expectedFinalizers:    []string{finalizer},
			expectedWebhooks:      []string{"http://cicd-webhook.com/webhook/test-ns/test-ic"},
		},
		"delete": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test-ic",
					Namespace:         "test-ns",
					Finalizers:        []string{finalizer},
					DeletionTimestamp: &nowTime,
				},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:       cicdv1.GitTypeFake,
						Repository: "test-repo",
						Token:      &cicdv1.GitToken{Value: "test-tkn"},
					},
				},
			},
			preRegisteredWebhooks: []string{"http://cicd-webhook.com/webhook/test-ns/test-ic"},
			doExit:                true,
		},
		"deleteGitCliErr": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test-ic",
					Namespace:         "test-ns",
					Finalizers:        []string{finalizer},
					DeletionTimestamp: &nowTime,
				},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:       cicdv1.GitTypeFake,
						Repository: "test-repo2",
						Token:      &cicdv1.GitToken{Value: "test-tkn"},
					},
				},
			},
			doExit: true,
		},
		"deleteGitCliUnknown": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test-ic",
					Namespace:         "test-ns",
					Finalizers:        []string{finalizer},
					DeletionTimestamp: &nowTime,
				},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:       "unknown",
						Repository: "test-repo2",
						Token:      &cicdv1.GitToken{Value: "test-tkn"},
					},
				},
			},
			doExit: true,
		},
		"deleteMultiFin": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test-ic",
					Namespace:         "test-ns",
					Finalizers:        []string{finalizer, "another-one"},
					DeletionTimestamp: &nowTime,
				},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:       cicdv1.GitTypeFake,
						Repository: "test-repo",
						Token:      &cicdv1.GitToken{Value: "test-tkn"},
					},
				},
			},
			preRegisteredWebhooks: []string{"http://cicd-webhook.com/webhook/test-ns/test-ic"},
			doExit:                true,
			expectedFinalizers:    []string{"another-one"},
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			configs.CurrentExternalHostName = "cicd-webhook.com"
			fakeCli := fake.NewFakeClientWithScheme(s)
			if !c.notApplied {
				require.NoError(t, fakeCli.Create(context.Background(), c.ic))
			}
			reconciler := &IntegrationConfigReconciler{Log: &fakeLogger{}, Scheme: s, Client: fakeCli}

			gitfake.Repos = map[string]*gitfake.Repo{
				"test-repo": {
					Webhooks: map[int]*git.WebhookEntry{},
				},
			}
			for i, w := range c.preRegisteredWebhooks {
				gitfake.Repos["test-repo"].Webhooks[i] = &git.WebhookEntry{ID: i, URL: w}
			}

			exit, err := reconciler.handleFinalizer(c.ic)
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, c.doExit, exit)

				// Check Finalizer
				require.Equal(t, c.expectedFinalizers, c.ic.Finalizers)

				// Check webhooks
				require.Len(t, gitfake.Repos["test-repo"].Webhooks, len(c.expectedWebhooks))
				for _, w := range c.expectedWebhooks {
					found := false
					for _, ww := range gitfake.Repos["test-repo"].Webhooks {
						if w == ww.URL {
							found = true
							break
						}
					}
					require.True(t, found)
				}
			}
		})
	}
}

func TestIntegrationConfigReconciler_setSecretString(t *testing.T) {
	tc := map[string]struct {
		ic *cicdv1.IntegrationConfig
	}{
		"notSet": {
			ic: &cicdv1.IntegrationConfig{},
		},
		"alreadySet": {
			ic: &cicdv1.IntegrationConfig{
				Status: cicdv1.IntegrationConfigStatus{Secrets: "secret-test"},
			},
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			reconciler := &IntegrationConfigReconciler{}
			reconciler.setSecretString(c.ic)
			require.NotEmpty(t, c.ic.Status.Secrets)
		})
	}
}

func TestIntegrationConfigReconciler_setWebhookRegisteredCond(t *testing.T) {
	tc := map[string]struct {
		ic                      *cicdv1.IntegrationConfig
		preRegisteredWebhookURL string

		expectedWebhookURL string
		expectedStatus     corev1.ConditionStatus
		expectedReason     status.ConditionReason
		expectedMessage    string
	}{
		"create": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ic",
					Namespace: "test-ns",
				},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:       cicdv1.GitTypeFake,
						Repository: "test-repo",
						Token:      &cicdv1.GitToken{Value: "test-tkn"},
					},
				},
			},
			expectedWebhookURL: "http://cicd-webhook.com/webhook/test-ns/test-ic",
			expectedStatus:     corev1.ConditionTrue,
			expectedReason:     "",
			expectedMessage:    "",
		},
		"noToken": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ic",
					Namespace: "test-ns",
				},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:       cicdv1.GitTypeFake,
						Repository: "test-repo",
					},
				},
			},
			expectedWebhookURL: "",
			expectedStatus:     corev1.ConditionFalse,
			expectedReason:     "noGitToken",
			expectedMessage:    "Skipped to register webhook",
		},
		"getGitCliErr": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ic",
					Namespace: "test-ns",
				},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:       "dummy",
						Repository: "test-repo",
						Token:      &cicdv1.GitToken{Value: "test-tkn"},
					},
				},
			},
			expectedWebhookURL: "",
			expectedStatus:     corev1.ConditionFalse,
			expectedReason:     "gitCliErr",
			expectedMessage:    "git type dummy is not supported",
		},
		"listWebhookErr": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ic",
					Namespace: "test-ns",
				},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:       cicdv1.GitTypeFake,
						Repository: "test-repo2",
						Token:      &cicdv1.GitToken{Value: "test-tkn"},
					},
				},
			},
			expectedWebhookURL: "",
			expectedStatus:     corev1.ConditionFalse,
			expectedReason:     "webhookRegisterFailed",
			expectedMessage:    "404 no such repository",
		},
		"webhookAlreadyRegistered": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ic",
					Namespace: "test-ns",
				},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:       cicdv1.GitTypeFake,
						Repository: "test-repo",
						Token:      &cicdv1.GitToken{Value: "test-tkn"},
					},
				},
			},
			preRegisteredWebhookURL: "http://cicd-webhook.com/webhook/test-ns/test-ic",
			expectedWebhookURL:      "",
			expectedStatus:          corev1.ConditionFalse,
			expectedReason:          "webhookRegisterFailed",
			expectedMessage:         "same webhook has already registered",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			configs.CurrentExternalHostName = "cicd-webhook.com"
			gitfake.Repos = map[string]*gitfake.Repo{
				"test-repo": {
					Webhooks: map[int]*git.WebhookEntry{},
				},
			}
			if c.preRegisteredWebhookURL != "" {
				gitfake.Repos["test-repo"].Webhooks[32] = &git.WebhookEntry{ID: 32, URL: c.preRegisteredWebhookURL}
			}

			reconciler := &IntegrationConfigReconciler{Log: &fakeLogger{}}
			reconciler.setWebhookRegisteredCond(c.ic)

			if c.expectedWebhookURL != "" {
				found := false
				for _, w := range gitfake.Repos["test-repo"].Webhooks {
					if w.URL == c.expectedWebhookURL {
						found = true
						break
					}
				}
				require.True(t, found)
			}

			cond := c.ic.Status.Conditions.GetCondition(cicdv1.IntegrationConfigConditionWebhookRegistered)
			require.NotNil(t, cond)
			require.Equal(t, c.expectedStatus, cond.Status)
			require.Equal(t, c.expectedReason, cond.Reason)
			require.Equal(t, c.expectedMessage, cond.Message)
		})
	}
}

func TestIntegrationConfigReconciler_setReadyCond(t *testing.T) {
	tc := map[string]struct {
		ic *cicdv1.IntegrationConfig

		expectedReadyCondStatus corev1.ConditionStatus
	}{
		"create": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ic",
					Namespace: "test-ns",
				},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:  cicdv1.GitTypeGitHub,
						Token: &cicdv1.GitToken{Value: "test-tkn"},
					},
				},
			},
			expectedReadyCondStatus: corev1.ConditionFalse,
		},
		"noop": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ic",
					Namespace: "test-ns",
				},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:  cicdv1.GitTypeGitHub,
						Token: &cicdv1.GitToken{Value: "test-tkn"},
					},
				},
				Status: cicdv1.IntegrationConfigStatus{
					Conditions: []status.Condition{
						{Type: "webhook-registered", Status: corev1.ConditionTrue},
						{Type: "ready", Status: corev1.ConditionTrue},
					},
				},
			},
			expectedReadyCondStatus: corev1.ConditionTrue,
		},
		"webhookRegistered": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ic",
					Namespace: "test-ns",
				},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:  cicdv1.GitTypeGitHub,
						Token: &cicdv1.GitToken{Value: "test-tkn"},
					},
				},
				Status: cicdv1.IntegrationConfigStatus{
					Conditions: []status.Condition{
						{Type: "webhook-registered", Status: corev1.ConditionTrue},
					},
					Secrets: "test-secret",
				},
			},
			expectedReadyCondStatus: corev1.ConditionTrue,
		},
		"noRegisterNeeded": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ic",
					Namespace: "test-ns",
				},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:  cicdv1.GitTypeGitHub,
						Token: &cicdv1.GitToken{Value: "test-tkn"},
					},
				},
				Status: cicdv1.IntegrationConfigStatus{
					Conditions: []status.Condition{
						{Type: "webhook-registered", Status: corev1.ConditionFalse, Reason: "noGitToken"},
					},
					Secrets: "test-secret",
				},
			},
			expectedReadyCondStatus: corev1.ConditionTrue,
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			reconciler := &IntegrationConfigReconciler{}
			cond := c.ic.Status.Conditions.GetCondition("ready")
			if cond == nil {
				cond = &status.Condition{
					Type:   cicdv1.IntegrationConfigConditionReady,
					Status: corev1.ConditionFalse,
				}
			}
			reconciler.setReadyCond(c.ic, cond)
			require.Equal(t, c.expectedReadyCondStatus, cond.Status)
		})
	}
}

func TestIntegrationConfigReconciler_createGitSecret(t *testing.T) {
	s := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(s))
	utilruntime.Must(cicdv1.AddToScheme(s))

	sNoCore := runtime.NewScheme()
	utilruntime.Must(cicdv1.AddToScheme(sNoCore))

	tc := map[string]struct {
		ic     *cicdv1.IntegrationConfig
		scheme *runtime.Scheme
		secret *corev1.Secret

		errorOccurs   bool
		errorMessage  string
		expectedToken string
	}{
		"create": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ic",
					Namespace: "test-ns",
				},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:  cicdv1.GitTypeGitHub,
						Token: &cicdv1.GitToken{Value: "test-tkn"},
					},
				},
			},
			scheme:        s,
			expectedToken: "test-tkn",
		},
		"secretGetErr": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ic",
					Namespace: "test-ns",
				},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:  cicdv1.GitTypeGitHub,
						Token: &cicdv1.GitToken{Value: "test-tkn"},
					},
				},
			},
			scheme:       sNoCore,
			errorOccurs:  true,
			errorMessage: "no kind is registered for the type v1.Secret in scheme \"pkg/runtime/scheme.go:101\"",
		},
		"updateSecretErr": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ic",
					Namespace: "test-ns",
				},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type: cicdv1.GitTypeGitHub,
						Token: &cicdv1.GitToken{ValueFrom: &cicdv1.GitTokenFrom{SecretKeyRef: corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "token-secret"}, Key: "asd"},
						}},
					},
				},
			},
			scheme:       s,
			errorOccurs:  true,
			errorMessage: "secrets \"token-secret\" not found",
		},
		"noPatchNeeded": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ic",
					Namespace: "test-ns",
				},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:  cicdv1.GitTypeGitHub,
						Token: &cicdv1.GitToken{Value: "test-tkn"},
					},
				},
			},
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cicdv1.GetSecretName("test-ic"),
					Namespace: "test-ns",
					Annotations: map[string]string{
						"tekton.dev/git-0": "https://github.com",
					},
				},
				Type: corev1.SecretTypeBasicAuth,
				Data: map[string][]byte{
					"username": []byte("tmax-cicd-bot"),
					"password": []byte("test-tkn"),
				},
			},
			scheme:        s,
			expectedToken: "test-tkn",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			fakeCli := fake.NewFakeClientWithScheme(c.scheme)
			if c.secret != nil {
				require.NoError(t, fakeCli.Create(context.Background(), c.secret))
			}

			reconciler := &IntegrationConfigReconciler{Scheme: c.scheme, Client: fakeCli}
			err := reconciler.createGitSecret(c.ic)
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)

				secret := &corev1.Secret{}
				require.NoError(t, fakeCli.Get(context.Background(), types.NamespacedName{Name: "test-ic", Namespace: "test-ns"}, secret))

				require.Equal(t, map[string]string{"tekton.dev/git-0": "https://github.com"}, secret.Annotations)
				require.Equal(t, map[string][]byte{"username": []byte("tmax-cicd-bot"), "password": []byte(c.expectedToken)}, secret.Data)
			}
		})
	}
}

func TestIntegrationConfigReconciler_updateGitSecret(t *testing.T) {
	s := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(s))
	utilruntime.Must(cicdv1.AddToScheme(s))

	fakeCli := fake.NewFakeClientWithScheme(s)

	tc := map[string]struct {
		ic     *cicdv1.IntegrationConfig
		secret *corev1.Secret

		errorOccurs  bool
		errorMessage string

		doPatch       bool
		expectedToken string
	}{
		"create": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ic",
					Namespace: "test-ns",
				},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:  cicdv1.GitTypeGitHub,
						Token: &cicdv1.GitToken{Value: "test-tkn"},
					},
				},
			},
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cicdv1.GetSecretName("test-ic"),
					Namespace: "test-ns",
				},
			},
			doPatch:       true,
			expectedToken: "test-tkn",
		},
		"gitHostErr": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ic",
					Namespace: "test-ns",
				},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:   cicdv1.GitTypeGitHub,
						APIUrl: "ht~~~p://~~**.",
						Token:  &cicdv1.GitToken{Value: "test-tkn"},
					},
				},
			},
			errorOccurs:  true,
			errorMessage: "parse \"ht~~~p://~~**.\": first path segment in URL cannot contain colon",
		},
		"wrongAnnotation": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ic",
					Namespace: "test-ns",
				},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:  cicdv1.GitTypeGitHub,
						Token: &cicdv1.GitToken{Value: "test-tkn"},
					},
				},
			},
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cicdv1.GetSecretName("test-ic"),
					Namespace: "test-ns",
					Annotations: map[string]string{
						"tekton.dev/git-0": "https://github.com/////",
					},
				},
			},
			doPatch:       true,
			expectedToken: "test-tkn",
		},
		"getTokenErr": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ic",
					Namespace: "test-ns",
				},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type: cicdv1.GitTypeGitHub,
						Token: &cicdv1.GitToken{ValueFrom: &cicdv1.GitTokenFrom{SecretKeyRef: corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "token-secret"}, Key: "asd"},
						}},
					},
				},
			},
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cicdv1.GetSecretName("test-ic"),
					Namespace: "test-ns",
				},
			},
			errorOccurs:  true,
			errorMessage: "secrets \"token-secret\" not found",
		},
		"wrongToken": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ic",
					Namespace: "test-ns",
				},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:  cicdv1.GitTypeGitHub,
						Token: &cicdv1.GitToken{Value: "test-tkn"},
					},
				},
			},
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cicdv1.GetSecretName("test-ic"),
					Namespace: "test-ns",
					Annotations: map[string]string{
						"tekton.dev/git-0": "https://github.com",
					},
				},
				Data: map[string][]byte{
					"username": []byte("tmax-cicd-bot"),
					"password": []byte("test-tkkkkkn"),
				},
			},
			doPatch:       true,
			expectedToken: "test-tkn",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			reconciler := &IntegrationConfigReconciler{Scheme: s, Client: fakeCli}
			doPatch, err := reconciler.updateGitSecret(c.ic, c.secret)
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, c.doPatch, doPatch)

				require.Equal(t, map[string]string{"tekton.dev/git-0": "https://github.com"}, c.secret.Annotations)
				require.Equal(t, map[string][]byte{"username": []byte("tmax-cicd-bot"), "password": []byte(c.expectedToken)}, c.secret.Data)
			}
		})
	}
}

func TestIntegrationConfigReconciler_createServiceAccount(t *testing.T) {
	s := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(s))
	utilruntime.Must(cicdv1.AddToScheme(s))

	sNoCicd := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(sNoCicd))

	sNoCore := runtime.NewScheme()
	utilruntime.Must(cicdv1.AddToScheme(sNoCore))

	tc := map[string]struct {
		ic           *cicdv1.IntegrationConfig
		dontCreateIC bool
		sa           *corev1.ServiceAccount
		scheme       *runtime.Scheme

		errorOccurs  bool
		errorMessage string
		verifyFunc   func(t *testing.T, reconciler *IntegrationConfigReconciler)
	}{
		"saGetError": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ic", Namespace: "test-ns"},
			},
			scheme:       sNoCore,
			errorOccurs:  true,
			errorMessage: "no kind is registered for the type v1.ServiceAccount in scheme \"pkg/runtime/scheme.go:101\"",
		},
		"secretBlankName": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ic", Namespace: "test-ns"},
				Spec: cicdv1.IntegrationConfigSpec{
					Secrets: []corev1.LocalObjectReference{
						{Name: ""},
						{Name: "test-secret"},
					},
				},
			},
			scheme: s,
			verifyFunc: func(t *testing.T, reconciler *IntegrationConfigReconciler) {
				saResult := &corev1.ServiceAccount{}
				require.NoError(t, reconciler.Client.Get(context.Background(), types.NamespacedName{Name: cicdv1.GetServiceAccountName("test-ic"), Namespace: "test-ns"}, saResult))
				require.Equal(t, []corev1.ObjectReference{
					{Name: "test-ic"},
					{Name: "test-secret"},
				}, saResult.Secrets)
			},
		},
		"secretAlreadySet": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ic", Namespace: "test-ns"},
				Spec: cicdv1.IntegrationConfigSpec{
					Secrets: []corev1.LocalObjectReference{
						{Name: "test-secret"},
						{Name: "test-secret2"},
					},
				},
			},
			sa: &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{Name: cicdv1.GetServiceAccountName("test-ic"), Namespace: "test-ns"},
				Secrets: []corev1.ObjectReference{
					{Name: "test-ic"},
					{Name: "test-secret"},
					{Name: "test-secret2"},
					{Name: "another-one"},
				},
			},
			scheme: s,
			verifyFunc: func(t *testing.T, reconciler *IntegrationConfigReconciler) {
				saResult := &corev1.ServiceAccount{}
				require.NoError(t, reconciler.Client.Get(context.Background(), types.NamespacedName{Name: cicdv1.GetServiceAccountName("test-ic"), Namespace: "test-ns"}, saResult))
				require.Equal(t, []corev1.ObjectReference{
					{Name: "test-ic"},
					{Name: "test-secret"},
					{Name: "test-secret2"},
					{Name: "another-one"},
				}, saResult.Secrets)
			},
		},
		"secretSetNow": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ic", Namespace: "test-ns"},
				Spec: cicdv1.IntegrationConfigSpec{
					Secrets: []corev1.LocalObjectReference{
						{Name: "test-secret"},
						{Name: "test-secret2"},
					},
				},
			},
			sa: &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{Name: cicdv1.GetServiceAccountName("test-ic"), Namespace: "test-ns"},
				Secrets: []corev1.ObjectReference{
					{Name: "test-ic"},
					{Name: "another-one"},
				},
			},
			scheme: s,
			verifyFunc: func(t *testing.T, reconciler *IntegrationConfigReconciler) {
				saResult := &corev1.ServiceAccount{}
				require.NoError(t, reconciler.Client.Get(context.Background(), types.NamespacedName{Name: cicdv1.GetServiceAccountName("test-ic"), Namespace: "test-ns"}, saResult))
				require.Equal(t, []corev1.ObjectReference{
					{Name: "test-ic"},
					{Name: "another-one"},
					{Name: "test-secret"},
					{Name: "test-secret2"},
				}, saResult.Secrets)
			},
		},
		"ownerRefError": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ic", Namespace: "test-ns"},
				Spec: cicdv1.IntegrationConfigSpec{
					Secrets: []corev1.LocalObjectReference{
						{Name: "test-secret"},
						{Name: "test-secret2"},
					},
				},
			},
			dontCreateIC: true,
			scheme:       sNoCicd,
			errorOccurs:  true,
			errorMessage: "no kind is registered for the type v1.IntegrationConfig in scheme \"pkg/runtime/scheme.go:101\"",
		},
		"createSaError": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ic", Namespace: "test-ns"},
				Spec: cicdv1.IntegrationConfigSpec{
					Secrets: []corev1.LocalObjectReference{
						{Name: "test-secret"},
					},
				},
			},
			scheme:       sNoCore,
			errorOccurs:  true,
			errorMessage: "no kind is registered for the type v1.ServiceAccount in scheme \"pkg/runtime/scheme.go:101\"",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			reconciler := &IntegrationConfigReconciler{Scheme: c.scheme}
			if c.dontCreateIC {
				reconciler.Client = fake.NewFakeClientWithScheme(c.scheme)
			} else {
				reconciler.Client = fake.NewFakeClientWithScheme(c.scheme, c.ic)
			}
			if c.sa != nil {
				_ = reconciler.Client.Create(context.Background(), c.sa)
			}

			err := reconciler.createServiceAccount(c.ic)
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
				c.verifyFunc(t, reconciler)
			}
		})
	}
}
