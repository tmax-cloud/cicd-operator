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

package v1

import (
	"fmt"
	"time"

	"github.com/operator-framework/operator-lib/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Approval's kind string
const (
	ApprovalKind = "approvals"
)

// ApprovalResult is a result of the Approval
type ApprovalResult string

// Results
const (
	ApprovalResultAwaiting ApprovalResult = "Awaiting"
	ApprovalResultApproved ApprovalResult = "Approved"
	ApprovalResultRejected ApprovalResult = "Rejected"
	ApprovalResultError    ApprovalResult = "Error"
)

// Condition keys for Approval
const (
	ApprovalConditionSentRequestMail = status.ConditionType("SentRequestMail")
	ApprovalConditionSentResultMail  = status.ConditionType("SentResultMail")
)

// ApprovalSpec defines the desired state of Approval
type ApprovalSpec struct {
	// PodName represents the name of the pod to be approved to proceed
	// Deprecated: not used from HyperCloud5, only for the backward compatibility with HyperCloud4
	PodName string `json:"podName,omitempty"`

	// SkipSendMail describes whether or not to send mail for request/result for approvers
	SkipSendMail bool `json:"skipSendMail,omitempty"`

	// PipelineRun points the actual pipeline run object which created this Approval
	PipelineRun string `json:"pipelineRun,omitempty"`

	// IntegrationJob is a related IntegrationJob name (maybe a grand-parent of Approval)
	IntegrationJob string `json:"integrationJob,omitempty"`

	// JobName is a name of actual job in IntegrationJob
	JobName string `json:"jobName,omitempty"`

	// Message is a message from requester
	Message string `json:"message,omitempty"`

	// Sender is a requester (probably be pull-request author or pusher)
	Sender *ApprovalSender `json:"sender,omitempty"`

	// Link is a description link approvers may refer to
	Link string `json:"link,omitempty"`

	// Users are the list of the users who are requested to approve the Approval
	Users []string `json:"users"`
}

// ApprovalSender is a git user who triggered the approval
type ApprovalSender struct {
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
}

// ApprovalStatus defines the observed state of Approval
type ApprovalStatus struct {
	// Decision result of Approval
	Result ApprovalResult `json:"result"`

	// Approver is a user who actually approved
	Approver string `json:"approver,omitempty"`

	// Decision message
	Reason string `json:"reason,omitempty"`

	// Decision time of Approval
	DecisionTime *metav1.Time `json:"decisionTime,omitempty"`

	// Conditions of Approval
	Conditions status.Conditions `json:"conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Approval is the Schema for the approvals API
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.result",description="Current status of Approval"
// +kubebuilder:printcolumn:name="Created",type="date",JSONPath=".metadata.creationTimestamp",description="Created time"
// +kubebuilder:printcolumn:name="Decided",type="date",JSONPath=".status.decisionTime",description="Decided time"
type Approval struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApprovalSpec   `json:"spec"`
	Status ApprovalStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ApprovalList contains a list of Approval
type ApprovalList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Approval `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Approval{}, &ApprovalList{})
}

// GetDecisionTimeInZone converts the time into the specific time zone
func (a *ApprovalStatus) GetDecisionTimeInZone(zone string) (*time.Time, error) {
	if a.DecisionTime == nil {
		return nil, fmt.Errorf("decision time is nil")
	}
	location, err := time.LoadLocation(zone)
	if err != nil {
		return nil, err
	}
	localTime := a.DecisionTime.Time.In(location)
	return &localTime, nil
}

// Approval API kinds
const (
	ApprovalAPIApprove = "approve"
	ApprovalAPIReject  = "reject"
)

// ApprovalAPIReqBody is a body struct for Approval's api request
type ApprovalAPIReqBody struct {
	Reason string `json:"reason"`
}
