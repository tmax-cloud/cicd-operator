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
	"testing"

	"github.com/stretchr/testify/require"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGenerateDefaultEnvs(t *testing.T) {
	tc := map[string]struct {
		jobSpec cicdv1.IntegrationJobSpec

		errorOccurs  bool
		errorMessage string
		expectedVars []corev1.EnvVar
	}{
		"urlError": {
			jobSpec: cicdv1.IntegrationJobSpec{
				Refs: cicdv1.IntegrationJobRefs{Link: "https://192.168.0.%31/"},
			},
			errorOccurs:  true,
			errorMessage: "parse \"https://192.168.0.%31/\": invalid URL escape \"%31\"",
		},
		"nilSender": {
			jobSpec: cicdv1.IntegrationJobSpec{
				ConfigRef: cicdv1.IntegrationJobConfigRef{Name: "test-ic", Type: cicdv1.JobTypePreSubmit},
				ID:        "dummy-id",
				Refs: cicdv1.IntegrationJobRefs{
					Repository: "yxzzzxh/test",
					Link:       "https://hub.docker.io/test-repo",
					Base:       cicdv1.IntegrationJobRefsBase{Ref: "master", Sha: "dkfpoekglfjpgarl2p4idmgisq"},
				},
			},
			expectedVars: []corev1.EnvVar{
				{Name: "CI_HEAD_SHA", Value: "dkfpoekglfjpgarl2p4idmgisq"},
				{Name: "CI_HEAD_REF", Value: "master"},
			},
		},
		"pull": {
			jobSpec: cicdv1.IntegrationJobSpec{
				ConfigRef: cicdv1.IntegrationJobConfigRef{Name: "test-ic", Type: cicdv1.JobTypePreSubmit},
				ID:        "dummy-id",
				Refs: cicdv1.IntegrationJobRefs{
					Repository: "yxzzzxh/test",
					Link:       "https://hub.docker.io/test-repo",
					Base:       cicdv1.IntegrationJobRefsBase{Ref: "master", Sha: "dkfpoekglfjpgarl2p4idmgisq"},
					Sender:     &cicdv1.IntegrationJobSender{Name: "test-user", Email: "test-user@test.com"},
					Pulls: []cicdv1.IntegrationJobRefsPull{
						{
							ID:     30,
							Ref:    "bugfix/first",
							Sha:    "0kokpenadiugpowkqe0qlemaogor",
							Link:   "first/pull",
							Author: cicdv1.IntegrationJobRefsPullAuthor{Name: "Amy"},
						},
					},
				},
			},
			expectedVars: []corev1.EnvVar{
				{Name: "CI_SENDER_NAME", Value: "test-user"},
				{Name: "CI_SENDER_EMAIL", Value: "test-user@test.com"},
				{Name: "CI_HEAD_SHA", Value: " 0kokpenadiugpowkqe0qlemaogor"},
				{Name: "CI_HEAD_REF", Value: " bugfix/first"},
				{Name: "CI_BASE_SHA", Value: "dkfpoekglfjpgarl2p4idmgisq"},
				{Name: "CI_BASE_REF", Value: "master"},
			},
		},
		"multiPull": {
			jobSpec: cicdv1.IntegrationJobSpec{
				ConfigRef: cicdv1.IntegrationJobConfigRef{Name: "test-ic", Type: cicdv1.JobTypePreSubmit},
				ID:        "dummy-id",
				Refs: cicdv1.IntegrationJobRefs{
					Repository: "yxzzzxh/test",
					Link:       "https://hub.docker.io/test-repo",
					Base:       cicdv1.IntegrationJobRefsBase{Ref: "master", Sha: "dkfpoekglfjpgarl2p4idmgisq"},
					Sender:     &cicdv1.IntegrationJobSender{Name: "test-user", Email: "test-user@test.com"},
					Pulls: []cicdv1.IntegrationJobRefsPull{
						{
							ID:     30,
							Ref:    "bugfix/first",
							Sha:    "0kokpenadiugpowkqe0qlemaogor",
							Link:   "first/pull",
							Author: cicdv1.IntegrationJobRefsPullAuthor{Name: "Amy"},
						},
						{
							ID:     40,
							Ref:    "bugfix/second",
							Sha:    "10292dkkwe9gndlaietpqke204m",
							Link:   "second/pull",
							Author: cicdv1.IntegrationJobRefsPullAuthor{Name: "Bob"},
						},
						{
							ID:     50,
							Ref:    "bugfix/third",
							Sha:    "mbnaohe923jlakmdgokdlo402lk",
							Link:   "third/pull",
							Author: cicdv1.IntegrationJobRefsPullAuthor{Name: "Chris"},
						},
					},
				},
			},
			expectedVars: []corev1.EnvVar{
				{Name: "CI_SENDER_NAME", Value: "test-user"},
				{Name: "CI_SENDER_EMAIL", Value: "test-user@test.com"},
				{Name: "CI_HEAD_SHA", Value: " 0kokpenadiugpowkqe0qlemaogor 10292dkkwe9gndlaietpqke204m mbnaohe923jlakmdgokdlo402lk"},
				{Name: "CI_HEAD_REF", Value: " bugfix/first bugfix/second bugfix/third"},
				{Name: "CI_BASE_SHA", Value: "dkfpoekglfjpgarl2p4idmgisq"},
				{Name: "CI_BASE_REF", Value: "master"},
			},
		},
		"push": {
			jobSpec: cicdv1.IntegrationJobSpec{
				ConfigRef: cicdv1.IntegrationJobConfigRef{Name: "test-ic", Type: cicdv1.JobTypePreSubmit},
				ID:        "dummy-id",
				Refs: cicdv1.IntegrationJobRefs{
					Repository: "yxzzzxh/test",
					Link:       "https://hub.docker.io/test-repo",
					Base:       cicdv1.IntegrationJobRefsBase{Ref: "master", Sha: "dkfpoekglfjpgarl2p4idmgisq"},
					Sender:     &cicdv1.IntegrationJobSender{Name: "test-user", Email: "test-user@test.com"},
				},
			},
			expectedVars: []corev1.EnvVar{
				{Name: "CI_SENDER_NAME", Value: "test-user"},
				{Name: "CI_SENDER_EMAIL", Value: "test-user@test.com"},
				{Name: "CI_HEAD_SHA", Value: "dkfpoekglfjpgarl2p4idmgisq"},
				{Name: "CI_HEAD_REF", Value: "master"},
			},
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			ij := &cicdv1.IntegrationJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ij",
					Namespace: "default",
				},
				Spec: c.jobSpec,
			}

			defaultEnv, err := generateDefaultEnvs(ij)
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)

				u, err := url.Parse(c.jobSpec.Refs.Link)
				require.NoError(t, err)
				expectedVars := append([]corev1.EnvVar{
					{Name: "CI", Value: "true"},
					{Name: "CI_CONFIG_NAME", Value: c.jobSpec.ConfigRef.Name},
					{Name: "CI_JOB_ID", Value: c.jobSpec.ID},
					{Name: "CI_REPOSITORY", Value: c.jobSpec.Refs.Repository},
					{Name: "CI_EVENT_TYPE", Value: string(c.jobSpec.ConfigRef.Type)},
					{Name: "CI_WORKSPACE", Value: DefaultWorkingDir},
					{Name: "CI_SERVER_URL", Value: fmt.Sprintf("%s://%s", u.Scheme, u.Host)},
				}, c.expectedVars...)
				require.Equal(t, expectedVars, defaultEnv)
			}
		})
	}
}
