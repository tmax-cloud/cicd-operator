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

package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/tmax-cloud/cicd-operator/pkg/git/fake"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	"github.com/tmax-cloud/cicd-operator/pkg/git/github"
	"github.com/tmax-cloud/cicd-operator/pkg/git/gitlab"
)

// GetGitCli generates git client, depending on the git type in the cfg
func GetGitCli(cfg *cicdv1.IntegrationConfig, cli client.Client) (git.Client, error) {
	var c git.Client
	switch cfg.Spec.Git.Type {
	case cicdv1.GitTypeGitHub:
		c = &github.Client{IntegrationConfig: cfg, K8sClient: cli}
	case cicdv1.GitTypeGitLab:
		c = &gitlab.Client{IntegrationConfig: cfg, K8sClient: cli}
	case cicdv1.GitTypeFake:
		c = &fake.Client{IntegrationConfig: cfg, K8sClient: cli}
	default:
		return nil, fmt.Errorf("git type %s is not supported", cfg.Spec.Git.Type)
	}
	if err := c.Init(); err != nil {
		return nil, err
	}
	return c, nil
}

// ParseApproversList parses user/email from line-separated and comma-separated approvers list
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
				return nil, fmt.Errorf("comma-separated approver %s is not in form of <user-name>[=<email>](optional)", approver)
			}
		}
	}

	return approvers, nil
}

// ParseEmailFromUsers parses email from approvers list
func ParseEmailFromUsers(users []cicdv1.ApprovalUser) []string {
	var emails []string

	emailRe := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	for _, u := range users {
		if emailRe.MatchString(u.Email) {
			emails = append(emails, u.Email)
		}
	}

	return emails
}
