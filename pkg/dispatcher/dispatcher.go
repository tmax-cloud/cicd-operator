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

func (d Dispatcher) Handle(webhook git.Webhook, config *cicdv1.IntegrationConfig) error {
	var job *cicdv1.IntegrationJob
	pr := webhook.PullRequest
	push := webhook.Push
	if pr == nil && push == nil {
		return fmt.Errorf("pull request and push struct is nil")
	}

	jobId := utils.RandomString(20)

	jobs, err := filterJobs(webhook, config)
	if err != nil {
		return err
	}
	if len(jobs) < 1 {
		return nil
	}

	if webhook.EventType == git.EventTypePullRequest {
		job = handlePullRequest(webhook, config, jobId, jobs)
	} else if webhook.EventType == git.EventTypePush {
		job = handlePush(webhook, config, jobId, jobs)
	}

	if job == nil {
		return nil
	}

	if err := d.Client.Create(context.Background(), job); err != nil {
		return err
	}

	return nil
}

func handlePullRequest(webhook git.Webhook, config *cicdv1.IntegrationConfig, jobId string, jobs []cicdv1.Job) *cicdv1.IntegrationJob {
	pr := webhook.PullRequest
	var job *cicdv1.IntegrationJob
	if pr.Action == git.PullRequestActionOpen || pr.Action == git.PullRequestActionSynchronize || pr.Action == git.PullRequestActionReOpen {
		job = &cicdv1.IntegrationJob{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s-%s", config.Name, pr.Head.Sha[:5], jobId[:5]),
				Namespace: config.Namespace,
				Labels:    map[string]string{}, // TODO
			},
			Spec: cicdv1.IntegrationJobSpec{
				ConfigRef: cicdv1.IntegrationJobConfigRef{
					Name: config.Name,
					Type: cicdv1.JobTypePreSubmit,
				},
				Id:         jobId,
				Jobs:       jobs,
				Workspaces: config.Spec.Workspaces,
				Refs: cicdv1.IntegrationJobRefs{
					Repository: webhook.Repo.Name,
					Link:       webhook.Repo.URL,
					Sender: &cicdv1.IntegrationJobSender{
						Name:  pr.Sender.Name,
						Email: pr.Sender.Email,
					},
					Base: cicdv1.IntegrationJobRefsBase{
						Ref:  pr.Base.Ref,
						Link: webhook.Repo.URL,
					},
					Pull: &cicdv1.IntegrationJobRefsPull{
						Id:   pr.ID,
						Ref:  pr.Head.Ref,
						Sha:  pr.Head.Sha,
						Link: pr.URL,
						Author: cicdv1.IntegrationJobRefsPullAuthor{
							Name: pr.Sender.Name,
						},
					},
				},
			},
		}
	}
	return job
}

func handlePush(webhook git.Webhook, config *cicdv1.IntegrationConfig, jobId string, jobs []cicdv1.Job) *cicdv1.IntegrationJob {
	push := webhook.Push
	job := &cicdv1.IntegrationJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s-%s", config.Name, push.Sha[:5], jobId[:5]),
			Namespace: config.Namespace,
			Labels:    map[string]string{}, // TODO
		},
		Spec: cicdv1.IntegrationJobSpec{
			ConfigRef: cicdv1.IntegrationJobConfigRef{
				Name: config.Name,
				Type: cicdv1.JobTypePostSubmit,
			},
			Id:         jobId,
			Jobs:       jobs,
			Workspaces: config.Spec.Workspaces,
			Refs: cicdv1.IntegrationJobRefs{
				Repository: webhook.Repo.Name,
				Link:       webhook.Repo.URL,
				Sender: &cicdv1.IntegrationJobSender{
					Name:  push.Sender.Name,
					Email: push.Sender.Email,
				},
				Base: cicdv1.IntegrationJobRefsBase{
					Ref:  push.Ref,
					Link: webhook.Repo.URL,
					Sha:  push.Sha,
				},
			},
		},
	}
	return job
}

func filterJobs(webhook git.Webhook, config *cicdv1.IntegrationConfig) ([]cicdv1.Job, error) {
	var jobs []cicdv1.Job

	var cand []cicdv1.Job
	switch webhook.EventType {
	case git.EventTypePullRequest:
		cand = config.Spec.Jobs.PreSubmit
	case git.EventTypePush:
		cand = config.Spec.Jobs.PostSubmit
	default:
		return cand, nil
	}

	jobs, err := filter(cand, webhook)
	if err != nil {
		return nil, err
	}

	return jobs, nil
}

func filter(cand []cicdv1.Job, webhook git.Webhook) ([]cicdv1.Job, error) {
	var filteredJobs []cicdv1.Job
	var incomingBranch string
	var incomingTag string

	if webhook.EventType == git.EventTypePullRequest {
		incomingBranch = webhook.PullRequest.Base.Ref
	} else if webhook.EventType == git.EventTypePush {
		if strings.Contains(webhook.Push.Ref, "refs/tags/") {
			incomingTag = strings.Replace(webhook.Push.Ref, "refs/tags/", "", -1)
		} else {
			incomingBranch = strings.Replace(webhook.Push.Ref, "refs/heads/", "", -1)
		}
	}

	//tag push events
	var err error
	filteredJobs, err = filterTags(cand, incomingTag)
	if err != nil {
		return nil, err
	}
	filteredJobs, err = filterBranches(filteredJobs, incomingBranch)
	if err != nil {
		return nil, err
	}
	return filteredJobs, nil
}

func filterTags(jobs []cicdv1.Job, incomingTag string) ([]cicdv1.Job, error) {
	var filteredJobs []cicdv1.Job

	for _, job := range jobs {
		if job.When == nil {
			filteredJobs = append(filteredJobs, job)
			continue
		}
		tags := job.When.Tag
		skipTags := job.When.SkipTag

		if tags != nil && skipTags == nil {
			isValidJob := false
			if incomingTag != "" {
				for _, tag := range tags {
					re, err := regexp.Compile(tag)
					if err != nil {
						return nil, err
					}
					if re.MatchString(incomingTag) {
						isValidJob = true
						break
					}
				}
			}
			if isValidJob {
				filteredJobs = append(filteredJobs, job)
			}
		} else if skipTags != nil && tags == nil {
			isInValidJob := false
			if incomingTag != "" {
				for _, tag := range skipTags {
					re, err := regexp.Compile(tag)
					if err != nil {
						return nil, err
					}
					if re.MatchString(incomingTag) {
						isInValidJob = true
						break
					}
				}
			} else {
				isInValidJob = true
			}
			if !isInValidJob {
				filteredJobs = append(filteredJobs, job)
			}

		} else if tags == nil && skipTags == nil {
			filteredJobs = append(filteredJobs, job)
		}
	}
	return filteredJobs, nil
}

func filterBranches(jobs []cicdv1.Job, incomingBranch string) ([]cicdv1.Job, error) {
	var filteredJobs []cicdv1.Job

	for _, job := range jobs {
		if job.When == nil {
			filteredJobs = append(filteredJobs, job)
			continue
		}
		branches := job.When.Branch
		skipBranches := job.When.SkipBranch

		if branches != nil && skipBranches == nil {
			isValidJob := false
			if incomingBranch != "" {
				for _, branch := range branches {
					re, err := regexp.Compile(branch)
					if err != nil {
						return nil, err
					}
					if re.MatchString(incomingBranch) {
						isValidJob = true
						break
					}
				}
			}
			if isValidJob {
				filteredJobs = append(filteredJobs, job)
			}
		} else if skipBranches != nil && branches == nil {
			isInValidJob := false
			if incomingBranch != "" {
				for _, branch := range skipBranches {
					re, err := regexp.Compile(branch)
					if err != nil {
						return nil, err
					}
					if re.MatchString(incomingBranch) {
						isInValidJob = true
						break
					}
				}
			} else {
				isInValidJob = true
			}
			if !isInValidJob {
				filteredJobs = append(filteredJobs, job)
			}

		} else if branches == nil && skipBranches == nil {
			filteredJobs = append(filteredJobs, job)
		}
	}
	return filteredJobs, nil
}
