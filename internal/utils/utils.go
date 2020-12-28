package utils

import (
	"fmt"
	"os"
	"regexp"
	"strings"

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

func ParseApproversList(str string) ([]string, error) {
	var approvers []string

	// Regexp for verifying if it's in form
	re := regexp.MustCompile("[^=]+(=.+)?")

	lineSep := strings.Split(strings.TrimSpace(str), "\n")
	for _, line := range lineSep {
		commaSep := strings.Split(strings.TrimSpace(line), ",")
		for _, approver := range commaSep {
			trimmed := strings.TrimSpace(approver)
			if re.MatchString(trimmed) {
				approvers = append(approvers, trimmed)
			} else {
				return nil, fmt.Errorf("comma-seperated approver %s is not in form of <user-name>[=<email>](optional)", approver)
			}
		}
	}

	return approvers, nil
}

func ParseEmailFromUsers(users []string) []string {
	var emails []string

	emailRe := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	for _, u := range users {
		subs := strings.Split(u, "=")
		if len(subs) < 2 {
			continue
		}
		trimmed := strings.TrimSpace(subs[1])
		if emailRe.MatchString(trimmed) {
			emails = append(emails, trimmed)
		}
	}

	return emails
}
