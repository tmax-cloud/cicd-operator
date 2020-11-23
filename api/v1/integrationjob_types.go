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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IntegrationJobSpec defines the desired state of IntegrationJob
type IntegrationJobSpec struct {
	// ConfigRef refers to the corresponding IntegrationConfig
	ConfigRef IntegrationJobConfigRef `json:"configRef"`

	// Id is a unique random string for the IntegrationJob
	Id string `json:"id"`

	// Jobs are the tasks to be executed
	Jobs []Job `json:"jobs"`

	// Refs
	Refs IntegrationJobRefs `json:"refs"`
}

type IntegrationJobConfigRef struct {
	Name string  `json:"name"`
	Type JobType `json:"type"`
}

type IntegrationJobRefs struct {
	// Repository name of git repository (in <org>/<repo> form, e.g., tmax-cloud/cicd-operator)
	// +kubebuilder:validation:Pattern=.+/.+
	Repository string `json:"repository"`

	// Link is a full url of the repository
	Link string `json:"link"`

	// Base is a base pointer for base commit for the pull request
	// If Pull is nil (i.e., is push event), Base works as Head
	Base IntegrationJobRefsBase `json:"base"`

	// Pull represents pull request head commit
	Pull *IntegrationJobRefsPull `json:"pull,omitempty"`
}

type IntegrationJobRefsBase struct {
	Ref  string `json:"ref"`
	Sha  string `json:"sha"`
	Link string `json:"link"`
}

type IntegrationJobRefsPull struct {
	Id     int                          `json:"id"`
	Sha    string                       `json:"sha"`
	Link   string                       `json:"link"`
	Author IntegrationJobRefsPullAuthor `json:"author"`
}

type IntegrationJobRefsPullAuthor struct {
	Name string `json:"name"`
	Link string `json:"link"`
}

// IntegrationJobStatus defines the observed state of IntegrationJob
type IntegrationJobStatus struct {
	// State is a current state of the IntegrationJob
	State IntegrationJobState `json:"state"`

	// CompletionTime is a time when the job is completed
	CompletionTime *metav1.Time `json:"completionTime"`

	// TaskStatus
	// TODO
}

type IntegrationJobState string

const (
	IntegrationJobStatePending   = IntegrationJobState("pending")
	IntegrationJobStateScheduled = IntegrationJobState("scheduled")
	IntegrationJobStateRunning   = IntegrationJobState("running")
	IntegrationJobStateCompleted = IntegrationJobState("completed")
	IntegrationJobStateFailed    = IntegrationJobState("failed")
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// IntegrationJob is the Schema for the integrationjobs API
type IntegrationJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IntegrationJobSpec   `json:"spec,omitempty"`
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

func (s *IntegrationJobStatus) SetDefaults() error {
	// TODO
	s.State = IntegrationJobStatePending
	return nil
}
