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

package utils

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestNamespace(t *testing.T) {
	tc := map[string]struct {
		path string
		data []byte
		env  string

		expectedNs string
	}{
		"inCluster": {
			path:       path.Join(os.TempDir(), "test-cicd-operator-ns-file-in-cluster"),
			data:       []byte("cicd-system3"),
			expectedNs: "cicd-system3",
		},
		"env": {
			path:       path.Join(os.TempDir(), "test-cicd-operator-ns-file-env"),
			env:        "cicd-system2",
			expectedNs: "cicd-system2",
		},
		"default": {
			path:       path.Join(os.TempDir(), "test-cicd-operator-ns-file-default"),
			expectedNs: "cicd-system",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			nsFilePath = c.path
			if c.path != "" && len(c.data) > 0 {
				require.NoError(t, os.MkdirAll(path.Dir(c.path), os.ModePerm))
				require.NoError(t, ioutil.WriteFile(c.path, c.data, os.ModePerm))
				defer func() {
					_ = os.RemoveAll(c.path)
				}()
			}

			if c.env != "" {
				require.NoError(t, os.Setenv(nsEnv, c.env))
				defer func() {
					_ = os.Unsetenv(nsEnv)
				}()
			}

			ns := Namespace()

			require.Equal(t, c.expectedNs, ns)
		})
	}
}

func TestCreateOrPatchObject(t *testing.T) {
	s := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(s))
	require.NoError(t, cicdv1.AddToScheme(s))

	tc := map[string]struct {
		obj      client.Object
		original client.Object
		parent   client.Object
		scheme   *runtime.Scheme
		objExist bool

		errorOccurs  bool
		errorMessage string
	}{
		"create": {
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "test-pod", Namespace: "test-ns"},
			},
			scheme: scheme.Scheme,
		},
		"createErr": {
			obj: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "test-pod", Namespace: "test-ns"},
			},
			scheme:       scheme.Scheme,
			errorOccurs:  true,
			errorMessage: "no kind is registered for the type v1.IntegrationConfig in scheme \"pkg/runtime/scheme.go:100\"",
		},
		"patch": {
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "test-pod", Namespace: "test-ns", ResourceVersion: "1111"},
			},
			original: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "test-pod", Namespace: "test-ns", ResourceVersion: "1111"},
			},
			objExist: true,
			scheme:   scheme.Scheme,
		},
		"patchErr": {
			obj: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{ResourceVersion: "1111"},
			},
			original: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{ResourceVersion: "1111"},
			},
			scheme:       scheme.Scheme,
			errorOccurs:  true,
			errorMessage: "no kind is registered for the type v1.IntegrationConfig in scheme \"pkg/runtime/scheme.go:100\"",
		},
		"patchOriginalNil": {
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{ResourceVersion: "1111"},
			},
			errorOccurs:  true,
			errorMessage: "original object exists but not passed",
		},
		"setConRefErr": {
			obj:          &corev1.Pod{},
			parent:       &cicdv1.IntegrationConfig{},
			scheme:       scheme.Scheme,
			errorOccurs:  true,
			errorMessage: "no kind is registered for the type v1.IntegrationConfig in scheme \"pkg/runtime/scheme.go:100\"",
		},
		"objNil": {
			errorOccurs:  true,
			errorMessage: "obj cannot be nil",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			var fakeCli client.Client
			if c.objExist {
				fakeCli = fake.NewClientBuilder().WithScheme(c.scheme).WithObjects(c.obj).Build()
			} else {
				fakeCli = fake.NewClientBuilder().WithScheme(c.scheme).Build()
			}

			err := CreateOrPatchObject(c.obj, c.original, c.parent, fakeCli, c.scheme)
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
