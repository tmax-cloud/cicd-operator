package v1

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"net/url"
)

const (
	GithubDefaultApiUrl = "https://api.github.com"
	GithubDefaultHost   = "https://github.com"

	GitlabDefaultApiUrl = "https://gitlab.com"
	GitlabDefaultHost   = "https://gitlab.com"
)

type GitConfig struct {
	// Type for git remote server
	// +kubebuilder:validation:Enum=github;gitlab
	Type GitType `json:"type"`

	// Repository name of git repository (in <org>/<repo> form, e.g., tmax-cloud/cicd-operator)
	// +kubebuilder:validation:Pattern=.+/.+
	Repository string `json:"repository"`

	// ApiUrl for api server (e.g., https://api.github.com for github type),
	// for the case where the git repository is self-hosted (should contain specific protocol otherwise webhook server returns error)
	ApiUrl string `json:"apiUrl,omitempty"`

	// Token
	Token GitToken `json:"token"`
}

// Get Git host
func (config *GitConfig) GetGitHost() (string, error) {
	gitUrl := config.GetApiUrl()
	if gitUrl == GithubDefaultApiUrl {
		gitUrl = GithubDefaultHost
	}
	gitU, err := url.Parse(gitUrl)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s://%s", gitU.Scheme, gitU.Host), nil
}

// Returns ApiUrl for api server
func (config *GitConfig) GetApiUrl() string {
	if config.Type == GitTypeGitHub && config.ApiUrl == "" {
		return GithubDefaultApiUrl
	} else if config.Type == GitTypeGitLab && config.ApiUrl == "" {
		return GitlabDefaultApiUrl
	}
	return config.ApiUrl
}

type GitToken struct {
	// Value is un-encrypted plain string of git token, not recommended
	Value string `json:"value,omitempty"`

	// ValueFrom refers secret. Recommended
	ValueFrom *GitTokenFrom `json:"valueFrom,omitempty"`
}

type GitTokenFrom struct {
	SecretKeyRef corev1.SecretKeySelector `json:"secretKeyRef"`
}

type GitType string

const (
	GitTypeGitHub = GitType("github")
	GitTypeGitLab = GitType("gitlab")
)
