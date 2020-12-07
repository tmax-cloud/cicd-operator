package v1

import (
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

type JobType string

const (
	JobTypePreSubmit  = JobType("preSubmit")
	JobTypePostSubmit = JobType("postSubmit")
)

type Job struct {
	corev1.Container `json:",inline"`

	// TektonTask is for referring local Tasks or the Tasks registered in tekton catalog github repo.
	TektonTask *TektonTask `json:"tektonTask,omitempty"`

	// When is condition for running the job
	When *JobWhen `json:"when,omitempty"`

	// After configures which jobs should be executed before this job runs
	After []string `json:"after,omitempty"`

	// Approval
	// TODO

	// MailNotification
	// TODO
}

type TektonTask struct {
	// TaskRef refers to the existing Task in local cluster or to the tekton catalog github repo.
	TaskRef JobTaskRef `json:"taskRef"`

	// Params are input params for the task
	Params []tektonv1beta1.Param `json:"params,omitempty"`

	// Resources are input/output resources for the task
	Resources *tektonv1beta1.TaskRunResources `json:"resources,omitempty"`

	// Workspaces are workspaces for the task
	Workspaces []tektonv1beta1.WorkspaceBinding `json:"workspaces,omitempty"`
}

type JobTaskRef struct {
	// Name for local Tasks
	Name string `json:"name,omitempty"`

	// Catalog is a name of the task @ tekton catalog github repo. (e.g., s2i@0.2)
	// FYI: https://github.com/tektoncd/catalog
	Catalog string `json:"catalog,omitempty"`
}

type JobWhen struct {
	Branch     []string `json:"branch,omitempty"`
	SkipBranch []string `json:"skipBranch,omitempty"`

	Tag     []string `json:"tag,omitempty"`
	SkipTag []string `json:"skipTag,omitempty"`

	Ref     []string `json:"ref,omitempty"`
	SkipRef []string `json:"skipRef,omitempty"`
}
