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
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

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
)

// PipelineManager implements utility functions for generating, watching PipelineRuns

var log = logf.Log.WithName("pipeline-manager")

// DefaultWorkingDir is a default value for 'working directory' of a container (tekton's step)
const DefaultWorkingDir = "/tekton/home/integ-source"

// Job messages
const (
	JobMessagePending    = "Job is running"
	JobMessageSuccessful = "Job succeeded"
	JobMessageFailure    = "Job failed"
)

const (
	statusDescriptionBaseSHAKey = "BaseSHA:"
	statusDescriptionMaxLength  = 140
	statusDescriptionEllipse    = "... "
)

// PipelineManager manages pipelines
type PipelineManager interface {
	Generate(job *cicdv1.IntegrationJob) (*tektonv1beta1.PipelineRun, error)
	ReflectStatus(pr *tektonv1beta1.PipelineRun, job *cicdv1.IntegrationJob, cfg *cicdv1.IntegrationConfig) error
}

// pipelineManager is an actual implementation
type pipelineManager struct {
	Client client.Client
	Scheme *runtime.Scheme
}

// NewPipelineManager initiates a new PipelineManager
func NewPipelineManager(c client.Client, s *runtime.Scheme) PipelineManager {
	return &pipelineManager{Client: c, Scheme: s}
}

// Generate generates (but not creates) a PipelineRun object
func (p *pipelineManager) Generate(job *cicdv1.IntegrationJob) (*tektonv1beta1.PipelineRun, error) {
	log.Info("Generating a pipeline run")
	// token for private repo
	var token string
	// Workspace defs
	var workspaceDefs []tektonv1beta1.PipelineWorkspaceDeclaration
	for _, w := range job.Spec.Workspaces {
		workspaceDefs = append(workspaceDefs, tektonv1beta1.PipelineWorkspaceDeclaration{Name: w.Name})
		if w.Secret.SecretName == "private-access-token" {
			token = p.getToken(w.Secret.SecretName, job.Namespace)
		}
	}

	// Params
	paramDefine, paramValue := getParams(job)

	// Resources
	var specResources []tektonv1beta1.PipelineDeclaredResource
	var runResources []tektonv1beta1.PipelineResourceBinding

	// Generate Tasks
	var tasks []tektonv1beta1.PipelineTask
	for _, j := range job.Spec.Jobs {
		taskSpec, resources, err := generateTask(job, &j, token)
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
				Params:     paramDefine,
			},
			PodTemplate: job.Spec.PodTemplate,
			Workspaces:  job.Spec.Workspaces,
			Timeout: &metav1.Duration{
				Duration: job.Spec.Timeout.Duration,
			},
			Params: paramValue,
		},
	}, nil
}

func getParams(job *cicdv1.IntegrationJob) ([]tektonv1beta1.ParamSpec, []tektonv1beta1.Param) {
	var paramSpec []tektonv1beta1.ParamSpec
	var param []tektonv1beta1.Param
	if job.Spec.ParamConfig != nil {
		paramSpec = cicdv1.ConvertToTektonParamSpecs(job.Spec.ParamConfig.ParamDefine)
		param = cicdv1.ConvertToTektonParams(job.Spec.ParamConfig.ParamValue)
	}

	return paramSpec, param
}

func (p *pipelineManager) getToken(name string, ns string) string {
	secret := &corev1.Secret{}
	if err := p.Client.Get(context.Background(), types.NamespacedName{Name: name, Namespace: ns}, secret); err != nil {
		return "fail to get token"
	}
	token, ok := secret.Data["access-token"]
	if !ok {
		return ""
	}
	return string(token)
}

func (p *pipelineManager) convertResourceToSpec(binding tektonv1beta1.PipelineResourceBinding, ns string) (*tektonv1beta1.PipelineDeclaredResource, error) {
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

func generateTask(job *cicdv1.IntegrationJob, j *cicdv1.Job, token string) (*tektonv1beta1.PipelineTask, []tektonv1beta1.TaskResourceBinding, error) {
	task := &tektonv1beta1.PipelineTask{Name: j.Name}
	var resources []tektonv1beta1.TaskResourceBinding

	// Handle TektonTask/Approval/Email
	if j.TektonTask != nil {
		res, err := generateTektonTaskRunTask(j, task, token)
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

	// TektonWhen
	task.WhenExpressions = append(task.WhenExpressions, j.TektonWhen...)

	// Results
	if task.TaskSpec != nil {
		task.TaskSpec.Results = append(task.TaskSpec.Results, j.Results...)
	}

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

	if j.Spec.Refs.Pulls != nil {
		label[cicdv1.RunLabelPullRequest] = strconv.Itoa(j.Spec.Refs.Pulls[0].ID)
		label[cicdv1.RunLabelPullRequestSha] = j.Spec.Refs.Pulls[0].Sha
	}

	return label
}

// ReflectStatus reflects PipelineRun's status into IntegrationJob's status
// It also set commit status for remote git server
func (p *pipelineManager) ReflectStatus(pr *tektonv1beta1.PipelineRun, job *cicdv1.IntegrationJob, cfg *cicdv1.IntegrationConfig) error {
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

	if job.Spec.ConfigRef.Type != cicdv1.JobTypePeriodic {
		// Set remote git's commit status for each job
		if err := p.updateGitCommitStatus(cfg, job, stateChanged); err != nil {
			return err
		}
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

func (p *pipelineManager) reflectJobStatus(pr *tektonv1beta1.PipelineRun, j *cicdv1.Job, jStatus *cicdv1.JobStatus, ij *cicdv1.IntegrationJob, cfg *cicdv1.IntegrationConfig) bool {
	changed := false

	runStatus := getJobRunStatus(pr.Status, j)

	// Only update if taskRun's status exists
	if runStatus != nil {
		// If something is changed, commit status should be posted (except for message - message is decided by the state)
		changed = jStatus.State != runStatus.State || !jStatus.StartTime.Equal(runStatus.StartTime) || !jStatus.CompletionTime.Equal(runStatus.CompletionTime)
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

func getJobRunStatus(prStatus tektonv1beta1.PipelineRunStatus, j *cicdv1.Job) *cicdv1.JobStatus {
	jobStatus := &cicdv1.JobStatus{Name: j.Name, State: cicdv1.CommitStatusStatePending}
	// Find in TaskRun first
	reflectFromTaskRuns(prStatus, j, jobStatus)
	// Now find in Run
	reflectFromRuns(prStatus, j, jobStatus)

	return jobStatus
}

func reflectFromTaskRuns(prStatus tektonv1beta1.PipelineRunStatus, j *cicdv1.Job, jobStatus *cicdv1.JobStatus) {
	for _, runStatus := range prStatus.TaskRuns {
		if runStatus.Status != nil && runStatus.PipelineTaskName == j.Name {
			rStatus := runStatus.Status
			jobStatus.PodName = rStatus.PodName
			jobStatus.StartTime = rStatus.StartTime.DeepCopy()
			jobStatus.CompletionTime = rStatus.CompletionTime.DeepCopy()
			if len(rStatus.Conditions) > 0 && rStatus.CompletionTime != nil {
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

func reflectFromRuns(prStatus tektonv1beta1.PipelineRunStatus, j *cicdv1.Job, jobStatus *cicdv1.JobStatus) {
	if jobStatus.PodName == "" {
		for _, runStatus := range prStatus.Runs {
			if runStatus.Status != nil && runStatus.PipelineTaskName == j.Name {
				rStatus := runStatus.Status
				jobStatus.StartTime = rStatus.StartTime.DeepCopy()
				jobStatus.CompletionTime = rStatus.CompletionTime.DeepCopy()
				if len(rStatus.Conditions) > 0 && rStatus.CompletionTime != nil {
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
}

func (p *pipelineManager) updateGitCommitStatus(cfg *cicdv1.IntegrationConfig, job *cicdv1.IntegrationJob, stateChanged []bool) error {
	// Skip if token is nil
	if cfg.Spec.Git.Token == nil {
		return nil
	}
	gitCli, err := utils.GetGitCli(cfg, p.Client)
	if err != nil {
		return err
	}

	// Skip if Multipie PRs exist
	if len(job.Spec.Refs.Pulls) > 1 {
		return nil
	}

	// If state is changed, update git commit status
	for i, j := range job.Status.Jobs {
		if stateChanged[i] {
			// Set simple message
			msg := JobMessagePending
			switch j.State {
			case cicdv1.CommitStatusStateSuccess:
				msg = JobMessageSuccessful
			case cicdv1.CommitStatusStateFailure:
				msg = JobMessageFailure
			}
			if job.Spec.Refs.Pulls != nil {
				msg = appendBaseShaToDescription(msg, job.Spec.Refs.Base.Sha)
			}

			// Get SHA of the commit
			var sha string
			if job.Spec.Refs.Pulls == nil {
				sha = job.Spec.Refs.Base.Sha
			} else {
				sha = job.Spec.Refs.Pulls[0].Sha
			}
			log.Info(fmt.Sprintf("Setting commit status %s:%s to %s's %s", j.Name, j.State, cfg.Spec.Git.Repository, sha))
			if err := gitCli.SetCommitStatus(sha, git.CommitStatus{Context: j.Name, State: git.CommitStatusState(j.State), Description: msg, TargetURL: job.GetReportServerAddress(j.Name)}); err != nil {
				log.Error(err, "")
			}
		}
	}

	return nil
}

// appendBaseShaToDescription appends Base SHA to the commit statuses' description.
// Merger can use this base SHA to check if the tests of the pull request is done against the most recent commit of the
// target branch before merging it.
func appendBaseShaToDescription(desc, sha string) string {
	if sha == "" {
		// Truncate message
		if len(desc) > statusDescriptionMaxLength {
			return desc[:statusDescriptionMaxLength]
		}
		return desc
	}

	base := statusDescriptionBaseSHAKey + sha

	// Truncate if msg is too long
	ellipseBase := statusDescriptionEllipse + base
	maxDescLen := statusDescriptionMaxLength - len(ellipseBase)
	if len(desc) > maxDescLen {
		return desc[:maxDescLen] + ellipseBase
	}

	// Add space if it's too short
	numSpace := statusDescriptionMaxLength - len(base) - len(desc)
	space := ""
	for i := 0; i < numSpace; i++ {
		space += " "
	}

	return desc + space + base
}

// ParseBaseFromDescription parses a SHA from the commit status's description
func ParseBaseFromDescription(desc string) string {
	sub := regexp.MustCompile(statusDescriptionBaseSHAKey + "([0-9a-f]{5,40})").FindStringSubmatch(desc)
	if len(sub) != 2 {
		return ""
	}
	return sub[1]
}

// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch;delete

func (p *pipelineManager) emitEvents(job *cicdv1.IntegrationJob, oldState cicdv1.IntegrationJobState, oldMessage string) error {
	if oldState == job.Status.State && oldMessage == job.Status.Message {
		return nil
	}
	return events.Emit(p.Client, job, corev1.EventTypeNormal, string(job.Status.State), job.Status.Message)
}

// Name is a PipelineRun's name for the IntegrationJob j
func Name(j *cicdv1.IntegrationJob) string {
	return j.Name
}
