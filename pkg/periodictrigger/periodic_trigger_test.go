package periodictrigger

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	v1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/cron"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type syncTestCase struct {
	ic *cicdv1.IntegrationConfig
}

func Test_sync(t *testing.T) {
	tc := map[string]syncTestCase{
		"Job": {
			ic: &v1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "test-config",
				},
				Spec: v1.IntegrationConfigSpec{
					Jobs: v1.IntegrationConfigJobs{
						Periodic: []v1.Periodic{
							{
								Job: v1.Job{
									Container: corev1.Container{
										Name: "test-cron",
									},
								},
								Cron: "@every 1m",
							},
						},
					},
				},
			},
		},
	}
	s := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(s))
	utilruntime.Must(cicdv1.AddToScheme(s))

	fakeCli := fake.NewClientBuilder().WithScheme(s).Build()

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			start := time.Now()
			cron := cron.New()
			err := sync(fakeCli, context.Background(), c.ic, cron, start)
			require.Equal(t, err, nil)
		})
	}
}
