package v1

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"net/url"
)

// Default hosts for remote git servers
const (
	GithubDefaultAPIUrl = "https://api.github.com"
	GithubDefaultHost   = "https://github.com"

	GitlabDefaultAPIUrl = "https://gitlab.com"
	GitlabDefaultHost   = "https://gitlab.com"
)

// GitConfig is a git repository where the IntegrationConfig to be configured
type GitConfig struct {
	// Type for git remote server
	// +kubebuilder:validation:Enum=github;gitlab
	Type GitType `json:"type"`

	// Repository name of git repository (in <org>/<repo> form, e.g., tmax-cloud/cicd-operator)
	// +kubebuilder:validation:Pattern=.+/.+
	Repository string `json:"repository"`

	// APIUrl for api server (e.g., https://api.github.com for github type),
	// for the case where the git repository is self-hosted (should contain specific protocol otherwise webhook server returns error)
	APIUrl string `json:"apiUrl,omitempty"`

	// Token
	Token GitToken `json:"token"`
}

// GetGitHost gets git host
func (config *GitConfig) GetGitHost() (string, error) {
	gitURL := config.GetAPIUrl()
	if gitURL == GithubDefaultAPIUrl {
		gitURL = GithubDefaultHost
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
)
