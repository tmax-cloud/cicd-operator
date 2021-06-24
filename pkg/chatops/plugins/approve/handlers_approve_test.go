package approve

import (
	"github.com/bmizerany/assert"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/chatops"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	gitfake "github.com/tmax-cloud/cicd-operator/pkg/git/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
	"time"
)

const (
	testRepo = "test/repo"
	testPRID = 11

	testNamespace  = "default"
	testConfigName = "test-ic"

	testUserID    = 32
	testUserName  = "test-user"
	testUserEmail = "test@test.com"
)

func TestChatOps_handleApprove(t *testing.T) {
	s := runtime.NewScheme()
	utilruntime.Must(cicdv1.AddToScheme(s))

	ic := buildTestConfigForApprove()
	wh := buildTestWebhookForApprove()

	fakeCli := fake.NewFakeClientWithScheme(s, ic)

	handler := &Handler{Client: fakeCli}

	gitfake.Users = map[string]*git.User{
		testUserName: {ID: testUserID, Name: testUserName, Email: testUserEmail},
	}
	gitfake.Repos = map[string]*gitfake.Repo{
		testRepo: {
			Webhooks:     map[int]*git.WebhookEntry{},
			UserCanWrite: map[string]bool{},

			PullRequests: map[int]*git.PullRequest{
				testPRID: {},
			},
			CommitStatuses: map[string][]git.CommitStatus{},
			Comments:       map[int][]git.IssueComment{},
		},
	}

	// /approve - fail - same user
	testJobApprove(t, handler, wh, ic, chatops.Command{Type: "approve"}, func() {
		repo := gitfake.Repos[testRepo]
		assert.Equal(t, 1, len(repo.Comments[testPRID]), "Comment length")
		assert.Equal(t, generateUserUnauthorizedComment(testUserName), repo.Comments[testPRID][0].Comment.Body, "Cannot approve comment")
		assert.Equal(t, 0, len(repo.PullRequests[testPRID].Labels), "Label length")
	})
	gitfake.Repos[testRepo].Comments[testPRID] = nil

	// /approve - fail - unauthorized
	newUserID := 111
	newUserName := "new-user"
	gitfake.Users[newUserName] = &git.User{ID: newUserID, Name: newUserName}
	gitfake.Repos[testRepo].UserCanWrite[newUserName] = false
	wh.Sender = *gitfake.Users[newUserName]
	wh.IssueComment.Author = wh.Sender
	testJobApprove(t, handler, wh, ic, chatops.Command{Type: "approve"}, func() {
		repo := gitfake.Repos[testRepo]
		assert.Equal(t, 1, len(repo.Comments[testPRID]), "Comment length")
		assert.Equal(t, generateUserUnauthorizedComment(newUserName), repo.Comments[testPRID][0].Comment.Body, "Cannot approve comment")
		assert.Equal(t, 0, len(repo.PullRequests[testPRID].Labels), "Label length")
	})
	gitfake.Repos[testRepo].Comments[testPRID] = nil

	// /approve asads - fail
	gitfake.Repos[testRepo].UserCanWrite[newUserName] = true
	testJobApprove(t, handler, wh, ic, chatops.Command{Type: "approve", Args: []string{"asd"}}, func() {
		repo := gitfake.Repos[testRepo]
		assert.Equal(t, 1, len(repo.Comments[testPRID]), "Comment length")
		assert.Equal(t, generateHelpComment(), repo.Comments[testPRID][0].Comment.Body, "Cannot approve comment")
		assert.Equal(t, 0, len(repo.PullRequests[testPRID].Labels), "Label length")
	})
	gitfake.Repos[testRepo].Comments[testPRID] = nil

	// /approve - success
	testJobApprove(t, handler, wh, ic, chatops.Command{Type: "approve"}, func() {
		repo := gitfake.Repos[testRepo]
		assert.Equal(t, 1, len(repo.Comments[testPRID]), "Comment length")
		assert.Equal(t, generateApprovedComment(newUserName), repo.Comments[testPRID][0].Comment.Body, "Successfully approved comment")
		assert.Equal(t, 1, len(repo.PullRequests[testPRID].Labels), "Label length")
		assert.Equal(t, "approved", repo.PullRequests[testPRID].Labels[0].Name, "Approved label exists")
	})
	gitfake.Repos[testRepo].Comments[testPRID] = nil

	// /approve cancel - success
	testJobApprove(t, handler, wh, ic, chatops.Command{Type: "approve", Args: []string{"cancel"}}, func() {
		repo := gitfake.Repos[testRepo]
		assert.Equal(t, 1, len(repo.Comments[testPRID]), "Comment length")
		assert.Equal(t, generateApproveCanceledComment(newUserName), repo.Comments[testPRID][0].Comment.Body, "Successfully approved comment")
		assert.Equal(t, 0, len(repo.PullRequests[testPRID].Labels), "Label length")
	})
}

type testApproveVerifier func()

func testJobApprove(t *testing.T, handler *Handler, wh *git.Webhook, ic *cicdv1.IntegrationConfig, command chatops.Command, verifyFunc testApproveVerifier) {
	if err := handler.HandleChatOps(command, wh, ic); err != nil {
		t.Fatal(err)
	}
	verifyFunc()
}

func buildTestConfigForApprove() *cicdv1.IntegrationConfig {
	return &cicdv1.IntegrationConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testConfigName,
			Namespace: testNamespace,
		},
		Spec: cicdv1.IntegrationConfigSpec{
			Git: cicdv1.GitConfig{
				Type:       cicdv1.GitTypeFake,
				Repository: testRepo,
				Token:      &cicdv1.GitToken{Value: "dummy"},
			},
		},
	}
}

func buildTestWebhookForApprove() *git.Webhook {
	return &git.Webhook{
		EventType: git.EventTypeIssueComment,
		Repo: git.Repository{
			Name: testRepo,
		},
		Sender: git.User{
			ID:    testUserID,
			Name:  testUserName,
			Email: testUserEmail,
		},
		IssueComment: &git.IssueComment{
			Comment: git.Comment{
				CreatedAt: &metav1.Time{Time: time.Now()},
			},
			Author: git.User{
				ID:    testUserID,
				Name:  testUserName,
				Email: testUserEmail,
			},
			Issue: git.Issue{
				PullRequest: &git.PullRequest{
					ID:    testPRID,
					Title: "test-pull-request",
					State: git.PullRequestStateOpen,
					Author: git.User{
						ID:    testUserID,
						Name:  testUserName,
						Email: testUserEmail,
					},
					URL: "https://github.com/tmax-cloud/cicd-operator/pulls/1",
					Base: git.Base{
						Ref: "master",
					},
					Head: git.Head{
						Ref: "new-feat",
						Sha: "sfoj39jfsidjf93jfsiljf20",
					},
				},
			},
		},
	}
}
