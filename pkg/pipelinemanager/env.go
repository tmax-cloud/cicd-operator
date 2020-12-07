package pipelinemanager

import (
	"fmt"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"net/url"
)

func fillDefaultEnvs(tasks []tektonv1beta1.PipelineTask, job *cicdv1.IntegrationJob) error {
	defaultEnvs, err := generateDefaultEnvs(job)
	if err != nil {
		return err
	}

	// Index-based loop, because Go creates a copy when iterating
	for i := range tasks {
		if tasks[i].TaskSpec == nil {
			continue
		}
		steps := tasks[i].TaskSpec.Steps
		for j := range steps {
			// Default values are prepended, not appended - they might be used from the other env.s
			steps[j].Env = append(defaultEnvs, steps[j].Env...)
		}
	}

	return nil
}

func generateDefaultEnvs(job *cicdv1.IntegrationJob) ([]corev1.EnvVar, error) {
	jobSpec := job.Spec
	u, err := url.Parse(jobSpec.Refs.Link)
	if err != nil {
		return nil, err
	}
	defaultEnvs := []corev1.EnvVar{
		{Name: "CI", Value: "true"},
		{Name: "CI_CONFIG_NAME", Value: jobSpec.ConfigRef.Name},
		{Name: "CI_JOB_ID", Value: jobSpec.Id},
		{Name: "CI_REPOSITORY", Value: jobSpec.Refs.Repository},
		{Name: "CI_SERVER_URL", Value: fmt.Sprintf("%s://%s", u.Scheme, u.Host)},
		{Name: "CI_API_URL", Value: ""},                                // TODO
		{Name: "CI_EVENT_TYPE", Value: string(jobSpec.ConfigRef.Type)}, // TODO
		{Name: "CI_EVENT_PATH", Value: ""},                             // TODO
	}

	refs := jobSpec.Refs
	if refs.Pull == nil {
		// Push event
		defaultEnvs = append(defaultEnvs, corev1.EnvVar{
			Name: "CI_HEAD_REF", Value: refs.Base.Ref,
		})
	} else {
		// Pull Request event
		defaultEnvs = append(defaultEnvs, []corev1.EnvVar{
			{Name: "CI_HEAD_SHA", Value: refs.Pull.Sha},
			{Name: "CI_HEAD_REF", Value: refs.Pull.Ref},
			{Name: "CI_BASE_REF", Value: refs.Base.Ref},
		}...)
	}

	return defaultEnvs, nil
}
