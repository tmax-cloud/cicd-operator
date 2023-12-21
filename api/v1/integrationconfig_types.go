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
	"time"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/pod"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tmax-cloud/cicd-operator/internal/configs"

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
	IntegrationConfigConditionWebhookRegistered = "webhook-registered"
	IntegrationConfigConditionReady             = "ready"
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

	// RequestBodyLogging is a boolean, which determines whether to enable logging reqeustBody coming toward webhook
	RequestBodyLogging bool `json:"requestBodyLogging,omitempty"`

	// IJManageSpec defines variables to manage created integration jobs
	IJManageSpec IntegrationJobManageSpec `json:"ijManageSpec,omitempty"`

	// ParamConfig specifies parameter
	ParamConfig *ParameterConfig `json:"paramConfig,omitempty"`

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

// ParameterConfig for parameters
type ParameterConfig struct {
	// ParamDefine is used to define parameter's spec
	ParamDefine []ParameterDefine `json:"paramDefine,omitempty"`
	// ParamValue can specify values of parameters
	ParamValue []ParameterValue `json:"paramValue,omitempty"`
}

// ParameterDefine defines a parameter's name, description & default values
type ParameterDefine struct {
	Name         string   `json:"name"`
	DefaultStr   string   `json:"defaultStr,omitempty"`
	DefaultArray []string `json:"defaultArray,omitempty"`
	Description  string   `json:"description,omitempty"`
}

// ConvertToTektonParamSpecs converts ParameterDefine array to tekton ParamSpec array
func ConvertToTektonParamSpecs(params []ParameterDefine) []tektonv1beta1.ParamSpec {
	var tektonParamSpecs []tektonv1beta1.ParamSpec
	for _, p := range params {
		if p.DefaultArray != nil {
			tektonParamSpecs = append(tektonParamSpecs, tektonv1beta1.ParamSpec{
				Name:        p.Name,
				Type:        tektonv1beta1.ParamTypeArray,
				Description: p.Description,
				Default:     tektonv1beta1.NewArrayOrString(p.DefaultArray[0], p.DefaultArray[1:]...),
			})
		} else {
			tektonParamSpecs = append(tektonParamSpecs, tektonv1beta1.ParamSpec{
				Name:        p.Name,
				Type:        tektonv1beta1.ParamTypeString,
				Description: p.Description,
				Default:     tektonv1beta1.NewArrayOrString(p.DefaultStr),
			})
		}

	}
	return tektonParamSpecs
}

// ParameterValue defines values of parameter
type ParameterValue struct {
	Name      string   `json:"name"`
	StringVal string   `json:"stringVal,omitempty"`
	ArrayVal  []string `json:"arrayVal,omitempty"`
}

// ConvertToTektonParams convert ParameterValue array to tekton Param array
func ConvertToTektonParams(params []ParameterValue) []tektonv1beta1.Param {
	var tektonParams []tektonv1beta1.Param
	for _, p := range params {
		v := tektonv1beta1.ArrayOrString{}

		if p.ArrayVal != nil {
			v.Type = tektonv1beta1.ParamTypeArray
			v.ArrayVal = append(v.ArrayVal, p.ArrayVal...)
		} else {
			v.Type = tektonv1beta1.ParamTypeString
			v.StringVal = p.StringVal
		}

		tektonParams = append(tektonParams, tektonv1beta1.Param{
			Name:  p.Name,
			Value: v,
		})
	}
	return tektonParams
}

// IntegrationJobManageSpec contains spec for ij managing
type IntegrationJobManageSpec struct {
	// Timeout for pending integration job gc
	Timeout *metav1.Duration `json:"timeout,omitempty"`
}

// IntegrationConfigJobs categorizes jobs into three types (pre-submit, post-submit and periodic jobs)
type IntegrationConfigJobs struct {
	// PreSubmit jobs are for pull-request events
	PreSubmit Jobs `json:"preSubmit,omitempty"`

	// PostSubmit jobs are for push events (including tag events)
	PostSubmit Jobs `json:"postSubmit,omitempty"`

	// Periodic are Periodicjobs can be run periodically
	Periodic Periodics `json:"periodic,omitempty"`
}

// IntegrationConfigStatus defines the observed state of IntegrationConfig
type IntegrationConfigStatus struct {
	// Conditions of IntegrationConfig
	Conditions []metav1.Condition `json:"conditions"`
	Secrets    string             `json:"secrets,omitempty"`
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

// GetDuration returns timeout duration. Default is TTL value
func (i *IntegrationConfig) GetDuration() *metav1.Duration {
	if i.Spec.IJManageSpec.Timeout != nil {
		return i.Spec.IJManageSpec.Timeout
	}
	return &metav1.Duration{
		Duration: time.Duration(configs.IntegrationJobTTL) * time.Hour,
	}
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
	BaseBranch          string               `json:"base_branch"`
	HeadBranch          string               `json:"head_branch"`
	AddTektonTaskParams []AddTektonTaskParam `json:"addTektonTaskParams,omitempty"`
}

// IntegrationConfigAPIReqRunPostBody is a body struct for IntegrationConfig's api request
// +kubebuilder:object:generate=false
type IntegrationConfigAPIReqRunPostBody struct {
	Branch              string               `json:"branch"`
	AddTektonTaskParams []AddTektonTaskParam `json:"addTektonTaskParams,omitempty"`
}

// AddTektonTaskParam represents additional Tekton task parameters
type AddTektonTaskParam struct {
	JobName    string          `json:"jobName"`
	TektonTask []TektonTaskDef `json:"tektonTask"`
}

// TektonTaskDef represents a definition for a Tekton task
type TektonTaskDef struct {
	Name      string `json:"name"`
	StringVal string `json:"stringVal"`
}

// IntegrationConfigAPIReqWebhookURL is a body struct for IntegrationConfig's api request
// +kubebuilder:object:generate=false
type IntegrationConfigAPIReqWebhookURL struct {
	URL    string `json:"url"`
	Secret string `json:"secret"`
}
