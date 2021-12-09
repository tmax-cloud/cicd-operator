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
	"testing"

	"github.com/stretchr/testify/require"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConvertToTektonParamSpecs(t *testing.T) {
	tc := map[string]struct {
		params            []ParameterDefine
		expectedParamSpec []tektonv1beta1.ParamSpec
	}{
		"params": {
			params: []ParameterDefine{
				{
					Name:         "array-param-spec",
					DefaultArray: []string{"array-string1", "array-string2"},
					Description:  "ParamSpec with default array",
				},
				{
					Name:        "string-param-spec",
					DefaultStr:  "string",
					Description: "ParamSpec with default string",
				},
			},
			expectedParamSpec: []tektonv1beta1.ParamSpec{
				{
					Name:        "array-param-spec",
					Type:        "array",
					Description: "ParamSpec with default array",
					Default:     tektonv1beta1.NewArrayOrString("array-string1", "array-string2"),
				},
				{
					Name:        "string-param-spec",
					Type:        "string",
					Description: "ParamSpec with default string",
					Default:     tektonv1beta1.NewArrayOrString("string"),
				},
			},
		},
	}
	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			paramSpec := ConvertToTektonParamSpecs(c.params)
			require.Equal(t, c.expectedParamSpec, paramSpec)
		})
	}
}

func TestConvertToTektonParams(t *testing.T) {
	tc := map[string]struct {
		params         []ParameterValue
		expectedParams []tektonv1beta1.Param
	}{
		"params": {
			params: []ParameterValue{
				{
					Name:     "array-param",
					ArrayVal: []string{"array-string1", "array-string2"},
				},
				{
					Name:      "string-param",
					StringVal: "string",
				},
			},
			expectedParams: []tektonv1beta1.Param{
				{
					Name:  "array-param",
					Value: *tektonv1beta1.NewArrayOrString("array-string1", "array-string2"),
				},
				{
					Name:  "string-param",
					Value: *tektonv1beta1.NewArrayOrString("string"),
				},
			},
		},
	}
	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			params := ConvertToTektonParams(c.params)
			require.Equal(t, c.expectedParams, params)
		})
	}
}

func TestIntegrationJobStatus_SetDefaults(t *testing.T) {
	tc := map[string]struct {
		currentState  IntegrationJobState
		expectedState IntegrationJobState
	}{
		"default": {
			expectedState: IntegrationJobStatePending,
		},
		"alreadySet": {
			currentState:  IntegrationJobStateCompleted,
			expectedState: IntegrationJobStateCompleted,
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			ijStatus := &IntegrationJobStatus{State: c.currentState}
			ijStatus.SetDefaults()
			require.Equal(t, c.expectedState, ijStatus.State)
		})
	}
}

func TestIntegrationJob_GetReportServerAddress(t *testing.T) {
	configs.CurrentExternalHostName = "test.host.com"
	ij := &IntegrationJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ij",
			Namespace: "test-ns",
		},
	}
	require.Equal(t, "http://test.host.com/report/test-ns/test-ij/test-job", ij.GetReportServerAddress("test-job"))
}
