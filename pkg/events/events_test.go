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

package events

import (
	"context"
	"testing"

	"github.com/bmizerany/assert"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestEmit(t *testing.T) {
	s := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(s))
	utilruntime.Must(cicdv1.AddToScheme(s))

	// Sample IntegrationJob test cases
	testCases := []*cicdv1.IntegrationJob{
		testCase("t-1", cicdv1.IntegrationJobStatePending, ""),
		testCase("t-2", cicdv1.IntegrationJobStateRunning, "Sample message"),
		testCase("t-3", cicdv1.IntegrationJobStateCompleted, "Test is completed"),
		testCase("t-4", cicdv1.IntegrationJobStateFailed, "Test failed!!"),
	}

	fakeCli := fake.NewClientBuilder().WithScheme(s).Build()

	for _, c := range testCases {
		if err := fakeCli.Create(context.Background(), c); err != nil {
			t.Fatal(err)
		}

		if err := Emit(fakeCli, c, corev1.EventTypeNormal, string(c.Status.State), c.Status.Message); err != nil {
			t.Fatal(err)
		}

		desiredState := string(c.Status.State)
		desiredMessage := c.Status.Message

		resultEvent := getEvent(t, fakeCli)

		assert.Equal(t, desiredState, resultEvent.Reason)
		assert.Equal(t, desiredMessage, resultEvent.Message)

		if err := fakeCli.Delete(context.Background(), resultEvent); err != nil {
			t.Fatal(err)
		}
	}
}

func testCase(name string, state cicdv1.IntegrationJobState, message string) *cicdv1.IntegrationJob {
	return &cicdv1.IntegrationJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "test-ns",
		},
		Spec: cicdv1.IntegrationJobSpec{},
		Status: cicdv1.IntegrationJobStatus{
			State:   state,
			Message: message,
		},
	}
}

func getEvent(t *testing.T, c client.Client) *corev1.Event {
	list := &corev1.EventList{}
	if err := c.List(context.Background(), list); err != nil {
		t.Fatal(err)
	}

	if len(list.Items) != 1 {
		t.Fatal("no event is created")
	}

	return &list.Items[0]
}
