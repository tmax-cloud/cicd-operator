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
	"net/url"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// Default hosts for remote git servers
const (
	GithubDefaultAPIUrl = "https://api.github.com"
	GithubDefaultHost   = "https://github.com"

	GitlabDefaultAPIUrl = "https://gitlab.com"
	GitlabDefaultHost   = "https://gitlab.com"

	GiteaDefaultAPIUrl = "https://gitea.com"
	GiteaDefaultHost   = "https://gitea.com/"
)

// GitConfig is a git repository where the IntegrationConfig to be configured
type GitConfig struct {
	// Type for git remote server
	// +kubebuilder:validation:Enum=github;gitlab;gitea
	Type GitType `json:"type"`

	// Repository name of git repository (in <org>/<repo> form, e.g., tmax-cloud/cicd-operator)
	// +kubebuilder:validation:Pattern=.+/.+
	Repository string `json:"repository"`

	// APIUrl for api server (e.g., https://api.github.com for github type),
	// for the case where the git repository is self-hosted (should contain specific protocol otherwise webhook server returns error)
	// Also, it should *NOT* contain repository path (e.g., tmax-cloud/cicd-operator)
	APIUrl string `json:"apiUrl,omitempty"`

	// Token is a token for accessing the remote git server. It can be empty, if you don't want to register a webhook
	// to the git server
	Token *GitToken `json:"token,omitempty"`
}

// GetGitHost gets git host
func (config *GitConfig) GetGitHost() (string, error) {
	gitURL := config.GetAPIUrl()
	if gitURL == GithubDefaultAPIUrl {
		gitURL = GithubDefaultHost
	} else if gitURL == GitlabDefaultAPIUrl {
		gitURL = GitlabDefaultHost
	} else if gitURL == GiteaDefaultAPIUrl {
		gitURL = GiteaDefaultHost
	}
	gitU, err := url.Parse(gitURL)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s://%s", gitU.Scheme, gitU.Host), nil
}

// GetAPIUrl returns APIUrl for api server
func (config *GitConfig) GetAPIUrl() string {
	if config.Type == GitTypeGitHub && config.APIUrl == "" {
		return GithubDefaultAPIUrl
	} else if config.Type == GitTypeGitLab && config.APIUrl == "" {
		return GitlabDefaultAPIUrl
	} else if config.Type == GitTypeGitea && config.APIUrl == "" {
		return GiteaDefaultAPIUrl
	}
	return config.APIUrl
}

// GitToken is a token for accessing the remote git server
type GitToken struct {
	// Value is un-encrypted plain string of git token, not recommended
	Value string `json:"value,omitempty"`

	// ValueFrom refers secret. Recommended
	ValueFrom *GitTokenFrom `json:"valueFrom,omitempty"`
}

// GitTokenFrom refers to the secret for the access token
type GitTokenFrom struct {
	SecretKeyRef corev1.SecretKeySelector `json:"secretKeyRef"`
}

// GitType is a type of remote git server
type GitType string

// Git Types
const (
	GitTypeGitHub = GitType("github")
	GitTypeGitLab = GitType("gitlab")
	GitTypeGitea  = GitType("gitea")
	GitTypeFake   = GitType("fake")
)

// GitRef is a git reference type
type GitRef string

func (g GitRef) String() string {
	return string(g)
}

// GetTag extracts tag from ref
func (g GitRef) GetTag() string {
	if !strings.HasPrefix(g.String(), "refs/") {
		return g.String()
	}
	if strings.HasPrefix(g.String(), "refs/tags/") {
		return strings.TrimPrefix(g.String(), "refs/tags/")
	}
	return ""
}

// GetBranch extracts branch from ref
func (g GitRef) GetBranch() string {
	if !strings.HasPrefix(g.String(), "refs/") {
		return g.String()
	}
	if strings.HasPrefix(g.String(), "refs/heads/") {
		return strings.TrimPrefix(g.String(), "refs/heads/")
	}
	return ""
}
