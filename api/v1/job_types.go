package v1

import (
	"fmt"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tmax-cloud/cicd-operator/pkg/structs"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// JobType is a type of Job
type JobType string

const (
	// JobTypePreSubmit is a pre-submit type (pull-request or merge-request)
	JobTypePreSubmit = JobType("preSubmit")
	// JobTypePostSubmit is a post-submit type (push or tag-push)
	JobTypePostSubmit = JobType("postSubmit")
)

// CommitStatusState is a state of git commit status
type CommitStatusState string

// Job's commit state
const (
	CommitStatusStateSuccess = CommitStatusState("success")
	CommitStatusStateFailure = CommitStatusState("failure")
	CommitStatusStateError   = CommitStatusState("error")
	CommitStatusStatePending = CommitStatusState("pending")
)

// Job is a specification of the job to be executed for specific events
// Same level of task of tekton
type Job struct {
	corev1.Container `json:",inline"`

	// Script will override command of container
	Script string `json:"script,omitempty"`

	// SkipCheckout describes whether or not to checkout from git before
	SkipCheckout bool `json:"skipCheckout,omitempty"`

	// When is condition for running the job
	When *JobWhen `json:"when,omitempty"`

	// After configures which jobs should be executed before this job runs
	After []string `json:"after,omitempty"`

	// TektonTask is for referring local Tasks or the Tasks registered in tekton catalog github repo.
	TektonTask *TektonTask `json:"tektonTask,omitempty"`

	// Approval
	Approval *JobApproval `json:"approval,omitempty"`

	// Email
	Email *JobEmail `json:"email,omitempty"`
}

// TektonTask refers to an existing tekton task, rather than using job's script or command
type TektonTask struct {
	// TaskRef refers to the existing Task in local cluster or to the tekton catalog github repo.
	TaskRef JobTaskRef `json:"taskRef"`

	// Params are input params for the task
	Params []TektonTaskParam `json:"params,omitempty"`

	// Resources are input/output resources for the task
	Resources *tektonv1beta1.TaskRunResources `json:"resources,omitempty"`

	// Workspaces are workspaces for the task
	Workspaces []tektonv1beta1.WorkspacePipelineTaskBinding `json:"workspaces,omitempty"`
}

// TektonTaskParam replicates tekton's parameter
type TektonTaskParam struct {
	Name      string   `json:"name"`
	StringVal string   `json:"stringVal,omitempty"`
	ArrayVal  []string `json:"arrayVal,omitempty"`
}

// JobTaskRef refers to the tekton task, both local and in catalog
type JobTaskRef struct {
	// Local refers to local tasks/cluster tasks
	Local *tektonv1beta1.TaskRef `json:"local,omitempty"`

	// Catalog is a name of the task @ tekton catalog github repo. (e.g., s2i@0.2)
	// FYI: https://github.com/tektoncd/catalog
	Catalog string `json:"catalog,omitempty"`
}

// JobApproval describes who can approve it
type JobApproval struct {
	// Approvers is a list of approvers, in a form of <User name>=<Email> (Email is optional)
	// e.g., admin-tmax.co.kr
	// e.g., admin-tmax.co.kr=sunghyun_kim3@tmax.co.kr
	Approvers []string `json:"approvers,omitempty"`

	// ApproversConfigMap is a configMap Name containing
	// approvers list should exist in configMap's 'approvers' key, as comma(,) separated list
	// e.g., admin-tmax.co.kr=sunghyun_kim3@tmax.co.kr,test-tmax.co.kr=kyunghoon_min@tmax.co.kr
	ApproversConfigMap *corev1.LocalObjectReference `json:"approversConfigMap,omitempty"`

	// RequestMessage is a message to be sent to approvers by email
	RequestMessage string `json:"requestMessage"`
}

// JobEmail sends email to receivers
type JobEmail struct {
	// Receivers is a list of email receivers
	Receivers []string `json:"receivers,omitempty"`

	// Title of the email
	Title string `json:"title"`

	// Content of the email
	Content string `json:"content"`

	// IsHTML describes if it's html content
	IsHTML bool `json:"isHtml,omitempty"`
}

// JobWhen describes when the Job should be executed
// All fields should be regular expressions
type JobWhen struct {
	Branch     []string `json:"branch,omitempty"`
	SkipBranch []string `json:"skipBranch,omitempty"`

	Tag     []string `json:"tag,omitempty"`
	SkipTag []string `json:"skipTag,omitempty"`
}

// JobStatus is a current status for each job
type JobStatus struct {
	// Name is a job name
	Name string `json:"name"`

	// StartTime is a timestamp when the job is started
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime is a timestamp when the job is started
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// State is current state of this job
	// It is actually a conversion of tekton task run's Status.Conditions[0].Reason
	State CommitStatusState `json:"state"`

	// Message is current state description for this job
	// It is actually tekton task run's Status.Conditions[0].Message
	Message string `json:"message"`

	// PodName is a name of pod where the job is running
	PodName string `json:"podName,omitempty"`

	// Containers is status list for each step in the job
	Containers []tektonv1beta1.StepState `json:"containers,omitempty"`
}

// Equals checks if i is equal to j
func (j *JobStatus) Equals(i *JobStatus) bool {
	return j.State == i.State &&
		j.Message == i.Message &&
		j.StartTime.Equal(i.StartTime) &&
		j.CompletionTime.Equal(i.CompletionTime)
}

// Jobs is an array of Job
type Jobs []Job

// GetGraph creates job dependency graph
func (j *Jobs) GetGraph() (*structs.Graph, error) {
	graph := structs.NewGraph()
	for _, job := range *j {
		for _, after := range job.After {
			graph.AddEdge(after, job.Name)
		}
	}

	// Check cyclic
	if graph.IsCyclic() {
		return nil, fmt.Errorf("job graph is cyclic")
	}

	return graph, nil
}
