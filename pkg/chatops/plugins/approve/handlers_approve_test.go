package approve

import (
	"github.com/stretchr/testify/require"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/chatops"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	gitfake "github.com/tmax-cloud/cicd-operator/pkg/git/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
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

	testUser2ID    = 111
	testUser2Name  = "new-user"
	testUser2Email = "new@test.com"
)

type approvalTestCase struct {
	preFunc    func(wh *git.Webhook)
	verifyFunc func(t *testing.T)
}

func TestHandler_Handle(t *testing.T) {
	if _, exist := os.LookupEnv("CI"); !exist {
		ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	}
	s := runtime.NewScheme()
	utilruntime.Must(cicdv1.AddToScheme(s))

	ic := buildTestConfigForApprove()
	fakeCli := fake.NewFakeClientWithScheme(s, ic)
	handler := &Handler{Client: fakeCli}

	tc := map[string]approvalTestCase{
		"noop": {
			preFunc: func(wh *git.Webhook) {
				wh.EventType = git.EventTypePullRequest
			},
			verifyFunc: func(t *testing.T) {
				repo := gitfake.Repos[testRepo]
				require.Len(t, repo.Comments[testPRID], 0, "Comment length")
				require.Len(t, repo.PullRequests[testPRID].Labels, 0, "Label length")
			},
		},
		"successApprove": {
			preFunc: func(wh *git.Webhook) {
				gitfake.Repos[testRepo].UserCanWrite[testUser2Name] = true
				wh.Sender = *gitfake.Users[testUser2Name]
				wh.IssueComment.Author = wh.Sender
			},
			verifyFunc: func(t *testing.T) {
				repo := gitfake.Repos[testRepo]
				require.Len(t, repo.Comments[testPRID], 1, "Comment length")
				require.Equal(t, generateApprovedComment(testUser2Name), repo.Comments[testPRID][0].Comment.Body, "Successfully approved comment")
				require.Len(t, repo.PullRequests[testPRID].Labels, 1, "Label length")
				require.Equal(t, "approved", repo.PullRequests[testPRID].Labels[0].Name, "Approved label exists")
			},
		},
		"successApproveCancel": {
			preFunc: func(wh *git.Webhook) {
				gitfake.Repos[testRepo].UserCanWrite[testUser2Name] = true
				gitfake.Repos[testRepo].PullRequests[testPRID].Labels = append(gitfake.Repos[testRepo].PullRequests[testPRID].Labels, git.IssueLabel{Name: "approved"})
				wh.Sender = *gitfake.Users[testUser2Name]
				wh.IssueComment.Author = wh.Sender
				wh.IssueComment.ReviewState = git.PullRequestReviewStateUnapproved
			},
			verifyFunc: func(t *testing.T) {
				repo := gitfake.Repos[testRepo]
				require.Len(t, repo.Comments[testPRID], 1, "Comment length")
				require.Equal(t, generateApproveCanceledComment(testUser2Name), repo.Comments[testPRID][0].Comment.Body, "Successfully approved comment")
				require.Len(t, repo.PullRequests[testPRID].Labels, 0, "Label length")
			},
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			// Init fake git
			initFakeGit()

			// Initialize webhook
			wh := buildTestWebhookApprove()
			c.preFunc(wh)

			err := handler.Handle(wh, ic)
			require.NoError(t, err)
			c.verifyFunc(t)
		})
	}
}

type chatOpsApprovalTestCase struct {
	command    chatops.Command
	preFunc    func(wh *git.Webhook)
	verifyFunc func(t *testing.T)
}

func TestChatOps_handleApprove(t *testing.T) {
	if _, exist := os.LookupEnv("CI"); !exist {
		ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	}
	s := runtime.NewScheme()
	utilruntime.Must(cicdv1.AddToScheme(s))

	ic := buildTestConfigForApprove()
	fakeCli := fake.NewFakeClientWithScheme(s, ic)
	handler := &Handler{Client: fakeCli}

	tc := map[string]chatOpsApprovalTestCase{
		"failSameUser": {
			command: chatops.Command{Type: "approve"},
			preFunc: func(_ *git.Webhook) {},
			verifyFunc: func(t *testing.T) {
				repo := gitfake.Repos[testRepo]
				require.Len(t, repo.Comments[testPRID], 1, "Comment length")
				require.Equal(t, generateUserUnauthorizedComment(testUserName), repo.Comments[testPRID][0].Comment.Body, "Cannot approve comment")
				require.Len(t, repo.PullRequests[testPRID].Labels, 0, "Label length")
			},
		},
		"failUnauthorized": {
			command: chatops.Command{Type: "approve"},
			preFunc: func(wh *git.Webhook) {
				gitfake.Repos[testRepo].UserCanWrite[testUser2Name] = false
				wh.Sender = *gitfake.Users[testUser2Name]
				wh.IssueComment.Author = wh.Sender
			},
			verifyFunc: func(t *testing.T) {
				repo := gitfake.Repos[testRepo]
				require.Len(t, repo.Comments[testPRID], 1, "Comment length")
				require.Equal(t, generateUserUnauthorizedComment(testUser2Name), repo.Comments[testPRID][0].Comment.Body, "Cannot approve comment")
				require.Len(t, repo.PullRequests[testPRID].Labels, 0, "Label length")
			},
		},
		"failMalformedCommand": {
			command: chatops.Command{Type: "approve", Args: []string{"asd"}},
			preFunc: func(wh *git.Webhook) {
				gitfake.Repos[testRepo].UserCanWrite[testUser2Name] = true
				wh.Sender = *gitfake.Users[testUser2Name]
				wh.IssueComment.Author = wh.Sender
			},
			verifyFunc: func(t *testing.T) {
				repo := gitfake.Repos[testRepo]
				require.Len(t, repo.Comments[testPRID], 1, "Comment length")
				require.Equal(t, generateHelpComment(), repo.Comments[testPRID][0].Comment.Body, "Cannot approve comment")
				require.Len(t, repo.PullRequests[testPRID].Labels, 0, "Label length")
			},
		},
		"successApprove": {
			command: chatops.Command{Type: "approve"},
			preFunc: func(wh *git.Webhook) {
				gitfake.Repos[testRepo].UserCanWrite[testUser2Name] = true
				wh.Sender = *gitfake.Users[testUser2Name]
				wh.IssueComment.Author = wh.Sender
			},
			verifyFunc: func(t *testing.T) {
				repo := gitfake.Repos[testRepo]
				require.Len(t, repo.Comments[testPRID], 1, "Comment length")
				require.Equal(t, generateApprovedComment(testUser2Name), repo.Comments[testPRID][0].Comment.Body, "Successfully approved comment")
				require.Len(t, repo.PullRequests[testPRID].Labels, 1, "Label length")
				require.Equal(t, "approved", repo.PullRequests[testPRID].Labels[0].Name, "Approved label exists")
			},
		},
		"successApproveCancel": {
			command: chatops.Command{Type: "approve", Args: []string{"cancel"}},
			preFunc: func(wh *git.Webhook) {
				gitfake.Repos[testRepo].UserCanWrite[testUser2Name] = true
				gitfake.Repos[testRepo].PullRequests[testPRID].Labels = append(gitfake.Repos[testRepo].PullRequests[testPRID].Labels, git.IssueLabel{Name: "approved"})
				wh.Sender = *gitfake.Users[testUser2Name]
				wh.IssueComment.Author = wh.Sender
			},
			verifyFunc: func(t *testing.T) {
				repo := gitfake.Repos[testRepo]
				require.Len(t, repo.Comments[testPRID], 1, "Comment length")
				require.Equal(t, generateApproveCanceledComment(testUser2Name), repo.Comments[testPRID][0].Comment.Body, "Successfully approved comment")
				require.Len(t, repo.PullRequests[testPRID].Labels, 0, "Label length")
			},
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			// Init fake git
			initFakeGit()

			// Initialize webhook
			wh := buildTestWebhookCommentApprove()
			c.preFunc(wh)

			err := handler.HandleChatOps(c.command, wh, ic)
			require.NoError(t, err)
			c.verifyFunc(t)
		})
	}
}

func initFakeGit() {
	gitfake.Users = map[string]*git.User{
		testUserName:  {ID: testUserID, Name: testUserName, Email: testUserEmail},
		testUser2Name: {ID: testUser2ID, Name: testUser2Name, Email: testUser2Email},
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

func buildTestWebhookCommentApprove() *git.Webhook {
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

func buildTestWebhookApprove() *git.Webhook {
	return &git.Webhook{
		EventType: git.EventTypePullRequestReview,
		Repo: git.Repository{
			Name: testRepo,
		},
		Sender: git.User{
			ID:    testUserID,
			Name:  testUserName,
			Email: testUserEmail,
		},
		IssueComment: &git.IssueComment{
			ReviewState: git.PullRequestReviewStateApproved,
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
