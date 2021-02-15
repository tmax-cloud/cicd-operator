package events

import (
	"context"
	"github.com/bmizerany/assert"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
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

	fakeCli := fake.NewFakeClientWithScheme(s)

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
