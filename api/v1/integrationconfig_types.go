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
	"context"
	"crypto/tls"
	"fmt"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/pod"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tmax-cloud/cicd-operator/internal/configs"

	"github.com/operator-framework/operator-lib/status"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// IntegrationConfigKind is kind string
const (
	IntegrationConfigKind = "integrationconfigs"
)

// Condition keys for IntegrationConfig
const (
	IntegrationConfigConditionWebhookRegistered = status.ConditionType("webhook-registered")
	IntegrationConfigConditionReady             = status.ConditionType("ready")
)

// IntegrationConfigConditionReasonNoGitToken is a Reason key
const (
	IntegrationConfigConditionReasonNoGitToken = "noGitToken"
)

// IntegrationConfigSpec defines the desired state of IntegrationConfig
type IntegrationConfigSpec struct {
	// Git config for target repository
	Git GitConfig `json:"git"`

	// Secrets are the list of secret names which are included in service account
	Secrets []corev1.LocalObjectReference `json:"secrets,omitempty"`

	// Workspaces list
	Workspaces []tektonv1beta1.WorkspaceBinding `json:"workspaces,omitempty"`

	// Jobs specify the tasks to be executed
	Jobs IntegrationConfigJobs `json:"jobs"`

	// MergeConfig specifies how to automate the PR merge
	MergeConfig *MergeConfig `json:"mergeConfig,omitempty"`

	// PodTemplate for the TaskRun pods. Same as tekton's pod template. Refer to https://github.com/tektoncd/pipeline/blob/master/docs/podtemplates.md
	PodTemplate *pod.Template `json:"podTemplate,omitempty"`

	// ReqeustBodyLogging is a boolean, which determines whether to enable logging reqeustBody coming toward webhook
	ReqeustBodyLogging bool `json:"requestBodyLogging,omitempty"`

	// TLSConfig set tls configurations
	TLSConfig *TLSConfig `json:"tlsConfig,omitempty"`

	// When is condition for running the job in global scope
	When *JobWhen `json:"when,omitempty"`

	// GolbalNotification sends notification when success/fail
	GolbalNotification *Notification `json:"globalNotification,omitempty"`
}

// TLSConfig is parameters for tls connection
type TLSConfig struct {
	// InsecureSkipVerify is flag for accepting any certificate presented by the server and any host name in that certificate.
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`
}

// IntegrationConfigJobs categorizes jobs into two types (pre-submit and post-submit)
type IntegrationConfigJobs struct {
	// PreSubmit jobs are for pull-request events
	PreSubmit Jobs `json:"preSubmit,omitempty"`

	// PostSubmit jobs are for push events (including tag events)
	PostSubmit Jobs `json:"postSubmit,omitempty"`
}

// IntegrationConfigStatus defines the observed state of IntegrationConfig
type IntegrationConfigStatus struct {
	// Conditions of IntegrationConfig
	Conditions status.Conditions `json:"conditions"`
	Secrets    string            `json:"secrets,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// IntegrationConfig is the Schema for the integrationconfigs API
// +kubebuilder:resource:shortName="ic"
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="ready")].status`
// +kubebuilder:printcolumn:name="WebhookRegistered",type=string,JSONPath=`.status.conditions[?(@.type=="webhook-registered")].status`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Creation time"
type IntegrationConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IntegrationConfigSpec   `json:"spec"`
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

// GetToken fetches git access token from IntegrationConfig
func (i *IntegrationConfig) GetToken(c client.Client) (string, error) {
	tokenStruct := i.Spec.Git.Token

	// Empty token
	if tokenStruct == nil {
		return "", nil
	}

	// Get from value
	if tokenStruct.ValueFrom == nil {
		if tokenStruct.Value != "" {
			return tokenStruct.Value, nil
		}
		return "", fmt.Errorf("token is empty")
	}

	// Get from secret
	secretName := tokenStruct.ValueFrom.SecretKeyRef.Name
	secretKey := tokenStruct.ValueFrom.SecretKeyRef.Key
	secret := &corev1.Secret{}
	if err := c.Get(context.Background(), types.NamespacedName{Name: secretName, Namespace: i.Namespace}, secret); err != nil {
		return "", err
	}
	token, ok := secret.Data[secretKey]
	if !ok {
		return "", fmt.Errorf("token secret/key %s/%s not valid", secretName, secretKey)
	}
	return string(token), nil
}

// GetServiceAccountName returns the name of the related ServiceAccount
func GetServiceAccountName(configName string) string {
	return fmt.Sprintf("%s-sa", configName)
}

// GetSecretName returns the name of related secret
func GetSecretName(configName string) string {
	return configName
}

// GetWebhookServerAddress returns Server address which webhook events will be received
func (i *IntegrationConfig) GetWebhookServerAddress() string {
	return fmt.Sprintf("http://%s/webhook/%s/%s", configs.CurrentExternalHostName, i.Namespace, i.Name)
}

// GetTLSConfig returns tls config from integration configs' tlsConfig
func (i *IntegrationConfig) GetTLSConfig() *tls.Config {
	if i.Spec.TLSConfig != nil {
		return &tls.Config{
			InsecureSkipVerify: i.Spec.TLSConfig.InsecureSkipVerify,
		}
	}
	return nil
}

// IntegrationConfig's API kinds
const (
	IntegrationConfigAPIRunPre     = "runpre"
	IntegrationConfigAPIRunPost    = "runpost"
	IntegrationConfigAPIWebhookURL = "webhookurl"
)

// IntegrationConfigAPIReqRunPreBody is a body struct for IntegrationConfig's api request
// +kubebuilder:object:generate=false
type IntegrationConfigAPIReqRunPreBody struct {
	BaseBranch string `json:"base_branch"`
	HeadBranch string `json:"head_branch"`
}

// IntegrationConfigAPIReqRunPostBody is a body struct for IntegrationConfig's api request
// +kubebuilder:object:generate=false
type IntegrationConfigAPIReqRunPostBody struct {
	Branch string `json:"branch"`
}

// IntegrationConfigAPIReqWebhookURL is a body struct for IntegrationConfig's api request
// +kubebuilder:object:generate=false
type IntegrationConfigAPIReqWebhookURL struct {
	URL    string `json:"url"`
	Secret string `json:"secret"`
}
