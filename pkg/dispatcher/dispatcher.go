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

package dispatcher

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Dispatcher dispatches IntegrationJob when webhook is called
// A kind of 'plugin' for webhook handler
type Dispatcher struct {
	Client client.Client
}

// Name returns a name of dispatcher plugin
func (d *Dispatcher) Name() string {
	return "dispatcher"
}

// Handle handles pull-request and push events
func (d Dispatcher) Handle(webhook *git.Webhook, config *cicdv1.IntegrationConfig) error {
	var job *cicdv1.IntegrationJob
	pr := webhook.PullRequest
	push := webhook.Push
	if pr == nil && push == nil {
		return fmt.Errorf("pull request and push struct is nil")
	}

	if webhook.EventType == git.EventTypePullRequest && pr != nil {
		if pr.Action == git.PullRequestActionOpen || pr.Action == git.PullRequestActionSynchronize || pr.Action == git.PullRequestActionReOpen {
			prs := []git.PullRequest{*pr}
			job = GeneratePreSubmit(prs, &webhook.Repo, &webhook.Sender, config)
		}
	} else if webhook.EventType == git.EventTypePush && push != nil {
		job = GeneratePostSubmit(push, &webhook.Repo, &webhook.Sender, config)
	}

	if job == nil {
		return nil
	}

	if err := d.Client.Create(context.Background(), job); err != nil {
		return err
	}

	return nil
}

// GeneratePreSubmit generates IntegrationJob for pull request event
func GeneratePreSubmit(prs []git.PullRequest, repo *git.Repository, sender *git.User, config *cicdv1.IntegrationConfig) *cicdv1.IntegrationJob {
	jobs := FilterJobs(config.Spec.Jobs.PreSubmit, git.EventTypePullRequest, prs[0].Base.Ref)
	if len(jobs) < 1 {
		return nil
	}

	var ijName string
	if len(prs) > 1 {
		ijName = "batch"
	} else {
		ijName = prs[0].Head.Sha // only one PR Exists
	}

	jobID := utils.RandomString(20)
	return &cicdv1.IntegrationJob{
		ObjectMeta: generateMeta(config.Name, config.Namespace, ijName, jobID),
		Spec: cicdv1.IntegrationJobSpec{
			ConfigRef: cicdv1.IntegrationJobConfigRef{
				Name: config.Name,
				Type: cicdv1.JobTypePreSubmit,
			},
			ID:         jobID,
			Jobs:       jobs,
			Workspaces: config.Spec.Workspaces,
			Refs: cicdv1.IntegrationJobRefs{
				Repository: repo.Name,
				Link:       repo.URL,
				Sender: &cicdv1.IntegrationJobSender{
					Name:  sender.Name,
					Email: sender.Email,
				},
				Base: cicdv1.IntegrationJobRefsBase{
					Ref:  cicdv1.GitRef(prs[0].Base.Ref),
					Sha:  prs[0].Base.Sha,
					Link: repo.URL,
				},
				Pulls: generatePulls(prs),
			},
			PodTemplate: config.Spec.PodTemplate,
			Timeout:     config.GetDuration(),
			ParamConfig: config.Spec.ParamConfig,
		},
	}
}

// GeneratePostSubmit generates IntegrationJob for push event
func GeneratePostSubmit(push *git.Push, repo *git.Repository, sender *git.User, config *cicdv1.IntegrationConfig) *cicdv1.IntegrationJob {
	jobs := FilterJobs(config.Spec.Jobs.PostSubmit, git.EventTypePush, push.Ref)
	if len(jobs) < 1 {
		return nil
	}
	jobID := utils.RandomString(20)
	return &cicdv1.IntegrationJob{
		ObjectMeta: generateMeta(config.Name, config.Namespace, push.Sha, jobID),
		Spec: cicdv1.IntegrationJobSpec{
			ConfigRef: cicdv1.IntegrationJobConfigRef{
				Name: config.Name,
				Type: cicdv1.JobTypePostSubmit,
			},
			ID:         jobID,
			Jobs:       jobs,
			Workspaces: config.Spec.Workspaces,
			Refs: cicdv1.IntegrationJobRefs{
				Repository: repo.Name,
				Link:       repo.URL,
				Sender: &cicdv1.IntegrationJobSender{
					Name:  sender.Name,
					Email: sender.Email,
				},
				Base: cicdv1.IntegrationJobRefsBase{
					Ref:  cicdv1.GitRef(push.Ref),
					Link: repo.URL,
					Sha:  push.Sha,
				},
			},
			PodTemplate: config.Spec.PodTemplate,
			Timeout:     config.GetDuration(),
			ParamConfig: config.Spec.ParamConfig,
		},
	}
}

func generateMeta(cfgName, cfgNamespace, sha, jobID string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      fmt.Sprintf("%s-%s-%s", cfgName, sha[:5], jobID[:5]),
		Namespace: cfgNamespace,
		Labels: map[string]string{
			cicdv1.JobLabelConfig: cfgName,
			cicdv1.JobLabelID:     jobID,
		},
	}
}

func generatePulls(prs []git.PullRequest) []cicdv1.IntegrationJobRefsPull {
	pulls := []cicdv1.IntegrationJobRefsPull{}
	for _, pr := range prs {
		pull := generatePull(pr)
		pulls = append(pulls, pull)
	}
	return pulls
}

func generatePull(pr git.PullRequest) cicdv1.IntegrationJobRefsPull {
	return cicdv1.IntegrationJobRefsPull{
		ID:   pr.ID,
		Ref:  cicdv1.GitRef(pr.Head.Ref),
		Sha:  pr.Head.Sha,
		Link: pr.URL,
		Author: cicdv1.IntegrationJobRefsPullAuthor{
			Name: pr.Author.Name,
		},
	}
}

// FilterJobs filters job depending on the events, and ref
func FilterJobs(cand []cicdv1.Job, evType git.EventType, ref string) []cicdv1.Job {
	var filteredJobs []cicdv1.Job
	var incomingBranch string
	var incomingTag string

	switch evType {
	case git.EventTypePullRequest:
		incomingBranch = ref
	case git.EventTypePush:
		if strings.Contains(ref, "refs/tags/") {
			incomingTag = strings.Replace(ref, "refs/tags/", "", -1)
		} else {
			incomingBranch = strings.Replace(ref, "refs/heads/", "", -1)
		}
	}

	// Commit comment events
	if incomingBranch == "" && incomingTag == "" {
		filteredJobs = filterCommits(cand)
		return filteredJobs
	}
	//tag push events
	filteredJobs = filterTags(cand, incomingTag)
	filteredJobs = filterBranches(filteredJobs, incomingBranch)
	return filteredJobs
}

func filterCommits(jobs []cicdv1.Job) []cicdv1.Job {
	var filteredJobs []cicdv1.Job

	for _, job := range jobs {
		if job.When == nil {
			filteredJobs = append(filteredJobs, job)
			continue
		}
		branches := job.When.Branch
		for _, branch := range branches {
			if match := matchString("[commit]", branch); match {
				filteredJobs = append(filteredJobs, job)
				break
			}
		}
	}

	return filteredJobs
}

func filterTags(jobs []cicdv1.Job, incomingTag string) []cicdv1.Job {
	var filteredJobs []cicdv1.Job

	for _, job := range jobs {
		if job.When == nil {
			filteredJobs = append(filteredJobs, job)
			continue
		}
		tags := job.When.Tag
		skipTags := job.When.SkipTag

		// Always run if no tag/skipTag is specified
		if tags == nil && skipTags == nil {
			filteredJobs = append(filteredJobs, job)
		}

		if incomingTag == "" {
			continue
		}

		if tags != nil && skipTags == nil {
			for _, tag := range tags {
				if match := matchString(incomingTag, tag); match {
					filteredJobs = append(filteredJobs, job)
					break
				}
			}
		} else if skipTags != nil && tags == nil {
			isInValidJob := false
			for _, tag := range skipTags {
				if match := matchString(incomingTag, tag); match {
					isInValidJob = true
					break
				}
			}
			if !isInValidJob {
				filteredJobs = append(filteredJobs, job)
			}
		}
	}
	return filteredJobs
}

func filterBranches(jobs []cicdv1.Job, incomingBranch string) []cicdv1.Job {
	var filteredJobs []cicdv1.Job

	for _, job := range jobs {
		if job.When == nil {
			filteredJobs = append(filteredJobs, job)
			continue
		}
		branches := job.When.Branch
		skipBranches := job.When.SkipBranch

		// Always run if no branch/skipBranch is specified
		if branches == nil && skipBranches == nil {
			filteredJobs = append(filteredJobs, job)
		}

		if incomingBranch == "" {
			continue
		}

		if branches != nil && skipBranches == nil {
			for _, branch := range branches {
				if match := matchString(incomingBranch, branch); match {
					filteredJobs = append(filteredJobs, job)
					break
				}
			}
		} else if skipBranches != nil && branches == nil {
			isInValidJob := false
			for _, branch := range skipBranches {
				if match := matchString(incomingBranch, branch); match {
					isInValidJob = true
					break
				}
			}
			if !isInValidJob {
				filteredJobs = append(filteredJobs, job)
			}
		}
	}
	return filteredJobs
}

func matchString(incoming, target string) bool {
	re, err := regexp.Compile(target)
	if err != nil {
		return false
	}
	return re.MatchString(incoming)
}
