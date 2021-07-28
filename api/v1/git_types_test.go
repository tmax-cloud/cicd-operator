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

	"github.com/bmizerany/assert"
)

func TestGitRef_GetBranch(t *testing.T) {
	tc := map[string]gitTypeTestCase{
		"non-ref": {Input: "master", ExpectedOutput: "master"},
		"branch":  {Input: "refs/heads/master", ExpectedOutput: "master"},
		"tag":     {Input: "refs/tags/v0.1.1", ExpectedOutput: ""},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, c.ExpectedOutput, c.Input.GetBranch())
		})
	}
}

func TestGitRef_GetTag(t *testing.T) {
	tc := map[string]gitTypeTestCase{
		"non-ref": {Input: "master", ExpectedOutput: "master"},
		"branch":  {Input: "refs/heads/master", ExpectedOutput: ""},
		"tag":     {Input: "refs/tags/v0.1.1", ExpectedOutput: "v0.1.1"},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, c.ExpectedOutput, c.Input.GetTag())
		})
	}
}

type gitTypeTestCase struct {
	Input          GitRef
	ExpectedOutput string
}
