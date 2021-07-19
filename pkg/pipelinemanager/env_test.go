package pipelinemanager

import (
	"testing"

	"github.com/bmizerany/assert"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGenerateDefaultEnvs(t *testing.T) {
	jobID1 := utils.RandomString(20)
	ij := &cicdv1.IntegrationJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ij",
			Namespace: "default",
		},
		Spec: cicdv1.IntegrationJobSpec{
			ConfigRef: cicdv1.IntegrationJobConfigRef{
				Name: "test-ic",
				Type: cicdv1.JobTypePreSubmit,
			},
			ID: jobID1,
			Refs: cicdv1.IntegrationJobRefs{
				Repository: "yxzzzxh/test",
				Link:       "https://hub.docker.io/test-repo",
				Base: cicdv1.IntegrationJobRefsBase{
					Ref: "master",
					Sha: "dkfpoekglfjpgarl2p4idmgisq",
				},
				Pulls: []cicdv1.IntegrationJobRefsPull{
					{
						ID:   30,
						Ref:  "bugfix/first",
						Sha:  "0kokpenadiugpowkqe0qlemaogor",
						Link: "first/pull",
						Author: cicdv1.IntegrationJobRefsPullAuthor{
							Name: "Amy",
						},
					},

					{
						ID:   40,
						Ref:  "bugfix/second",
						Sha:  "10292dkkwe9gndlaietpqke204m",
						Link: "second/pull",
						Author: cicdv1.IntegrationJobRefsPullAuthor{
							Name: "Bob",
						},
					},
					{
						ID:   50,
						Ref:  "bugfix/third",
						Sha:  "mbnaohe923jlakmdgokdlo402lk",
						Link: "third/pull",
						Author: cicdv1.IntegrationJobRefsPullAuthor{
							Name: "Chris",
						},
					},
				},
			},
		},
	}

	defaultEnv, err := generateDefaultEnvs(ij)
	assert.Equal(t, nil, err)
	for _, env := range defaultEnv {
		if env.Name == "CI_HEAD_REF" {
			assert.Equal(t, " bugfix/first bugfix/second bugfix/third", env.Value)
		} else if env.Name == "CI_HEAD_SHA" {
			assert.Equal(t, " 0kokpenadiugpowkqe0qlemaogor 10292dkkwe9gndlaietpqke204m mbnaohe923jlakmdgokdlo402lk", env.Value)
		}
	}
}
