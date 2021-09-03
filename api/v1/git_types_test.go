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
)

func TestGitConfig_GetGitHost(t *testing.T) {
	tc := map[string]struct {
		cfg *GitConfig

		errorOccurs  bool
		errorMessage string
		expectedHost string
	}{
		"github": {
			cfg:          &GitConfig{Type: GitTypeGitHub},
			expectedHost: "https://github.com",
		},
		"gitlab": {
			cfg:          &GitConfig{Type: GitTypeGitLab},
			expectedHost: "https://gitlab.com",
		},
		"private": {
			cfg:          &GitConfig{Type: GitTypeGitLab, APIUrl: "https://gitlab.my.com/path"},
			expectedHost: "https://gitlab.my.com",
		},
		"error": {
			cfg:          &GitConfig{Type: GitTypeGitLab, APIUrl: "https://user:abc{DEf1=ghi@example.com:5432/db?sslmode=require"},
			errorOccurs:  true,
			errorMessage: "parse \"https://user:abc{DEf1=ghi@example.com:5432/db?sslmode=require\": net/url: invalid userinfo",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			host, err := c.cfg.GetGitHost()
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, c.expectedHost, host)
			}
		})
	}
}

func TestGitConfig_GetAPIUrl(t *testing.T) {
	tc := map[string]struct {
		cfg *GitConfig

		expectedURL string
	}{
		"github": {
			cfg:         &GitConfig{Type: GitTypeGitHub},
			expectedURL: "https://api.github.com",
		},
		"gitlab": {
			cfg:         &GitConfig{Type: GitTypeGitLab},
			expectedURL: "https://gitlab.com",
		},
		"private": {
			cfg:         &GitConfig{Type: GitTypeGitLab, APIUrl: "https://gitlab.my.com/path"},
			expectedURL: "https://gitlab.my.com/path",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, c.expectedURL, c.cfg.GetAPIUrl())
		})
	}
}

func TestGitRef_String(t *testing.T) {
	tc := map[string]gitTypeTestCase{
		"non-ref": {Input: "master", ExpectedOutput: "master"},
		"branch":  {Input: "refs/heads/master", ExpectedOutput: "refs/heads/master"},
		"tag":     {Input: "refs/tags/v0.1.1", ExpectedOutput: "refs/tags/v0.1.1"},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, c.ExpectedOutput, c.Input.String())
		})
	}
}

func TestGitRef_GetBranch(t *testing.T) {
	tc := map[string]gitTypeTestCase{
		"non-ref": {Input: "master", ExpectedOutput: "master"},
		"branch":  {Input: "refs/heads/master", ExpectedOutput: "master"},
		"tag":     {Input: "refs/tags/v0.1.1", ExpectedOutput: ""},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, c.ExpectedOutput, c.Input.GetBranch())
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
			require.Equal(t, c.ExpectedOutput, c.Input.GetTag())
		})
	}
}

type gitTypeTestCase struct {
	Input          GitRef
	ExpectedOutput string
}
