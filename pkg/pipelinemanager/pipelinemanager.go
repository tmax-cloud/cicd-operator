package pipelinemanager

import (
	"context"
	"fmt"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"github.com/tmax-cloud/cicd-operator/pkg/events"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"strconv"
	"time"
)

// PipelineManager implements utility functions for generating, watching PipelineRuns

var log = logf.Log.WithName("pipeline-manager")

const (
	// DefaultWorkingDir is a default value for 'working directory' of a container (tekton's step)
	DefaultWorkingDir = "/tekton/home/integ-source"
)

// PipelineManager manages pipelines
type PipelineManager struct {
	Client client.Client
	Scheme *runtime.Scheme
}

// Generate generates (but not creates) a PipelineRun object
func (p *PipelineManager) Generate(job *cicdv1.IntegrationJob) (*tektonv1beta1.PipelineRun, error) {
	log.Info("Generating a pipeline run")

	// Workspace defs
	var workspaceDefs []tektonv1beta1.PipelineWorkspaceDeclaration
	for _, w := range job.Spec.Workspaces {
		workspaceDefs = append(workspaceDefs, tektonv1beta1.PipelineWorkspaceDeclaration{Name: w.Name})
	}

	// Resources
	var specResources []tektonv1beta1.PipelineDeclaredResource
	var runResources []tektonv1beta1.PipelineResourceBinding

	// Generate Tasks
	var tasks []tektonv1beta1.PipelineTask
	for _, j := range job.Spec.Jobs {
		taskSpec, resources, err := generateTask(job, &j)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *taskSpec)

		// Append resources
		for _, res := range resources {
			specRes, err := p.convertResourceToSpec(res.PipelineResourceBinding, job.Namespace)
			if err != nil {
				return nil, err
			}
			runResources = append(runResources, res.PipelineResourceBinding)
			specResources = append(specResources, *specRes)
		}
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
			Resources:          runResources,
			PipelineSpec: &tektonv1beta1.PipelineSpec{
				Resources:  specResources,
				Tasks:      tasks,
				Workspaces: workspaceDefs,
			},
			PodTemplate: job.Spec.PodTemplate,
			Workspaces:  job.Spec.Workspaces,
			Timeout:     &metav1.Duration{Duration: 120 * time.Hour},
		},
	}, nil
}

func (p *PipelineManager) convertResourceToSpec(binding tektonv1beta1.PipelineResourceBinding, ns string) (*tektonv1beta1.PipelineDeclaredResource, error) {
	// Get resource type
	var t tektonv1beta1.PipelineResourceType
	if binding.ResourceSpec != nil {
		t = binding.ResourceSpec.Type
	} else if binding.ResourceRef != nil {
		// Get PipelineResource
		res := &tektonv1alpha1.PipelineResource{}
		if err := p.Client.Get(context.Background(), types.NamespacedName{Name: binding.ResourceRef.Name, Namespace: ns}, res); err != nil {
			return nil, err
		}
		t = res.Spec.Type
	} else {
		return nil, fmt.Errorf("resource specs are all nil")
	}

	return &tektonv1beta1.PipelineDeclaredResource{
		Name: binding.Name,
		Type: t,
	}, nil
}

func generateTask(job *cicdv1.IntegrationJob, j *cicdv1.Job) (*tektonv1beta1.PipelineTask, []tektonv1beta1.TaskResourceBinding, error) {
	task := &tektonv1beta1.PipelineTask{Name: j.Name}

	var resources []tektonv1beta1.TaskResourceBinding

	// Handle TektonTask/Approval/Email
	if j.TektonTask != nil {
		res, err := generateTektonTaskRunTask(j, task)
		if err != nil {
			return nil, nil, err
		}
		resources = append(resources, res...)
	} else if j.Approval != nil {
		generateApprovalRunTask(job, j, task)
	} else if j.Email != nil {
		generateEmailRunTask(job, j, task)
	} else if j.Slack != nil {
		generateSlackRunTask(job, j, task)
	} else {
		// Steps
		steps, err := generateSteps(j)
		if err != nil {
			return nil, nil, err
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

	return task, resources, nil
}

func generateSteps(j *cicdv1.Job) ([]tektonv1beta1.Step, error) {
	var steps []tektonv1beta1.Step

	if !j.SkipCheckout {
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
		cicdv1.RunLabelJobID: j.Spec.ID,
		//cicdv1.RunLabelRepository: j.Spec.Refs.Repository,
		//cicdv1.RunLabelSender: j.Spec.Refs.Sender,
	}

	if j.Spec.Refs.Pull != nil {
		label[cicdv1.RunLabelPullRequest] = strconv.Itoa(j.Spec.Refs.Pull.ID)
		label[cicdv1.RunLabelPullRequestSha] = j.Spec.Refs.Pull.Sha
	}

	return label
}

// ReflectStatus reflects PipelineRun's status into IntegrationJob's status
// It also set commit status for remote git server
func (p *PipelineManager) ReflectStatus(pr *tektonv1beta1.PipelineRun, job *cicdv1.IntegrationJob, cfg *cicdv1.IntegrationConfig) error {
	oldState := job.Status.State
	oldMessage := job.Status.Message

	// If PR is nil but IntegrationJob's status is running, set as error
	// Also, schedule next pipelineRun
	if pr == nil && job.Status.State == cicdv1.IntegrationJobStateRunning {
		job.Status.State = cicdv1.IntegrationJobStateFailed
	}

	// Initialize status.jobs
	stateChanged := initState(job)

	// If PR exists, default state is running
	if pr != nil {
		job.Status.State = cicdv1.IntegrationJobStateRunning

		job.Status.StartTime = pr.CreationTimestamp.DeepCopy()
		job.Status.CompletionTime = pr.Status.CompletionTime.DeepCopy()

		prCond := pr.Status.GetCondition(apis.ConditionSucceeded)
		if prCond != nil {
			// Set message
			job.Status.Message = prCond.Message

			// Set state
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
			stateChanged[i] = p.reflectJobStatus(pr, &j, &job.Status.Jobs[i], job, cfg)
		}
	}

	// If it's start/completed but completion time is not set, set it as now
	if job.Status.State == cicdv1.IntegrationJobStateFailed || job.Status.State == cicdv1.IntegrationJobStateCompleted {
		t := &metav1.Time{Time: time.Now()}
		if job.Status.StartTime == nil {
			job.Status.StartTime = t
		}
		if job.Status.CompletionTime == nil {
			job.Status.CompletionTime = t
		}
	}

	// Set remote git's commit status for each job
	if err := p.updateGitCommitStatus(cfg, job, stateChanged); err != nil {
		return err
	}

	// Emit events
	if err := p.emitEvents(job, oldState, oldMessage); err != nil {
		return err
	}

	return nil
}

func initState(job *cicdv1.IntegrationJob) []bool {
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
	return stateChanged
}

func (p *PipelineManager) reflectJobStatus(pr *tektonv1beta1.PipelineRun, j *cicdv1.Job, jStatus *cicdv1.JobStatus, ij *cicdv1.IntegrationJob, cfg *cicdv1.IntegrationConfig) bool {
	changed := false

	runStatus := getJobRunStatus(pr, j)

	// Only update if taskRun's status exists
	if runStatus != nil {
		// If something is changed, commit status should be posted
		changed = !jStatus.Equals(runStatus)
		runStatus.DeepCopyInto(jStatus)

		// Handle post-run notifications for the completed jobs
		if runStatus.CompletionTime != nil {
			if err := p.handleNotification(runStatus, ij, cfg); err != nil {
				jStatus.Message = err.Error()
			}
		}
	}

	return changed
}

func getJobRunStatus(pr *tektonv1beta1.PipelineRun, j *cicdv1.Job) *cicdv1.JobStatus {
	jobStatus := &cicdv1.JobStatus{Name: j.Name, State: cicdv1.CommitStatusStatePending}
	// Find in TaskRun first
	for _, runStatus := range pr.Status.TaskRuns {
		if runStatus.Status != nil && runStatus.PipelineTaskName == j.Name {
			rStatus := runStatus.Status
			jobStatus.PodName = rStatus.PodName
			jobStatus.StartTime = rStatus.StartTime.DeepCopy()
			jobStatus.CompletionTime = rStatus.CompletionTime.DeepCopy()
			if len(rStatus.Conditions) > 0 {
				jobStatus.Message = rStatus.Conditions[0].Message
				switch tektonv1beta1.TaskRunReason(rStatus.Conditions[0].Reason) {
				case tektonv1beta1.TaskRunReasonSuccessful:
					jobStatus.State = cicdv1.CommitStatusStateSuccess
				case tektonv1beta1.TaskRunReasonFailed, tektonv1beta1.TaskRunReasonCancelled, tektonv1beta1.TaskRunReasonTimedOut:
					jobStatus.State = cicdv1.CommitStatusStateFailure
				}
			}
			jobStatus.Containers = nil
			for _, s := range rStatus.Steps {
				stepStatus := s.DeepCopy()
				jobStatus.Containers = append(jobStatus.Containers, *stepStatus)
			}
			break
		}
	}
	// Now find in Run
	if jobStatus.PodName == "" {
		for _, runStatus := range pr.Status.Runs {
			if runStatus.Status != nil && runStatus.PipelineTaskName == j.Name {
				rStatus := runStatus.Status
				jobStatus.StartTime = rStatus.StartTime.DeepCopy()
				jobStatus.CompletionTime = rStatus.CompletionTime.DeepCopy()
				if len(rStatus.Conditions) > 0 {
					jobStatus.Message = rStatus.Conditions[0].Message
					switch rStatus.Conditions[0].Status {
					case corev1.ConditionTrue:
						jobStatus.State = cicdv1.CommitStatusStateSuccess
					case corev1.ConditionFalse:
						jobStatus.State = cicdv1.CommitStatusStateFailure
					}
				}
				break
			}
		}
	}
	return jobStatus
}

func (p *PipelineManager) updateGitCommitStatus(cfg *cicdv1.IntegrationConfig, job *cicdv1.IntegrationJob, stateChanged []bool) error {
	gitCli, err := utils.GetGitCli(cfg, p.Client)
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
			log.Info(fmt.Sprintf("Setting commit status %s to %s", j.State, cfg.Spec.Git.Repository))
			if err := gitCli.SetCommitStatus(job, j.Name, git.CommitStatusState(j.State), msg, job.GetReportServerAddress(j.Name)); err != nil {
				log.Error(err, "")
			}
		}
	}

	return nil
}

func (p *PipelineManager) emitEvents(job *cicdv1.IntegrationJob, oldState cicdv1.IntegrationJobState, oldMessage string) error {
	if oldState == job.Status.State && oldMessage == job.Status.Message {
		return nil
	}
	return events.Emit(p.Client, job, corev1.EventTypeNormal, string(job.Status.State), job.Status.Message)
}

// Name is a PipelineRun's name for the IntegrationJob j
func Name(j *cicdv1.IntegrationJob) string {
	return j.Name
}
