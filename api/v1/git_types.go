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
	// for the case where the git repository is self-hosted
	ApiUrl string `json:"apiUrl,omitempty"`

	// Token
	Token GitToken `json:"token"`
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
