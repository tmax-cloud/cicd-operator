package pipelinemanager

import (
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"strconv"
)

// PipelineManager implements utility functions for generating, watching PipelineRuns

var log = logf.Log.WithName("pipeline-manager")

const (
	DefaultWorkingDir = "/tekton/home/integ-source"
)

func Generate(job *cicdv1.IntegrationJob) (*tektonv1beta1.PipelineRun, error) {
	log.Info("Generating a pipeline run")

	// Generate Tasks
	var tasks []tektonv1beta1.PipelineTask
	for _, j := range job.Spec.Jobs {
		taskSpec, err := generateTask(j)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *taskSpec)
	}

	// Fill default env.s
	if err := fillDefaultEnvs(tasks, job); err != nil {
		return nil, err
	}

	return &tektonv1beta1.PipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      Name(job),
			Namespace: job.Namespace,
			Labels:    generateLabel(job),
		},
		Spec: tektonv1beta1.PipelineRunSpec{
			PipelineSpec: &tektonv1beta1.PipelineSpec{
				Tasks: tasks,
			},
		},
	}, nil
}

func generateTask(j cicdv1.Job) (*tektonv1beta1.PipelineTask, error) {
	task := &tektonv1beta1.PipelineTask{Name: j.Name}

	// Steps
	steps, err := generateSteps(j)
	if err != nil {
		return nil, err
	}

	task.TaskSpec = &tektonv1beta1.TaskSpec{Steps: steps}

	// After
	task.RunAfter = append(task.RunAfter, j.After...)

	return task, nil
}

func generateSteps(j cicdv1.Job) ([]tektonv1beta1.Step, error) {
	var steps []tektonv1beta1.Step

	steps = append(steps, gitCheckout())

	step := tektonv1beta1.Step{}
	step.Container = j.Container

	if step.Name != "" {
		step.Name = "step-0"
	}
	if step.WorkingDir == "" {
		step.WorkingDir = DefaultWorkingDir
	}
	step.Script = j.Script
	steps = append(steps, step)
	return steps, nil
}

func generateLabel(j *cicdv1.IntegrationJob) map[string]string {
	label := map[string]string{
		cicdv1.RunLabelJob:   j.Name,
		cicdv1.RunLabelJobId: j.Spec.Id,
		//cicdv1.RunLabelRepository: j.Spec.Refs.Repository,
		//cicdv1.RunLabelSender: j.Spec.Refs.Sender,
	}

	if j.Spec.Refs.Pull != nil {
		label[cicdv1.RunLabelPullRequest] = strconv.Itoa(j.Spec.Refs.Pull.Id)
		label[cicdv1.RunLabelPullRequestSha] = j.Spec.Refs.Pull.Sha
	}

	return label
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

		// TODO - temporary!
		if pr.Status.CompletionTime != nil {
			j.Status.State = cicdv1.IntegrationJobStateCompleted
		}
	}

	// TODO - copy PipelineRun's status to IntegrationJob

	// TODO - Set remote git's commit status for each job

	return nil
}

func Name(j *cicdv1.IntegrationJob) string {
	return j.Name // TODO - should we add postfix or something? Job:PipelineRun=1:1...
}
