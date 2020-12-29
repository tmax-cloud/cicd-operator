package dispatcher

import (
	"context"
	"fmt"
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
	jobName := fmt.Sprintf("%s-%s-%s", config.Name, pr.Head.Sha[:5], jobId[:5])

	jobs, err := filterJobs(webhook, config)
	if err != nil {
		return err
	}
	if webhook.EventType == git.EventTypePullRequest {
		job = handlePullRequest(webhook, config, jobId, jobName, jobs)
	} else if webhook.EventType == git.EventTypePush {
		job = handlePush(webhook, config, jobId, jobName, jobs)
	}

	if err := d.Client.Create(context.Background(), job); err != nil {
		return err
	}

	return nil
}

func handlePullRequest(webhook git.Webhook, config *cicdv1.IntegrationConfig, jobId, jobName string, jobs []cicdv1.Job) *cicdv1.IntegrationJob {
	pr := webhook.PullRequest
	var job *cicdv1.IntegrationJob
	if webhook.PullRequest.Action == git.PullRequestActionOpen || webhook.PullRequest.Action == git.PullRequestActionSynchronize {
		job = &cicdv1.IntegrationJob{
			ObjectMeta: metav1.ObjectMeta{
				Name:      jobName,
				Namespace: config.Namespace,
				Labels:    map[string]string{}, // TODO
			},
			Spec: cicdv1.IntegrationJobSpec{
				ConfigRef: cicdv1.IntegrationJobConfigRef{
					Name: config.Name,
					Type: cicdv1.JobTypePreSubmit,
				},
				Id:   jobId,
				Jobs: jobs,
				Refs: cicdv1.IntegrationJobRefs{
					Repository: webhook.Repo.Name,
					Link:       webhook.Repo.URL,
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

func handlePush(webhook git.Webhook, config *cicdv1.IntegrationConfig, jobId, jobName string, jobs []cicdv1.Job) *cicdv1.IntegrationJob {
	push := webhook.Push
	job := &cicdv1.IntegrationJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: config.Namespace,
			Labels:    map[string]string{}, // TODO
		},
		Spec: cicdv1.IntegrationJobSpec{
			ConfigRef: cicdv1.IntegrationJobConfigRef{
				Name: config.Name,
				Type: cicdv1.JobTypePostSubmit,
			},
			Id:   jobId,
			Jobs: jobs,
			Refs: cicdv1.IntegrationJobRefs{
				Repository: webhook.Repo.Name,
				Link:       webhook.Repo.URL,
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

	// TODO - filter
	jobs = filter(cand, webhook, config)

	return jobs, nil
}

func filter(cand []cicdv1.Job, webhook git.Webhook, config *cicdv1.IntegrationConfig) []cicdv1.Job {
	var filteredJobs []cicdv1.Job
	var incomingBranch string
	var incomingTag string

	if webhook.EventType == git.EventTypePullRequest {
		incomingBranch = webhook.PullRequest.Base.Ref
	} else if webhook.EventType == git.EventTypePush {
		if strings.Contains(webhook.Push.Ref, "refs/tags/") {
			incomingTag = strings.Replace(webhook.Push.Ref, "refs/tags/", "")
		} else {
			incomingBranch = strings.Replace(webhook.Push.Ref, "refs/heads/", "")
		}
	}

	//tag push events
	if incomingTag != nil {
		filteredJobs = filterTags(cand, incomingTag)
	} else {
		filteredJobs = filterBranches(cand, incomingBranch)
	}
	return filteredJobs
}

func filterTags(jobs []cicdv1.Job, incomingTag string) []cicdv1.Job {
	var filteredJobs []cicdv1.Job

	for _, job := range cand {
		tags := job.When.Tag
		skipTags := job.When.SkipTag
		isValidJob := false

		if tags != nil && skipTags == nil {
			for _, tag := range tags {
				if incomingTag == tag {
					isValidJob = true
					break
				}
			}
			if isValidJob {
				append(filteredJobs, job)
			}
		} else if skipTags != nil && tags == nil {
			for _, job := range cand {
				isInValidJob := false
				for _, tag := range tags {
					if incomingTag == tag {
						isInValidJob = true
						break
					}
				}
				if !isInValidJob {
					append(filteredJobs, job)
				}
			}
		} else if tags == nil && skipTags == nil {
			append(filteredJobs, job)
		}
	}
	return filteredJobs
}

func filterBranches(jobs []cicdv1.Job, incomingBranch string) []cicdv1.Job {
	var filteredJobs []cicdv1.Job

	for _, job := range cand {
		branches := job.When.Branch
		skipBranches := job.When.SkipBranch
		isValidJob := false

		if branches != nil && skipBranches == nil {
			for _, branch := range branches {
				if incomingBranch == branch {
					isValidJob = true
					break
				}
			}
			if isValidJob {
				append(filteredJobs, job)
			}
		} else if skipBranches != nil && branches == nil {
			for _, job := range cand {
				isInValidJob := false
				for _, branch := range branches {
					if incomingBranch == branch {
						isInValidJob = true
						break
					}
				}
				if !isInValidJob {
					append(filteredJobs, job)
				}
			}
		} else if branches == nil && skipBranches == nil {
			append(filteredJobs, job)
		}
	}
	return filteredJobs
}
