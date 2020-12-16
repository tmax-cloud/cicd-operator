package pipelinemanager

import (
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
			ServiceAccountName: cicdv1.GetServiceAccountName(job.Spec.ConfigRef.Name),
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

	task.TaskSpec = &tektonv1beta1.EmbeddedTask{}
	task.TaskSpec.Steps = steps

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

func ReflectStatus(pr *tektonv1beta1.PipelineRun, job *cicdv1.IntegrationJob, cfg *cicdv1.IntegrationConfig, client client.Client) error {
	// If PR is nil but IntegrationJob's status is running, set as error
	// Also, schedule next pipelineRun
	if pr == nil && job.Status.State == cicdv1.IntegrationJobStateRunning {
		job.Status.State = cicdv1.IntegrationJobStateFailed
	}

	// Initialize status.jobs
	stateChanged := make([]bool, len(job.Spec.Jobs))
	reset := len(job.Status.Jobs) != len(job.Spec.Jobs)
	if reset {
		job.Status.Jobs = nil
	}
	for _, j := range job.Spec.Jobs {
		if reset {
			job.Status.Jobs = append(job.Status.Jobs, cicdv1.JobStatus{
				Name:  j.Name,
				State: cicdv1.CommitStatusStatePending,
			})
		}
		stateChanged = append(stateChanged, reset)
	}

	// If PR exists, default state is running
	if pr != nil {
		job.Status.State = cicdv1.IntegrationJobStateRunning

		job.Status.StartTime = pr.CreationTimestamp.DeepCopy()
		job.Status.CompletionTime = pr.Status.CompletionTime.DeepCopy()

		if job.Status.CompletionTime != nil && len(pr.Status.Conditions) == 1 {
			switch tektonv1beta1.PipelineRunReason(pr.Status.Conditions[0].Reason) {
			case tektonv1beta1.PipelineRunReasonSuccessful, tektonv1beta1.PipelineRunReasonCompleted:
				job.Status.State = cicdv1.IntegrationJobStateCompleted
			case tektonv1beta1.PipelineRunReasonFailed, tektonv1beta1.PipelineRunReasonCancelled, tektonv1beta1.PipelineRunReasonTimedOut:
				job.Status.State = cicdv1.IntegrationJobStateFailed
			}
		}

		// Reflect status of each task(job)
		// Be sure job.Status.Jobs[i] is set sequentially
		for i, j := range job.Spec.Jobs {
			var jStatus *tektonv1beta1.TaskRunStatus
			for _, runStatus := range pr.Status.TaskRuns {
				if runStatus.Status != nil && runStatus.PipelineTaskName == j.Name {
					jStatus = runStatus.Status
					break
				}
			}

			// Only update if taskRun's status exists
			if jStatus != nil && len(jStatus.Conditions) > 0 {
				startTime := jStatus.StartTime.DeepCopy()
				completionTime := jStatus.CompletionTime.DeepCopy()
				message := jStatus.Conditions[0].Message

				state := cicdv1.CommitStatusStatePending
				switch tektonv1beta1.TaskRunReason(jStatus.Conditions[0].Reason) {
				case tektonv1beta1.TaskRunReasonSuccessful:
					state = cicdv1.CommitStatusStateSuccess
				case tektonv1beta1.TaskRunReasonFailed, tektonv1beta1.TaskRunReasonCancelled, tektonv1beta1.TaskRunReasonTimedOut:
					state = cicdv1.CommitStatusStateFailure
				}

				// If something is changed, commit status should be posted
				stateChanged[i] = job.Status.Jobs[i].State != state ||
					job.Status.Jobs[i].Message != message ||
					!job.Status.Jobs[i].StartTime.Equal(startTime) ||
					!job.Status.Jobs[i].CompletionTime.Equal(completionTime)

				job.Status.Jobs[i].PodName = jStatus.PodName
				job.Status.Jobs[i].State = state
				job.Status.Jobs[i].Message = message
				job.Status.Jobs[i].StartTime = startTime
				job.Status.Jobs[i].CompletionTime = completionTime

				job.Status.Jobs[i].Containers = nil
				for _, s := range jStatus.Steps {
					stepStatus := s.DeepCopy()
					job.Status.Jobs[i].Containers = append(job.Status.Jobs[i].Containers, *stepStatus)
				}
			}
		}
	}

	// Set remote git's commit status for each job
	gitCli, err := utils.GetGitCli(cfg)
	if err != nil {
		return err
	}

	// If state is changed, update git commit status
	for i, j := range job.Status.Jobs {
		if stateChanged[i] {
			if err := gitCli.SetCommitStatus(job, cfg, j.Name, git.CommitStatusState(j.State), j.Message, job.GetReportServerAddress(j.Name), &client); err != nil {
				return err
			}
		}
	}

	return nil
}

func Name(j *cicdv1.IntegrationJob) string {
	return j.Name // TODO - should we add postfix or something? Job:PipelineRun=1:1...
}
