package utils

import (
	"fmt"
	"os"

	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	"github.com/tmax-cloud/cicd-operator/pkg/git/github"
	"github.com/tmax-cloud/cicd-operator/pkg/git/gitlab"
)

func FileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func GetGitCli(cfg *cicdv1.IntegrationConfig) (git.Client, error) {
	switch cfg.Spec.Git.Type {
	case cicdv1.GitTypeGitHub:
		return &github.Client{}, nil
	case cicdv1.GitTypeGitLab:
		return &gitlab.Client{}, nil
	default:
		return nil, fmt.Errorf("git type %s is not supported", cfg.Spec.Git.Type)
	}
}
