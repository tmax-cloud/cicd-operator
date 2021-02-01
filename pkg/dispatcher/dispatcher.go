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

// Handle handles pull-request and push events
func (d Dispatcher) Handle(webhook *git.Webhook, config *cicdv1.IntegrationConfig) error {
	var job *cicdv1.IntegrationJob
	pr := webhook.PullRequest
	push := webhook.Push
	if pr == nil && push == nil {
		return fmt.Errorf("pull request and push struct is nil")
	}

	jobID := utils.RandomString(20)

	jobs, err := filterJobs(webhook, config)
	if err != nil {
		return err
	}
	if len(jobs) < 1 {
		return nil
	}

	if webhook.EventType == git.EventTypePullRequest {
		job = handlePullRequest(webhook, config, jobID, jobs)
	} else if webhook.EventType == git.EventTypePush {
		job = handlePush(webhook, config, jobID, jobs)
	}

	if job == nil {
		return nil
	}

	if err := d.Client.Create(context.Background(), job); err != nil {
		return err
	}

	return nil
}

func handlePullRequest(webhook *git.Webhook, config *cicdv1.IntegrationConfig, jobID string, jobs []cicdv1.Job) *cicdv1.IntegrationJob {
	pr := webhook.PullRequest
	var job *cicdv1.IntegrationJob
	if pr.Action == git.PullRequestActionOpen || pr.Action == git.PullRequestActionSynchronize || pr.Action == git.PullRequestActionReOpen {
		job = &cicdv1.IntegrationJob{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s-%s", config.Name, pr.Head.Sha[:5], jobID[:5]),
				Namespace: config.Namespace,
				Labels:    map[string]string{}, // TODO
			},
			Spec: cicdv1.IntegrationJobSpec{
				ConfigRef: cicdv1.IntegrationJobConfigRef{
					Name: config.Name,
					Type: cicdv1.JobTypePreSubmit,
				},
				ID:         jobID,
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
						ID:   pr.ID,
						Ref:  pr.Head.Ref,
						Sha:  pr.Head.Sha,
						Link: pr.URL,
						Author: cicdv1.IntegrationJobRefsPullAuthor{
							Name: pr.Sender.Name,
						},
					},
				},
				PodTemplate: config.Spec.PodTemplate,
			},
		}
	}
	return job
}

func handlePush(webhook *git.Webhook, config *cicdv1.IntegrationConfig, jobID string, jobs []cicdv1.Job) *cicdv1.IntegrationJob {
	push := webhook.Push
	job := &cicdv1.IntegrationJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s-%s", config.Name, push.Sha[:5], jobID[:5]),
			Namespace: config.Namespace,
			Labels:    map[string]string{}, // TODO
		},
		Spec: cicdv1.IntegrationJobSpec{
			ConfigRef: cicdv1.IntegrationJobConfigRef{
				Name: config.Name,
				Type: cicdv1.JobTypePostSubmit,
			},
			ID:         jobID,
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
			PodTemplate: config.Spec.PodTemplate,
		},
	}
	return job
}

func filterJobs(webhook *git.Webhook, config *cicdv1.IntegrationConfig) ([]cicdv1.Job, error) {
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

func filter(cand []cicdv1.Job, webhook *git.Webhook) ([]cicdv1.Job, error) {
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
	return filteredJobs, nil
}

func matchString(incoming, target string) bool {
	re, err := regexp.Compile(target)
	if err != nil {
		return false
	}
	return re.MatchString(incoming)
}
