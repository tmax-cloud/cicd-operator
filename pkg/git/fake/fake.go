package fake

import (
	"fmt"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math/rand"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Store as global variables - only for testing! test data should be able to be set from the outside
var (
	Users    map[string]*git.User
	Repos    map[string]*Repo
	Branches map[string]*git.Branch
)

// Repo is a repository storage
type Repo struct {
	Webhooks     map[int]*git.WebhookEntry
	UserCanWrite map[string]bool

	PullRequests   map[int]*git.PullRequest
	CommitStatuses map[string][]git.CommitStatus
	Comments       map[int][]git.IssueComment
}

// Client is a gitlab client struct
type Client struct {
	IntegrationConfig *cicdv1.IntegrationConfig
	K8sClient         client.Client
}

// Init initiates the Client
func (c *Client) Init() error {
	return nil
}

// ParseWebhook parses a webhook body for github
func (c *Client) ParseWebhook(_ http.Header, _ []byte) (*git.Webhook, error) {
	return nil, nil
}

// ListWebhook lists registered webhooks
func (c *Client) ListWebhook() ([]git.WebhookEntry, error) {
	if Repos == nil {
		return nil, fmt.Errorf("repos not initialized")
	}
	repo, repoExist := Repos[c.IntegrationConfig.Spec.Git.Repository]
	if !repoExist {
		return nil, fmt.Errorf("404 no such repository")
	}

	var res []git.WebhookEntry
	for _, w := range repo.Webhooks {
		res = append(res, *w)
	}
	return res, nil
}

// RegisterWebhook registers our webhook server to the remote git server
func (c *Client) RegisterWebhook(url string) error {
	if Repos == nil {
		return fmt.Errorf("repos not initialized")
	}
	repo, repoExist := Repos[c.IntegrationConfig.Spec.Git.Repository]
	if !repoExist {
		return fmt.Errorf("404 no such repository")
	}

	if repo.Webhooks == nil {
		return fmt.Errorf("webhooks not initialized")
	}

	id := rand.Intn(100)
	repo.Webhooks[id] = &git.WebhookEntry{ID: id, URL: url}
	return nil
}

// DeleteWebhook deletes registered webhook
func (c *Client) DeleteWebhook(id int) error {
	if Repos == nil {
		return fmt.Errorf("repos not initialized")
	}
	repo, repoExist := Repos[c.IntegrationConfig.Spec.Git.Repository]
	if !repoExist {
		return fmt.Errorf("404 no such repository")
	}

	delete(repo.Webhooks, id)
	return nil
}

// ListCommitStatuses lists commit status of the specific commit
func (c *Client) ListCommitStatuses(ref string) ([]git.CommitStatus, error) {
	if Repos == nil {
		return nil, fmt.Errorf("repos not initialized")
	}
	repo, repoExist := Repos[c.IntegrationConfig.Spec.Git.Repository]
	if !repoExist {
		return nil, fmt.Errorf("404 no such repository")
	}

	if repo.CommitStatuses == nil {
		return nil, fmt.Errorf("commit statuses not initialized")
	}

	statuses, exist := repo.CommitStatuses[ref]
	if !exist {
		return nil, fmt.Errorf("404 no such ref")
	}
	return statuses, nil
}

// SetCommitStatus sets commit status for the specific commit
func (c *Client) SetCommitStatus(sha string, status git.CommitStatus) error {
	if Repos == nil {
		return fmt.Errorf("repos not initialized")
	}
	repo, repoExist := Repos[c.IntegrationConfig.Spec.Git.Repository]
	if !repoExist {
		return fmt.Errorf("404 no such repository")
	}

	if repo.CommitStatuses == nil {
		return fmt.Errorf("commit statuses not initialized")
	}

	repo.CommitStatuses[sha] = append(repo.CommitStatuses[sha], status)
	return nil
}

// GetUserInfo gets a user's information
func (c *Client) GetUserInfo(userName string) (*git.User, error) {
	if Users == nil {
		return nil, fmt.Errorf("users not initialized")
	}
	u, exist := Users[userName]
	if !exist {
		return nil, fmt.Errorf("404 no such user")
	}
	return u, nil
}

// CanUserWriteToRepo decides if the user has write permission on the repo
func (c *Client) CanUserWriteToRepo(user git.User) (bool, error) {
	if Repos == nil {
		return false, fmt.Errorf("repos not initialized")
	}
	repo, repoExist := Repos[c.IntegrationConfig.Spec.Git.Repository]
	if !repoExist {
		return false, fmt.Errorf("404 no such repository")
	}

	if repo.UserCanWrite == nil {
		return false, fmt.Errorf("userCanWrite not initialized")
	}

	privilege, exist := repo.UserCanWrite[user.Name]
	if !exist {
		return false, fmt.Errorf("404 no such user")
	}

	return privilege, nil
}

// RegisterComment registers comment to an issue
func (c *Client) RegisterComment(_ git.IssueType, issueNo int, body string) error {
	if Repos == nil {
		return fmt.Errorf("repos not initialized")
	}
	repo, repoExist := Repos[c.IntegrationConfig.Spec.Git.Repository]
	if !repoExist {
		return fmt.Errorf("404 no such repository")
	}

	if repo.Comments == nil {
		return fmt.Errorf("comments not initialized")
	}

	t := metav1.Now()
	repo.Comments[issueNo] = append(repo.Comments[issueNo], git.IssueComment{
		Comment: git.Comment{Body: body, CreatedAt: &t},
		Issue: git.Issue{
			PullRequest: &git.PullRequest{
				ID: issueNo,
			},
		},
		Sender: git.User{},
	})
	return nil
}

// ListPullRequests gets pull request list
func (c *Client) ListPullRequests(_ bool) ([]git.PullRequest, error) {
	if Repos == nil {
		return nil, fmt.Errorf("repos not initialized")
	}
	repo, repoExist := Repos[c.IntegrationConfig.Spec.Git.Repository]
	if !repoExist {
		return nil, fmt.Errorf("404 no such repository")
	}

	var prs []git.PullRequest
	for _, pr := range repo.PullRequests {
		prs = append(prs, *pr)
	}

	return prs, nil
}

// GetPullRequest gets PR given id
func (c *Client) GetPullRequest(id int) (*git.PullRequest, error) {
	if Repos == nil {
		return nil, fmt.Errorf("repos not initialized")
	}
	repo, repoExist := Repos[c.IntegrationConfig.Spec.Git.Repository]
	if !repoExist {
		return nil, fmt.Errorf("404 no such repository")
	}

	if repo.PullRequests == nil {
		return nil, fmt.Errorf("pull requests not initialized")
	}

	pr, exist := repo.PullRequests[id]
	if !exist {
		return nil, fmt.Errorf("404 no such pr")
	}

	return pr, nil
}

// MergePullRequest merges a pull request
func (c *Client) MergePullRequest(id int, sha string, method git.MergeMethod, message string) error {
	return nil
}

// SetLabel sets label to the issue id
func (c *Client) SetLabel(_ git.IssueType, id int, label string) error {
	if Repos == nil {
		return fmt.Errorf("repos not initialized")
	}
	repo, repoExist := Repos[c.IntegrationConfig.Spec.Git.Repository]
	if !repoExist {
		return fmt.Errorf("404 no such repository")
	}

	if repo.PullRequests == nil {
		return fmt.Errorf("pull requests not initialized")
	}

	pr, exist := repo.PullRequests[id]
	if !exist {
		return fmt.Errorf("404 no such PR")
	}

	pr.Labels = append(pr.Labels, git.IssueLabel{Name: label})

	return nil
}

// DeleteLabel deletes label from the issue id
func (c *Client) DeleteLabel(_ git.IssueType, id int, label string) error {
	return DeleteLabel(c.IntegrationConfig.Spec.Git.Repository, id, label)
}

// GetBranch returns branch info
func (c *Client) GetBranch(branch string) (*git.Branch, error) {
	if Branches == nil {
		return nil, fmt.Errorf("branches not initialized")
	}
	b, exist := Branches[branch]
	if !exist {
		return nil, fmt.Errorf("404 no such branch")
	}
	return b, nil
}

// DeleteLabel deletes label from a pull request
func DeleteLabel(repoName string, id int, label string) error {
	if Repos == nil {
		return fmt.Errorf("repos not initialized")
	}
	repo, repoExist := Repos[repoName]
	if !repoExist {
		return fmt.Errorf("404 no such repository")
	}

	if repo.PullRequests == nil {
		return fmt.Errorf("pull requests not initialized")
	}

	pr, exist := repo.PullRequests[id]
	if !exist {
		return fmt.Errorf("404 no such PR")
	}

	idx := -1
	for i, l := range pr.Labels {
		if l.Name == label {
			idx = i
			break
		}
	}
	if idx >= 0 {
		if idx == len(pr.Labels)-1 {
			pr.Labels = pr.Labels[:idx]
		} else {
			pr.Labels = append(pr.Labels[:idx], pr.Labels[idx+1:]...)
		}
	}

	return nil
}
