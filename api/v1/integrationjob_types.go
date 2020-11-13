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
}

// IntegrationJobStatus defines the observed state of IntegrationJob
type IntegrationJobStatus struct {
}

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
