/*


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

package v1

import (
	"fmt"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/pod"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IntegrationJobState is a state of the IntegrationJob
type IntegrationJobState string

// IntegrationJob's states
const (
	IntegrationJobStatePending   = IntegrationJobState("Pending")
	IntegrationJobStateRunning   = IntegrationJobState("Running")
	IntegrationJobStateCompleted = IntegrationJobState("Completed")
	IntegrationJobStateFailed    = IntegrationJobState("Failed")
)

// IntegrationJobSpec defines the desired state of IntegrationJob
type IntegrationJobSpec struct {
	// ConfigRef refers to the corresponding IntegrationConfig
	ConfigRef IntegrationJobConfigRef `json:"configRef"`

	// ID is a unique random string for the IntegrationJob
	ID string `json:"id"`

	// Workspaces list
	Workspaces []tektonv1beta1.WorkspaceBinding `json:"workspaces,omitempty"`

	// Jobs are the tasks to be executed
	Jobs Jobs `json:"jobs"`

	// Refs
	Refs IntegrationJobRefs `json:"refs"`

	// PodTemplate for the TaskRun pods. Same as tekton's pod template
	PodTemplate *pod.Template `json:"podTemplate,omitempty"`
}

// IntegrationJobConfigRef refers to the IntegrationConfig
type IntegrationJobConfigRef struct {
	Name string  `json:"name"`
	Type JobType `json:"type"`
}

// IntegrationJobRefs describes the git event
type IntegrationJobRefs struct {
	// Repository name of git repository (in <org>/<repo> form, e.g., tmax-cloud/cicd-operator)
	// +kubebuilder:validation:Pattern=.+/.+
	Repository string `json:"repository"`

	// Link is a full url of the repository
	Link string `json:"link"`

	// Sender is a git user who triggered the webhook
	Sender *IntegrationJobSender `json:"sender"`

	// Base is a base pointer for base commit for the pull request
	// If Pull is nil (i.e., is push event), Base works as Head
	Base IntegrationJobRefsBase `json:"base"`

	// Pull represents pull request head commit
	Pull *IntegrationJobRefsPull `json:"pull,omitempty"`
}

// IntegrationJobSender is a git user who triggered the IntegrationJob
type IntegrationJobSender struct {
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
}

// IntegrationJobRefsBase refers to the base commit
type IntegrationJobRefsBase struct {
	Ref  GitRef `json:"ref"`
	Link string `json:"link"`
	Sha  string `json:"sha"`
}

// IntegrationJobRefsPull refers to the pull request
type IntegrationJobRefsPull struct {
	ID     int                          `json:"id"`
	Ref    GitRef                       `json:"ref"`
	Sha    string                       `json:"sha"`
	Link   string                       `json:"link"`
	Author IntegrationJobRefsPullAuthor `json:"author"`
}

// IntegrationJobRefsPullAuthor is an author of the pull request
type IntegrationJobRefsPullAuthor struct {
	Name string `json:"name"`
}

// IntegrationJobStatus defines the observed state of IntegrationJob
type IntegrationJobStatus struct {
	// State is a current state of the IntegrationJob
	State IntegrationJobState `json:"state"`

	// Message is a message for the IntegrationJob (normally an error string)
	Message string `json:"message,omitempty"`

	// StartTime is actual time the task started
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime is a time when the job is completed
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// Jobs are status list for each Job in the IntegrationJob
	Jobs []JobStatus `json:"jobs,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// IntegrationJob is the Schema for the integrationjobs API
// +kubebuilder:resource:shortName="ij"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state",description="Current status of IntegrationJob"
// +kubebuilder:printcolumn:name="StartTime",type="date",JSONPath=".status.startTime",description="Start time"
// +kubebuilder:printcolumn:name="CompletionTime",type="date",JSONPath=".status.completionTime",description="Completion time"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Creation time"
type IntegrationJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IntegrationJobSpec   `json:"spec"`
	Status IntegrationJobStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IntegrationJobList contains a list of IntegrationJob
type IntegrationJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IntegrationJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IntegrationJob{}, &IntegrationJobList{})
}

// SetDefaults sets default values for the status
func (s *IntegrationJobStatus) SetDefaults() error {
	if s.State == "" {
		s.State = IntegrationJobStatePending
	}
	return nil
}

// GetReportServerAddress returns Server address for reports (IntegrationJob details)
func (i *IntegrationJob) GetReportServerAddress(jobName string) string {
	return fmt.Sprintf("http://%s/report/%s/%s/%s", configs.ExternalHostName, i.Namespace, i.Name, jobName)
}
