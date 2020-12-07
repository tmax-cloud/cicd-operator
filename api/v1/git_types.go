package v1

import corev1 "k8s.io/api/core/v1"

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

// Returns ApiUrl for api server
func (config *GitConfig) GetApiUrl() string {
	if config.Type == GitTypeGitHub && config.ApiUrl == "" {
		return "https://api.github.com"
	} else if config.Type == GitTypeGitLab && config.ApiUrl == "" {
		return "https://gitlab.com"
	}
	return config.ApiUrl
}

// Returns Server address which webhook events will be received
func (config *GitConfig) GetServerAddress() string {
	return ""
}

type GitToken struct {
	// Value is un-encrypted plain string of git token, not recommended
	Value string `json:"value"`

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
