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
	"testing"

	"github.com/stretchr/testify/require"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestNewIntegrationJobReconciler(t *testing.T) {
	s := runtime.NewScheme()
	fakeCli := fake.NewClientBuilder().WithScheme(s).Build()
	logger := &fakeLogger{info: []string{"hi"}}
	reconciler := NewIntegrationJobReconciler(fakeCli, s, logger)

	require.Equal(t, s, reconciler.Scheme)
	require.Equal(t, fakeCli, reconciler.Client)
	require.Equal(t, logger, reconciler.Log)
}

func TestIntegrationJobReconciler_Reconcile(t *testing.T) {
	if _, exist := os.LookupEnv("CI"); !exist {
		ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	}

	s := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(s))
	utilruntime.Must(cicdv1.AddToScheme(s))
	utilruntime.Must(tektonv1beta1.AddToScheme(s))

	sNoCicd := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(sNoCicd))

	sNoTekton := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(sNoTekton))
	utilruntime.Must(cicdv1.AddToScheme(sNoTekton))

	tt := metav1.Now()

	logger := &fakeLogger{}

	tc := map[string]struct {
		ij         *cicdv1.IntegrationJob
		ic         *cicdv1.IntegrationConfig
		pr         *tektonv1beta1.PipelineRun
		dontCreate bool
		scheme     *runtime.Scheme
		key        types.NamespacedName
		verifyFunc func(t *testing.T, job *cicdv1.IntegrationJob)

		errorOccurs  bool
		errorMessage string
	}{
		"successful": {
			ij: &cicdv1.IntegrationJob{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ij", Namespace: "test-ns", Finalizers: []string{finalizer}},
				Spec: cicdv1.IntegrationJobSpec{
					ConfigRef: cicdv1.IntegrationJobConfigRef{Name: "test-ic"},
				},
			},
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ic", Namespace: "test-ns"},
			},
			pr: &tektonv1beta1.PipelineRun{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ij", Namespace: "test-ns"},
			},
			scheme: s,
			key:    types.NamespacedName{Name: "test-ij", Namespace: "test-ns"},
			verifyFunc: func(t *testing.T, ij *cicdv1.IntegrationJob) {
				require.Equal(t, "yes", ij.Annotations["reflected"])
				require.Equal(t, cicdv1.IntegrationJobStatePending, ij.Status.State)
			},
		},
		"notFound": {
			key:        types.NamespacedName{Name: "test-ij", Namespace: "test-ns"},
			scheme:     s,
			verifyFunc: func(t *testing.T, job *cicdv1.IntegrationJob) {},
		},
		"unknown": {
			key:          types.NamespacedName{Name: "test-ij", Namespace: "test-ns"},
			scheme:       sNoCicd,
			errorOccurs:  true,
			errorMessage: "no kind is registered for the type v1.IntegrationJob in scheme \"pkg/runtime/scheme.go:100\"",
		},
		"handleFinalizerError": {
			ij: &cicdv1.IntegrationJob{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ij", Namespace: "test-ns", ResourceVersion: "asd"},
				Spec: cicdv1.IntegrationJobSpec{
					ConfigRef: cicdv1.IntegrationJobConfigRef{Name: "test-ic"},
				},
			},
			scheme: s,
			key:    types.NamespacedName{Name: "test-ij", Namespace: "test-ns"},
			verifyFunc: func(t *testing.T, job *cicdv1.IntegrationJob) {
				require.Len(t, logger.error, 2)
				require.Equal(t, "can not convert resourceVersion \"asd\" to int: strconv.ParseUint: parsing \"asd\": invalid syntax", logger.error[0].Error())
				require.Equal(t, "can not convert resourceVersion \"asd\" to int: strconv.ParseUint: parsing \"asd\": invalid syntax", logger.error[1].Error())
			},
		},
		"handleFinalizerExit": {
			ij: &cicdv1.IntegrationJob{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ij", Namespace: "test-ns"},
				Spec: cicdv1.IntegrationJobSpec{
					ConfigRef: cicdv1.IntegrationJobConfigRef{Name: "test-ic"},
				},
			},
			scheme: s,
			key:    types.NamespacedName{Name: "test-ij", Namespace: "test-ns"},
			verifyFunc: func(t *testing.T, ij *cicdv1.IntegrationJob) {
				require.Len(t, ij.Finalizers, 1)
				require.Equal(t, finalizer, ij.Finalizers[0])
			},
		},
		"alreadyCompleted": {
			ij: &cicdv1.IntegrationJob{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ij", Namespace: "test-ns"},
				Spec: cicdv1.IntegrationJobSpec{
					ConfigRef: cicdv1.IntegrationJobConfigRef{Name: "test-ic"},
				},
				Status: cicdv1.IntegrationJobStatus{
					CompletionTime: &tt,
				},
			},
			scheme:     s,
			key:        types.NamespacedName{Name: "test-ij", Namespace: "test-ns"},
			verifyFunc: func(t *testing.T, ij *cicdv1.IntegrationJob) {},
		},
		"icGetError": {
			ij: &cicdv1.IntegrationJob{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ij", Namespace: "test-ns", Finalizers: []string{finalizer}},
				Spec: cicdv1.IntegrationJobSpec{
					ConfigRef: cicdv1.IntegrationJobConfigRef{Name: "test-ic"},
				},
			},
			scheme: s,
			key:    types.NamespacedName{Name: "test-ij", Namespace: "test-ns"},
			verifyFunc: func(t *testing.T, ij *cicdv1.IntegrationJob) {
				require.Equal(t, cicdv1.IntegrationJobStateFailed, ij.Status.State)
				require.Equal(t, "integrationconfigs.cicd.tmax.io \"test-ic\" not found", ij.Status.Message)
			},
		},
		"prNotFound": {
			ij: &cicdv1.IntegrationJob{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ij", Namespace: "test-ns", Finalizers: []string{finalizer}},
				Spec: cicdv1.IntegrationJobSpec{
					ConfigRef: cicdv1.IntegrationJobConfigRef{Name: "test-ic"},
				},
			},
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ic", Namespace: "test-ns"},
			},
			scheme: s,
			key:    types.NamespacedName{Name: "test-ij", Namespace: "test-ns"},
			verifyFunc: func(t *testing.T, ij *cicdv1.IntegrationJob) {
				require.Equal(t, "yes", ij.Annotations["reflected"])
				require.Equal(t, cicdv1.IntegrationJobStatePending, ij.Status.State)
			},
		},
		"prUnknown": {
			ij: &cicdv1.IntegrationJob{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ij", Namespace: "test-ns", Finalizers: []string{finalizer}},
				Spec: cicdv1.IntegrationJobSpec{
					ConfigRef: cicdv1.IntegrationJobConfigRef{Name: "test-ic"},
				},
			},
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ic", Namespace: "test-ns"},
			},
			scheme: sNoTekton,
			key:    types.NamespacedName{Name: "test-ij", Namespace: "test-ns"},
			verifyFunc: func(t *testing.T, ij *cicdv1.IntegrationJob) {
				require.Equal(t, cicdv1.IntegrationJobStateFailed, ij.Status.State)
				require.Equal(t, "no kind is registered for the type v1beta1.PipelineRun in scheme \"pkg/runtime/scheme.go:100\"", ij.Status.Message)
			},
		},
		"reflectStatusError": {
			ij: &cicdv1.IntegrationJob{
				ObjectMeta: metav1.ObjectMeta{Name: "reflect-fail", Namespace: "test-ns", Finalizers: []string{finalizer}},
				Spec: cicdv1.IntegrationJobSpec{
					ConfigRef: cicdv1.IntegrationJobConfigRef{Name: "test-ic"},
				},
			},
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ic", Namespace: "test-ns"},
			},
			pr: &tektonv1beta1.PipelineRun{
				ObjectMeta: metav1.ObjectMeta{Name: "reflect-fail", Namespace: "test-ns"},
			},
			scheme: s,
			key:    types.NamespacedName{Name: "reflect-fail", Namespace: "test-ns"},
			verifyFunc: func(t *testing.T, ij *cicdv1.IntegrationJob) {
				require.Equal(t, cicdv1.IntegrationJobStateFailed, ij.Status.State)
				require.Equal(t, "expected-error", ij.Status.Message)
			},
		},
		"patchError": {
			ij: &cicdv1.IntegrationJob{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ij", Namespace: "test-ns", Finalizers: []string{finalizer}, ResourceVersion: "asd"},
				Spec: cicdv1.IntegrationJobSpec{
					ConfigRef: cicdv1.IntegrationJobConfigRef{Name: "test-ic"},
				},
			},
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ic", Namespace: "test-ns"},
			},
			pr: &tektonv1beta1.PipelineRun{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ij", Namespace: "test-ns"},
			},
			scheme:       s,
			key:          types.NamespacedName{Name: "test-ij", Namespace: "test-ns"},
			errorOccurs:  true,
			errorMessage: "can not convert resourceVersion \"asd\" to int: strconv.ParseUint: parsing \"asd\": invalid syntax",
		},
	}

	reconciler := &integrationJobReconciler{pm: &fakePipelineManager{}, Log: logger, scheduler: &fakeScheduler{}}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			if c.ij == nil {
				reconciler.Client = fake.NewClientBuilder().WithScheme(c.scheme).Build()
			} else {
				reconciler.Client = fake.NewClientBuilder().WithScheme(c.scheme).WithObjects(c.ij).Build()
			}
			if c.ic != nil {
				_ = reconciler.Client.Create(context.Background(), c.ic)
			}
			if c.pr != nil {
				_ = reconciler.Client.Create(context.Background(), c.pr)
			}

			logger.Clear()

			_, err := reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: c.key})
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)

				result := &cicdv1.IntegrationJob{}
				_ = reconciler.Client.Get(context.Background(), c.key, result)
				c.verifyFunc(t, result)
			}
		})
	}
}

type fakePipelineManager struct{}

func (f *fakePipelineManager) Generate(_ *cicdv1.IntegrationJob) (*tektonv1beta1.PipelineRun, error) {
	return nil, nil
}

func (f *fakePipelineManager) ReflectStatus(_ *tektonv1beta1.PipelineRun, job *cicdv1.IntegrationJob, _ *cicdv1.IntegrationConfig) error {
	if job.Name == "reflect-fail" {
		return fmt.Errorf("expected-error")
	}
	if job.Annotations == nil {
		job.Annotations = map[string]string{}
	}
	job.Annotations["reflected"] = "yes"
	return nil
}

func TestIntegrationJobReconciler_handleFinalizer(t *testing.T) {
	s := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(s))
	utilruntime.Must(cicdv1.AddToScheme(s))

	tt := metav1.Now()

	tc := map[string]struct {
		ijMetaModifier func(meta *metav1.ObjectMeta)

		errorOccurs       bool
		errorMessage      string
		doExit            bool
		expectedFinalizer []string
		deleted           bool
	}{
		"setNow": {
			ijMetaModifier:    func(meta *metav1.ObjectMeta) {},
			doExit:            true,
			expectedFinalizer: []string{"cicd.tmax.io/finalizer"},
		},
		"setPatchError": {
			ijMetaModifier: func(meta *metav1.ObjectMeta) {
				meta.Name = "another"
			},
			errorOccurs:  true,
			errorMessage: "integrationjobs.cicd.tmax.io \"another\" not found",
		},
		"alreadySet": {
			ijMetaModifier: func(meta *metav1.ObjectMeta) {
				meta.Finalizers = []string{"cicd.tmax.io/finalizer"}
			},
			doExit:            false,
			expectedFinalizer: []string{"cicd.tmax.io/finalizer"},
		},
		"deletedOnlyFinalizer": {
			ijMetaModifier: func(meta *metav1.ObjectMeta) {
				meta.Finalizers = []string{"cicd.tmax.io/finalizer"}
				meta.DeletionTimestamp = &tt
			},
			doExit:            true,
			expectedFinalizer: nil,
			deleted:           true,
		},
		"deletedMultipleFinalizers": {
			ijMetaModifier: func(meta *metav1.ObjectMeta) {
				meta.Finalizers = []string{"cicd.tmax.io/finalizer", "another/finalizer"}
				meta.DeletionTimestamp = &tt
			},
			doExit:            true,
			expectedFinalizer: []string{"another/finalizer"},
		},
		"deletedPatchError": {
			ijMetaModifier: func(meta *metav1.ObjectMeta) {
				meta.Name = "another"
				meta.Finalizers = []string{"cicd.tmax.io/finalizer", "another/finalizer"}
				meta.DeletionTimestamp = &tt
			},
			errorOccurs:  true,
			errorMessage: "integrationjobs.cicd.tmax.io \"another\" not found",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			reconciler := &integrationJobReconciler{}
			original := &cicdv1.IntegrationJob{ObjectMeta: metav1.ObjectMeta{Name: "test-ij", Namespace: "test-ns"}}

			logger := &fakeLogger{}
			reconciler.Client = fake.NewClientBuilder().WithScheme(s).WithObjects(original).Build()
			reconciler.Log = logger
			reconciler.scheduler = &fakeScheduler{}

			ij := original.DeepCopy()
			c.ijMetaModifier(&ij.ObjectMeta)
			if original.Name == ij.Name {
				require.NoError(t, reconciler.Client.Update(context.Background(), ij))
			}
			doExit, err := reconciler.handleFinalizer(ij.DeepCopy(), ij)
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.Equal(t, c.doExit, doExit)

				if !c.deleted {
					result := &cicdv1.IntegrationJob{}
					require.NoError(t, reconciler.Client.Get(context.Background(), types.NamespacedName{Name: "test-ij", Namespace: "test-ns"}, result))

					require.Equal(t, c.expectedFinalizer, result.Finalizers)
				}
			}
		})
	}
}

type fakeScheduler struct{}

func (f *fakeScheduler) Notify(_ *cicdv1.IntegrationJob) {}

func TestIntegrationJobReconciler_patchJobFailed(t *testing.T) {
	s := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(s))
	utilruntime.Must(cicdv1.AddToScheme(s))

	original := &cicdv1.IntegrationJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ij",
			Namespace: "test-ns",
		},
	}

	reconciler := &integrationJobReconciler{}

	tc := map[string]struct {
		ijModifier func(ij *cicdv1.IntegrationJob)
		message    string

		errorOccurs  bool
		errorMessage string
	}{
		"success": {
			ijModifier: func(ij *cicdv1.IntegrationJob) {
				ij.Spec.ID = "test-id"
			},
			message: "test-msg",
		},
		"fail": {
			ijModifier: func(ij *cicdv1.IntegrationJob) {
				ij.Name = "another-name"
				ij.Spec.ID = "test-id"
			},
			message:      "test-msg",
			errorOccurs:  true,
			errorMessage: "integrationjobs.cicd.tmax.io \"another-name\" not found",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			logger := &fakeLogger{}
			reconciler.Client = fake.NewClientBuilder().WithScheme(s).WithObjects(original).Build()
			reconciler.Log = logger

			ij := original.DeepCopy()
			c.ijModifier(ij)

			reconciler.patchJobFailed(ij, original, c.message)
			if c.errorOccurs {
				require.Len(t, logger.error, 1)
				require.Equal(t, c.errorMessage, logger.error[0].Error())
			} else {
				result := &cicdv1.IntegrationJob{}
				require.NoError(t, reconciler.Client.Get(context.Background(), types.NamespacedName{Name: "test-ij", Namespace: "test-ns"}, result))

				require.Equal(t, cicdv1.IntegrationJobStateFailed, result.Status.State)
				require.Equal(t, c.message, result.Status.Message)
			}
		})
	}
}

func TestIntegrationJobReconciler_SetupWithManager(t *testing.T) {
	s := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(s))
	utilruntime.Must(cicdv1.AddToScheme(s))

	mgr := &fakeManager{scheme: s}

	reconciler := &integrationJobReconciler{}
	require.NoError(t, reconciler.SetupWithManager(mgr))
}
