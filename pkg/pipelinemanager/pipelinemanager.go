package pipelinemanager

import (
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"strconv"
	"time"
)

// PipelineManager implements utility functions for generating, watching PipelineRuns

var log = logf.Log.WithName("pipeline-manager")

const (
	DefaultWorkingDir = "/tekton/home/integ-source"
)

func Generate(job *cicdv1.IntegrationJob) (*tektonv1beta1.PipelineRun, error) {
	log.Info("Generating a pipeline run")

	// Workspace defs
	var workspaceDefs []tektonv1beta1.PipelineWorkspaceDeclaration
	for _, w := range job.Spec.Workspaces {
		workspaceDefs = append(workspaceDefs, tektonv1beta1.PipelineWorkspaceDeclaration{Name: w.Name})
	}

	// Generate Tasks
	var tasks []tektonv1beta1.PipelineTask
	for _, j := range job.Spec.Jobs {
		taskSpec, err := generateTask(job, j)
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
				Tasks:      tasks,
				Workspaces: workspaceDefs,
			},
			Workspaces: job.Spec.Workspaces,
			Timeout:    &metav1.Duration{Duration: 120 * time.Hour},
		},
	}, nil
}

func generateTask(job *cicdv1.IntegrationJob, j cicdv1.Job) (*tektonv1beta1.PipelineTask, error) {
	task := &tektonv1beta1.PipelineTask{Name: j.Name}

	// Handle Approval
	if j.Approval != nil {
		if err := generateApprovalRunTask(job, j, task); err != nil {
			return nil, err
		}
	} else if j.Email != nil {
		if err := generateEmailRunTask(j, task); err != nil {
			return nil, err
		}
	} else {
		// Steps
		steps, err := generateSteps(j)
		if err != nil {
			return nil, err
		}

		task.TaskSpec = &tektonv1beta1.EmbeddedTask{}
		task.TaskSpec.Steps = steps

		// Workspaces
		var wsDefs []tektonv1beta1.WorkspaceDeclaration
		for _, w := range job.Spec.Workspaces {
			wsDefs = append(wsDefs, tektonv1beta1.WorkspaceDeclaration{Name: w.Name})
		}

		var wsBindings []tektonv1beta1.WorkspacePipelineTaskBinding
		for _, w := range job.Spec.Workspaces {
			wsBindings = append(wsBindings, tektonv1beta1.WorkspacePipelineTaskBinding{Name: w.Name, Workspace: w.Name})
		}

		task.TaskSpec.Workspaces = wsDefs
		task.Workspaces = wsBindings
	}

	// After
	task.RunAfter = append(task.RunAfter, j.After...)

	return task, nil
}

func generateCustomTaskRef(kind string) *tektonv1beta1.TaskRef {
	return &tektonv1beta1.TaskRef{
		APIVersion: cicdv1.CustomTaskApiVersion,
		Kind:       tektonv1beta1.TaskKind(kind),
	}
}

func generateSteps(j cicdv1.Job) ([]tektonv1beta1.Step, error) {
	var steps []tektonv1beta1.Step

	if j.GitCheckout {
		steps = append(steps, gitCheckout())
	}

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
			isTaskRun := false
			isRun := false
			var runPodName string
			var runStartTime *metav1.Time
			var runCompletionTime *metav1.Time
			var runReason string
			var runMessage string

			state := cicdv1.CommitStatusStatePending
			// Find in TaskRun first
			for _, runStatus := range pr.Status.TaskRuns {
				if runStatus.Status != nil && runStatus.PipelineTaskName == j.Name {
					rStatus := runStatus.Status
					runPodName = rStatus.PodName
					runStartTime = rStatus.StartTime.DeepCopy()
					runCompletionTime = rStatus.CompletionTime.DeepCopy()
					if len(rStatus.Conditions) > 0 {
						runMessage = rStatus.Conditions[0].Message
						runReason = rStatus.Conditions[0].Reason
					}
					job.Status.Jobs[i].Containers = nil
					for _, s := range rStatus.Steps {
						stepStatus := s.DeepCopy()
						job.Status.Jobs[i].Containers = append(job.Status.Jobs[i].Containers, *stepStatus)
					}
					switch tektonv1beta1.TaskRunReason(runReason) {
					case tektonv1beta1.TaskRunReasonSuccessful:
						state = cicdv1.CommitStatusStateSuccess
					case tektonv1beta1.TaskRunReasonFailed, tektonv1beta1.TaskRunReasonCancelled, tektonv1beta1.TaskRunReasonTimedOut:
						state = cicdv1.CommitStatusStateFailure
					}
					isTaskRun = true
					break
				}
			}
			// Now find in Run
			if !isTaskRun {
				for _, runStatus := range pr.Status.Runs {
					if runStatus.Status != nil && runStatus.PipelineTaskName == j.Name {
						rStatus := runStatus.Status
						runStartTime = rStatus.StartTime.DeepCopy()
						runCompletionTime = rStatus.CompletionTime.DeepCopy()
						if len(rStatus.Conditions) > 0 {
							runMessage = rStatus.Conditions[0].Message
							switch rStatus.Conditions[0].Status {
							case corev1.ConditionTrue:
								state = cicdv1.CommitStatusStateSuccess
							case corev1.ConditionFalse:
								state = cicdv1.CommitStatusStateFailure
							}
						}
						isRun = true
						break
					}
				}
			}

			// Only update if taskRun's status exists
			if isRun || isTaskRun {
				// If something is changed, commit status should be posted
				stateChanged[i] = job.Status.Jobs[i].State != state ||
					job.Status.Jobs[i].Message != runMessage ||
					!job.Status.Jobs[i].StartTime.Equal(runStartTime) ||
					!job.Status.Jobs[i].CompletionTime.Equal(runCompletionTime)

				job.Status.Jobs[i].PodName = runPodName
				job.Status.Jobs[i].State = state
				job.Status.Jobs[i].Message = runMessage
				job.Status.Jobs[i].StartTime = runStartTime
				job.Status.Jobs[i].CompletionTime = runCompletionTime
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
			msg := j.Message
			if len(msg) > 140 {
				msg = msg[:139]
			}
			if err := gitCli.SetCommitStatus(job, cfg, j.Name, git.CommitStatusState(j.State), msg, job.GetReportServerAddress(j.Name), client); err != nil {
				return err
			}
		}
	}

	return nil
}

func Name(j *cicdv1.IntegrationJob) string {
	return j.Name // TODO - should we add postfix or something? Job:PipelineRun=1:1...
}
