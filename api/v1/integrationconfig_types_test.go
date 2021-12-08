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

package v1

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestIntegrationConfig_GetToken(t *testing.T) {
	s := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(s))
	utilruntime.Must(AddToScheme(s))

	secret1 := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret1",
			Namespace: "test-ns",
		},
		Data: map[string][]byte{
			"token": []byte("test-tkn"),
		},
	}

	cli := fake.NewClientBuilder().WithScheme(s).WithObjects(secret1).Build()

	tc := map[string]struct {
		gitToken *GitToken

		errorOccurs   bool
		errorMessage  string
		expectedToken string
	}{
		"noToken": {},
		"value": {
			gitToken:      &GitToken{Value: "test-tkn"},
			expectedToken: "test-tkn",
		},
		"valueNoValue": {
			gitToken:     &GitToken{},
			errorOccurs:  true,
			errorMessage: "token is empty",
		},
		"valueFrom": {
			gitToken: &GitToken{
				ValueFrom: &GitTokenFrom{
					SecretKeyRef: corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "secret1",
						},
						Key: "token",
					},
				},
			},
			expectedToken: "test-tkn",
		},
		"noSecret": {
			gitToken: &GitToken{
				ValueFrom: &GitTokenFrom{
					SecretKeyRef: corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "secret3",
						},
						Key: "token",
					},
				},
			},
			errorOccurs:  true,
			errorMessage: "secrets \"secret3\" not found",
		},
		"noSecretKey": {
			gitToken: &GitToken{
				ValueFrom: &GitTokenFrom{
					SecretKeyRef: corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "secret1",
						},
						Key: "token1",
					},
				},
			},
			errorOccurs:  true,
			errorMessage: "token secret/key secret1/token1 not valid",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			ic := &IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ic",
					Namespace: "test-ns",
				},
				Spec: IntegrationConfigSpec{
					Git: GitConfig{Token: c.gitToken},
				},
			}
			tok, err := ic.GetToken(cli)
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, c.expectedToken, tok)
			}
		})
	}
}

func TestIntegrationConfig_GetWebhookServerAddress(t *testing.T) {
	configs.CurrentExternalHostName = "test.host.com"
	ic := &IntegrationConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ic",
			Namespace: "test-ns",
		},
	}
	require.Equal(t, "http://test.host.com/webhook/test-ns/test-ic", ic.GetWebhookServerAddress())
}

func TestGetServiceAccountName(t *testing.T) {
	require.Equal(t, "test-cfg-sa", GetServiceAccountName("test-cfg"))
}

func TestGetSecretName(t *testing.T) {
	require.Equal(t, "test-cfg", GetSecretName("test-cfg"))
}

func TestGetDuration(t *testing.T) {
	configs.IntegrationJobTTL = 120
	tc := map[string]struct {
		timeout string

		expectedErrOccur bool
		expectedDuration time.Duration
	}{
		"validTimeout": {
			timeout: "1h",

			expectedErrOccur: false,
			expectedDuration: 1 * time.Hour,
		},
		"invalidTimeout": {
			timeout: "1",

			expectedErrOccur: true,
			expectedDuration: 1 * time.Hour,
		},
		"default": {
			expectedDuration: 120 * time.Hour,
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			var ic *IntegrationConfig = &IntegrationConfig{
				Spec: IntegrationConfigSpec{},
			}
			if c.timeout != "" {
				duration, err := time.ParseDuration(c.timeout)
				if c.expectedErrOccur {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
				ic = &IntegrationConfig{
					Spec: IntegrationConfigSpec{
						IJManageSpec: IntegrationJobManageSpec{
							Timeout: &metav1.Duration{
								Duration: duration,
							},
						},
					},
				}

			}
			require.Equal(t, c.expectedDuration, ic.GetDuration().Duration)
		})
	}
}
