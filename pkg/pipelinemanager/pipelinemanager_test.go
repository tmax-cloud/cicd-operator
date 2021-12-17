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

package pipelinemanager

import (
	"testing"

	"github.com/bmizerany/assert"
	"github.com/stretchr/testify/require"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
)

func TestAppendBaseShaToDescription(t *testing.T) {
	desc := "test description"
	sha := git.FakeSha

	appended := appendBaseShaToDescription(desc, sha)
	assert.Equal(t, desc, appended[:len(desc)], "Description")
	assert.Equal(t, statusDescriptionBaseSHAKey+git.FakeSha, appended[len(appended)-len(statusDescriptionBaseSHAKey+git.FakeSha):], "BaseSHA")

	desc = "description which is very longlonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglong"
	msgLen := statusDescriptionMaxLength - len(statusDescriptionBaseSHAKey) - len(git.FakeSha) - len(statusDescriptionEllipse)
	appended = appendBaseShaToDescription(desc, sha)
	assert.Equal(t, desc[:msgLen], appended[:len(desc[:msgLen])], "Description")
	assert.Equal(t, statusDescriptionBaseSHAKey+git.FakeSha, appended[len(appended)-len(statusDescriptionBaseSHAKey+git.FakeSha):], "BaseSHA")

	sha = ""
	appended = appendBaseShaToDescription(desc, sha)
	assert.Equal(t, desc[:statusDescriptionMaxLength], appended, "Description")
}

func TestParseBaseFromDescription(t *testing.T) {
	fullDesc := "Job is running... BaseSHA:2641c89aac959fb804ec6f2a4a22e129f4ac4900"
	sha := ParseBaseFromDescription(fullDesc)
	assert.Equal(t, "2641c89aac959fb804ec6f2a4a22e129f4ac4900", sha)

	fullDesc = "Job is running... BaseSHA:zzzzzzzzzzzzzzzzz"
	sha = ParseBaseFromDescription(fullDesc)
	assert.Equal(t, "", sha)
}

func TestGetParams(t *testing.T) {
	tc := map[string]struct {
		job *cicdv1.IntegrationJob

		expectedParamSpec []tektonv1beta1.ParamSpec
		expectedParam     []tektonv1beta1.Param
	}{
		"nilConfig": {
			job: &cicdv1.IntegrationJob{
				Spec: cicdv1.IntegrationJobSpec{},
			},
			expectedParamSpec: nil,
			expectedParam:     nil,
		},
		"existConfig": {
			job: &cicdv1.IntegrationJob{
				Spec: cicdv1.IntegrationJobSpec{
					ParamConfig: &cicdv1.ParameterConfig{
						ParamDefine: []cicdv1.ParameterDefine{
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
						ParamValue: []cicdv1.ParameterValue{
							{
								Name:     "array-param",
								ArrayVal: []string{"array-string1", "array-string2"},
							},
							{
								Name:      "string-param",
								StringVal: "string",
							},
						},
					},
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
			expectedParam: []tektonv1beta1.Param{
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
			paramSpec, param := getParams(c.job)

			require.Equal(t, c.expectedParamSpec, paramSpec)
			require.Equal(t, c.expectedParam, param)
		})
	}
}
