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
	"github.com/bmizerany/assert"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	"testing"
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
