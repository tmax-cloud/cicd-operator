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
	"github.com/operator-framework/operator-lib/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IntegrationConfigSpec defines the desired state of IntegrationConfig
type IntegrationConfigSpec struct {
}

// IntegrationConfigStatus defines the observed state of IntegrationConfig
type IntegrationConfigStatus struct {
	Conditions status.Conditions `json:"conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// IntegrationConfig is the Schema for the integrationconfigs API
type IntegrationConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IntegrationConfigSpec   `json:"spec,omitempty"`
	Status IntegrationConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IntegrationConfigList contains a list of IntegrationConfig
type IntegrationConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IntegrationConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IntegrationConfig{}, &IntegrationConfigList{})
}
