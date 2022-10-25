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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/internal/test"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	clitesting "k8s.io/client-go/testing"
)

const (
	testConfigMapNamespace = "test-ns"
	testConfigMapName      = "test-cm"
)

func TestNewConfigReconciler(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		_, err := NewConfigReconciler(&rest.Config{})
		require.NoError(t, err)
	})

	t.Run("cliErr", func(t *testing.T) {
		_, err := NewConfigReconciler(&rest.Config{QPS: 32, Burst: -1})
		require.Error(t, err)
		require.Equal(t, "burst is required to be greater than 0 when RateLimiter is not set and QPS is set to greater than 0", err.Error())
	})
}

func TestConfigReconciler_Start(t *testing.T) {
	tc := map[string]struct {
		genErr  bool
		cm      *corev1.ConfigMap
		handler configs.Handler

		errorOccurs  bool
		errorMessage string
	}{
		"normal": {
			cm: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: testConfigMapName, ResourceVersion: "1111"},
			},
			handler: func(_ *corev1.ConfigMap) error { return nil },
		},
		"watchErr": {
			genErr:       true,
			errorOccurs:  true,
			errorMessage: "error~~~",
		},
		"reconcileErr": {
			cm: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: testConfigMapName, ResourceVersion: "1111"},
			},
			handler:      func(cm *corev1.ConfigMap) error { return fmt.Errorf("reconcile err") },
			errorOccurs:  true,
			errorMessage: "reconcile err",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			fakeSet := fake.NewSimpleClientset()
			if c.genErr {
				fakeSet.WatchReactionChain = append([]clitesting.WatchReactor{&errWatchReactor{}}, fakeSet.WatchReactionChain...)
			}
			fakeCli := fakeSet.CoreV1().ConfigMaps(testConfigMapNamespace)
			fakeLog := &test.FakeLogger{}
			r := &configReconciler{
				client:               fakeCli,
				log:                  fakeLog,
				lastResourceVersions: map[string]string{},
				handlers:             map[string]configs.Handler{},
			}

			go r.Start()
			time.Sleep(100 * time.Millisecond)
			if c.cm != nil {
				r.handlers[c.cm.Name] = c.handler

				_, err := fakeCli.Create(context.Background(), c.cm, metav1.CreateOptions{})
				require.NoError(t, err)
				time.Sleep(100 * time.Millisecond)
			}

			if c.errorOccurs {
				require.NotEmpty(t, fakeLog.Errors)
				require.Equal(t, c.errorMessage, fakeLog.Errors[0].Error())
			} else {
				require.Empty(t, fakeLog.Errors)
			}
		})
	}
}

type errWatchReactor struct{}

func (e *errWatchReactor) Handles(_ clitesting.Action) bool {
	return true
}

func (e *errWatchReactor) React(_ clitesting.Action) (bool, watch.Interface, error) {
	return true, nil, fmt.Errorf("error~~~")
}

func TestConfigReconciler_Add(t *testing.T) {
	reconciler := &configReconciler{handlers: map[string]configs.Handler{}}
	reconciler.Add("test", func(cm *corev1.ConfigMap) error { return nil })

	handler, exist := reconciler.handlers["test"]
	require.True(t, exist)
	require.NotNil(t, handler)
}

func TestConfigReconciler_Reconcile(t *testing.T) {
	var handled bool
	tc := map[string]struct {
		cm                  *corev1.ConfigMap
		lastResourceVersion string
		handler             func(cm *corev1.ConfigMap) error

		errorOccurs  bool
		errorMessage string
	}{
		"newCM": {
			cm: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: testConfigMapName, Namespace: testConfigMapNamespace, ResourceVersion: "11111"},
			},
			handler: func(cm *corev1.ConfigMap) error {
				handled = true
				return nil
			},
		},
		"cmNil": {},
		"sameResourceVersion": {
			cm: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: testConfigMapName, Namespace: testConfigMapNamespace, ResourceVersion: "11111"},
			},
			lastResourceVersion: "11111",
		},
		"noHandler": {
			cm: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: testConfigMapName, Namespace: testConfigMapNamespace, ResourceVersion: "11111"},
			},
		},
		"handlerErr": {
			cm: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: testConfigMapName, Namespace: testConfigMapNamespace, ResourceVersion: "11111"},
			},
			handler: func(cm *corev1.ConfigMap) error {
				return fmt.Errorf("sample error")
			},
			errorOccurs:  true,
			errorMessage: "sample error",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			handled = false
			fakeCli := fake.NewSimpleClientset().CoreV1().ConfigMaps(testConfigMapNamespace)
			reconciler := &configReconciler{
				client:               fakeCli,
				log:                  &test.FakeLogger{},
				lastResourceVersions: map[string]string{},
				handlers:             map[string]configs.Handler{},
			}
			if c.cm != nil {
				_, err := fakeCli.Create(context.Background(), c.cm, metav1.CreateOptions{})
				require.NoError(t, err)
			}
			if c.handler != nil {
				reconciler.handlers[testConfigMapName] = c.handler
			}
			if c.lastResourceVersion != "" {
				reconciler.lastResourceVersions[testConfigMapName] = c.lastResourceVersion
			}

			err := reconciler.Reconcile(c.cm)
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
				if c.handler != nil {
					require.True(t, handled)
				}
			}
		})
	}
}

func Test_newConfigMapClient(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		_, err := newConfigMapClient(&rest.Config{})
		require.NoError(t, err)
	})

	t.Run("clientErr", func(t *testing.T) {
		_, err := newConfigMapClient(&rest.Config{QPS: 32, Burst: -1})
		require.Error(t, err)
		require.Equal(t, "burst is required to be greater than 0 when RateLimiter is not set and QPS is set to greater than 0", err.Error())
	})
}
