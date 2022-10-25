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

	"github.com/stretchr/testify/require"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/test"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestCustomRunReconciler_AddKindHandler(t *testing.T) {
	reconciler := &CustomRunReconciler{
		KindHandlerMap:  map[string]KindHandler{},
		HandlerChildren: map[string][]client.Object{},
	}
	reconciler.AddKindHandler("test", &testKindHandler{}, &corev1.ConfigMap{})

	require.NotNil(t, reconciler.KindHandlerMap["test"])
	require.Len(t, reconciler.HandlerChildren["test"], 1)
	require.Equal(t, &corev1.ConfigMap{}, reconciler.HandlerChildren["test"][0])
}

func TestCustomRunReconciler_Reconcile(t *testing.T) {
	s := runtime.NewScheme()
	require.NoError(t, tektonv1alpha1.AddToScheme(s))

	tc := map[string]struct {
		run     *tektonv1alpha1.Run
		name    string
		scheme  *runtime.Scheme
		handler KindHandler

		errorOccurs  bool
		errorMessage string
	}{
		"normal": {
			run: &tektonv1alpha1.Run{
				ObjectMeta: metav1.ObjectMeta{Name: "test-run", Namespace: "test-ns"},
				Spec: tektonv1alpha1.RunSpec{
					Ref: &tektonv1alpha1.TaskRef{
						APIVersion: cicdv1.CustomTaskAPIVersion,
						Kind:       "TestKind",
					},
				},
			},
			name:    "test-run",
			scheme:  s,
			handler: &testKindHandler{},
		},
		"notFound": {
			name:    "test-run",
			scheme:  s,
			handler: &testKindHandler{},
		},
		"getErr": {
			name:         "test-run",
			scheme:       runtime.NewScheme(),
			handler:      &testKindHandler{},
			errorOccurs:  true,
			errorMessage: "no kind is registered for the type v1alpha1.Run in scheme \"pkg/runtime/scheme.go:100\"",
		},
		"notCiCdRun": {
			run: &tektonv1alpha1.Run{
				ObjectMeta: metav1.ObjectMeta{Name: "test-run", Namespace: "test-ns"},
				Spec: tektonv1alpha1.RunSpec{
					Ref: &tektonv1alpha1.TaskRef{
						APIVersion: "another.api/v1",
					},
				},
			},
			name:    "test-run",
			scheme:  s,
			handler: &testKindHandler{},
		},
		"noHandler": {
			run: &tektonv1alpha1.Run{
				ObjectMeta: metav1.ObjectMeta{Name: "test-run", Namespace: "test-ns"},
				Spec: tektonv1alpha1.RunSpec{
					Ref: &tektonv1alpha1.TaskRef{
						APIVersion: cicdv1.CustomTaskAPIVersion,
						Kind:       "TestKind",
					},
				},
			},
			name:   "test-run",
			scheme: s,
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			fakeCli := fake.NewClientBuilder().WithScheme(c.scheme).Build()
			if c.run != nil {
				require.NoError(t, fakeCli.Create(context.Background(), c.run))
			}

			reconciler := &CustomRunReconciler{
				Client:          fakeCli,
				Log:             &test.FakeLogger{},
				Scheme:          c.scheme,
				KindHandlerMap:  map[string]KindHandler{},
				HandlerChildren: map[string][]client.Object{},
			}
			if c.handler != nil {
				reconciler.KindHandlerMap["TestKind"] = c.handler
			}

			_, err := reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: types.NamespacedName{Name: c.name, Namespace: "test-ns"}})
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

type testKindHandler struct{}

func (t *testKindHandler) Handle(_ *tektonv1alpha1.Run) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func TestCustomRunReconciler_SetupWithManager(t *testing.T) {
	s := runtime.NewScheme()
	utilruntime.Must(tektonv1alpha1.AddToScheme(s))

	mgr := &fakeManager{scheme: s}

	reconciler := &CustomRunReconciler{
		HandlerChildren: map[string][]client.Object{
			"test": {&corev1.Pod{}},
		},
	}
	require.NoError(t, reconciler.SetupWithManager(mgr))
}
