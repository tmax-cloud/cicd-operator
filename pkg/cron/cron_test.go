package cron

import (
	"testing"

	"github.com/stretchr/testify/require"
	v1 "github.com/tmax-cloud/cicd-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
)

type SyncIntegrationConfigTestCase struct {
	ic *v1.IntegrationConfig

	expectedJobsNames []string
}

func TestSyncIntegrationConfig(t *testing.T) {
	tc := map[string]SyncIntegrationConfigTestCase{
		"initial": {
			ic: &v1.IntegrationConfig{
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
			expectedJobsNames: []string{"test-cron"},
		},
		"addAndUpdate": {
			ic: &v1.IntegrationConfig{
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
							{
								Job: v1.Job{
									Container: corev1.Container{
										Name: "test-cron-2",
									},
								},
								Cron: "@every 1m",
							},
						},
					},
				},
			},
			expectedJobsNames: []string{"test-cron", "test-cron-2"},
		},
		"deleted": {
			ic: &v1.IntegrationConfig{
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
			expectedJobsNames: []string{"test-cron"},
		},
	}

	cr := New()
	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			err := cr.SyncIntegrationConfig(c.ic)
			require.NoError(t, err)

			for _, jobName := range c.expectedJobsNames {
				exist := cr.HasJob(jobName)
				require.Equal(t, exist, true)
			}
		})
	}
}
