package pipelinemanager

import (
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// PipelineManager implements utility functions for generating, watching PipelineRuns

var log = logf.Log.WithName("pipeline-manager")

func Generate(j *cicdv1.IntegrationJob) (*tektonv1beta1.PipelineRun, error) {
	// TODO
	// Following is a prototype, just for a creation test
	log.Info("Generating a pipeline run")

	return &tektonv1beta1.PipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      Name(j),
			Namespace: j.Namespace,
		},
		Spec: tektonv1beta1.PipelineRunSpec{
			PipelineSpec: &tektonv1beta1.PipelineSpec{
				Description: "dummy",
			},
		},
	}, nil
}

func ReflectStatus(pr *tektonv1beta1.PipelineRun, j *cicdv1.IntegrationJob) error {
	// If PR is nil but IntegrationJob's status is running, set as error
	// Also, schedule next pipelineRun
	if pr == nil && j.Status.State == cicdv1.IntegrationJobStateRunning {
		j.Status.State = cicdv1.IntegrationJobStateFailed
	}

	// If PR exists, default state is running
	if pr != nil {
		j.Status.State = cicdv1.IntegrationJobStateRunning
	}

	// TODO - copy PipelineRun's status to IntegrationJob

	// TODO - Set remote git's commit status for each job

	return nil
}

func Name(j *cicdv1.IntegrationJob) string {
	return j.Name // TODO - should we add postfix or something? Job:PipelineRun=1:1...
}
