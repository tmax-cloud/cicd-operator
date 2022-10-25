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

package pipelinemanager

import (
	"fmt"
	"net/url"

	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
)

func fillDefaultEnvs(tasks []tektonv1beta1.PipelineTask, job *cicdv1.IntegrationJob) error {
	// Index-based loop, because Go creates a copy when iterating
	for i := range tasks {
		if tasks[i].TaskSpec == nil {
			continue
		}
		steps := tasks[i].TaskSpec.Steps
		for j := range steps {
			defaultEnvs, err := generateDefaultEnvs(job)
			if err != nil {
				return err
			}
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
		{Name: "CI_JOB_ID", Value: jobSpec.ID},
		{Name: "CI_REPOSITORY", Value: jobSpec.Refs.Repository},
		{Name: "CI_EVENT_TYPE", Value: string(jobSpec.ConfigRef.Type)},
		{Name: "CI_WORKSPACE", Value: DefaultWorkingDir},
		{Name: "CI_SERVER_URL", Value: fmt.Sprintf("%s://%s", u.Scheme, u.Host)},
	}

	ann := job.ObjectMeta.Annotations
	if reqbody, ok := ann["requestBody"]; ok {
		defaultEnvs = append(defaultEnvs, []corev1.EnvVar{
			{Name: "REQUEST_BODY", Value: reqbody},
		}...)
	}
	// Sender
	if jobSpec.Refs.Sender != nil {
		defaultEnvs = append(defaultEnvs, []corev1.EnvVar{
			{Name: "CI_SENDER_NAME", Value: jobSpec.Refs.Sender.Name},
			{Name: "CI_SENDER_EMAIL", Value: jobSpec.Refs.Sender.Email},
		}...)
	}

	refs := jobSpec.Refs
	if refs.Pulls == nil {
		// Push event
		defaultEnvs = append(defaultEnvs, []corev1.EnvVar{
			{Name: "CI_HEAD_SHA", Value: refs.Base.Sha},
			{Name: "CI_HEAD_REF", Value: refs.Base.Ref.String()},
		}...)
	} else {
		// Pull Request event
		// For many PRs, head-related environment variables(CI_HEAD_SHA, CI_HEAD_REF)
		// are set to a single string separated by white-spaces(" ").
		var headShas string
		var headRefs string
		for _, pull := range refs.Pulls {
			headShas = headShas + " " + pull.Sha
			headRefs = headRefs + " " + pull.Ref.String()
		}

		defaultEnvs = append(defaultEnvs, []corev1.EnvVar{
			{Name: "CI_HEAD_SHA", Value: headShas},
			{Name: "CI_HEAD_REF", Value: headRefs},
			{Name: "CI_BASE_SHA", Value: refs.Base.Sha},
			{Name: "CI_BASE_REF", Value: refs.Base.Ref.String()},
		}...)
	}

	return defaultEnvs, nil
}
