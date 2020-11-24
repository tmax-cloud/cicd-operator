package pipelinemanager

import (
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
)

// PipelineManager implements utility functions for generating, watching PipelineRuns

func Generate(_ *cicdv1.IntegrationJob) (*tektonv1beta1.PipelineRun, error) {
	// TODO
	return &tektonv1beta1.PipelineRun{}, nil
}

func Name(j *cicdv1.IntegrationJob) string {
	return j.Name // TODO - should we add some prefix/suffix...?
}
