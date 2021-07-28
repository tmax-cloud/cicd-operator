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
	"context"
	"github.com/bmizerany/assert"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func Test_compileTitleContent(t *testing.T) {
	s := runtime.NewScheme()
	utilruntime.Must(cicdv1.AddToScheme(s))
	utilruntime.Must(tektonv1beta1.AddToScheme(s))
	utilruntime.Must(tektonv1alpha1.AddToScheme(s))

	ij := &cicdv1.IntegrationJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ij-email",
			Namespace: "default",
		},
		Spec: cicdv1.IntegrationJobSpec{
			Refs: cicdv1.IntegrationJobRefs{
				Base: cicdv1.IntegrationJobRefsBase{
					Ref: cicdv1.GitRef("refs/tags/v0.2.3"),
				},
			},
			Jobs: []cicdv1.Job{{}},
		},
	}
	ij.Spec.Jobs[0].Name = "test-job"

	cli := fake.NewFakeClientWithScheme(s, ij)

	// TEST - Title1
	title := "Release Version {{ .Spec.Refs.Base.Ref.GetTag }}"
	compiledTitle, err := compileString(ij.Namespace, ij.Name, ij.Spec.Jobs[0].Name, title, cli)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "Release Version v0.2.3", compiledTitle)

	// TEST - Title2
	ij.Spec.Refs.Base.Ref = "refs/heads/new-feat"
	_ = cli.Update(context.Background(), ij)
	title = "Release Branch {{ .Spec.Refs.Base.Ref.GetBranch }}"
	compiledTitle, err = compileString(ij.Namespace, ij.Name, ij.Spec.Jobs[0].Name, title, cli)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "Release Branch new-feat", compiledTitle)

	// TEST - Content1
	content := "$INTEGRATION_JOB_NAME $JOB_NAME"
	compiledContent, err := compileString(ij.Namespace, ij.Name, ij.Spec.Jobs[0].Name, content, cli)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "test-ij-email test-job", compiledContent)
}
