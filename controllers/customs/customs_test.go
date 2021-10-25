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

package customs

import (
	"testing"

	"github.com/stretchr/testify/require"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_compileTitleContent(t *testing.T) {
	tc := map[string]struct {
		ijName string
		ijSpec cicdv1.IntegrationJobSpec
		title  string

		errorOccurs    bool
		errorMessage   string
		expectedOutput string
	}{
		"normalTag": {
			ijName: "test-ij-email",
			ijSpec: cicdv1.IntegrationJobSpec{
				Refs: cicdv1.IntegrationJobRefs{
					Base: cicdv1.IntegrationJobRefsBase{
						Ref: cicdv1.GitRef("refs/tags/v0.2.3"),
					},
				},
				Jobs: []cicdv1.Job{{Container: corev1.Container{Name: "test-job"}}},
			},
			title:          "Release Version {{ .Spec.Refs.Base.Ref.GetTag }}",
			expectedOutput: "Release Version v0.2.3",
		},
		"normalBranch": {
			ijName: "test-ij-email",
			ijSpec: cicdv1.IntegrationJobSpec{
				Refs: cicdv1.IntegrationJobRefs{
					Base: cicdv1.IntegrationJobRefsBase{
						Ref: cicdv1.GitRef("refs/heads/new-feat"),
					},
				},
				Jobs: []cicdv1.Job{{Container: corev1.Container{Name: "test-job"}}},
			},
			title:          "Release Branch {{ .Spec.Refs.Base.Ref.GetBranch }}",
			expectedOutput: "Release Branch new-feat",
		},
		"content": {
			ijName: "test-ij-email",
			ijSpec: cicdv1.IntegrationJobSpec{
				Refs: cicdv1.IntegrationJobRefs{
					Base: cicdv1.IntegrationJobRefsBase{
						Ref: cicdv1.GitRef("refs/heads/new-feat"),
					},
				},
				Jobs: []cicdv1.Job{{Container: corev1.Container{Name: "test-job"}}},
			},
			title:          "$INTEGRATION_JOB_NAME $JOB_NAME",
			expectedOutput: "test-ij-email test-job",
		},
		"noIJ": {
			ijName:       "test-ij-email-no",
			ijSpec:       cicdv1.IntegrationJobSpec{Jobs: []cicdv1.Job{{Container: corev1.Container{Name: "test-job"}}}},
			title:        "$INTEGRATION_JOB_NAME $JOB_NAME",
			errorOccurs:  true,
			errorMessage: "integrationjobs.cicd.tmax.io \"test-ij-email-no\" not found",
		},
		"parseErr": {
			ijName:       "test-ij-email",
			ijSpec:       cicdv1.IntegrationJobSpec{Jobs: []cicdv1.Job{{Container: corev1.Container{Name: "test-job"}}}},
			title:        "Release Branch {{ ..Spec.Refs.Base.Ref.GetBranch }}",
			errorOccurs:  true,
			errorMessage: "template: contentTemplate:1: unexpected . after term \".\"",
		},
		"executeError": {
			ijName:       "test-ij-email",
			ijSpec:       cicdv1.IntegrationJobSpec{Jobs: []cicdv1.Job{{Container: corev1.Container{Name: "test-job"}}}},
			title:        "Release Branch {{ .Spec.Refs.Base.Refffff }}",
			errorOccurs:  true,
			errorMessage: "template: contentTemplate:1:23: executing \"contentTemplate\" at <.Spec.Refs.Base.Refffff>: can't evaluate field Refffff in type v1.IntegrationJobRefsBase",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			s := runtime.NewScheme()
			utilruntime.Must(cicdv1.AddToScheme(s))
			utilruntime.Must(tektonv1beta1.AddToScheme(s))
			utilruntime.Must(tektonv1alpha1.AddToScheme(s))

			ij := &cicdv1.IntegrationJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ij-email",
					Namespace: "default",
				},
				Spec: c.ijSpec,
			}

			cli := fake.NewClientBuilder().WithScheme(s).WithObjects(ij).Build()
			output, err := compileString("default", c.ijName, ij.Spec.Jobs[0].Name, c.title, cli)
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, c.expectedOutput, output)
			}
		})
	}
}
