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

package hold

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/pkg/chatops"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	gitfake "github.com/tmax-cloud/cicd-operator/pkg/git/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
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

type chatOpsHoldTestCase struct {
	command    chatops.Command
	preFunc    func(wh *git.Webhook)
	verifyFunc func(t *testing.T)
}

func TestHandler_HandleChatOps(t *testing.T) {
	if _, exist := os.LookupEnv("CI"); !exist {
		ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	}
	s := runtime.NewScheme()
	utilruntime.Must(cicdv1.AddToScheme(s))

	ic := buildTestConfigForHold()
	fakeCli := fake.NewClientBuilder().WithScheme(s).WithObjects(ic).Build()
	handler := &Handler{Client: fakeCli}

	// Set configs value
	configs.MergeBlockLabel = "ci/hold"

	tc := map[string]chatOpsHoldTestCase{
		"hold": {
			command: chatops.Command{Type: "hold", Args: []string{}},
			preFunc: func(wh *git.Webhook) {},
			verifyFunc: func(t *testing.T) {
				require.Len(t, gitfake.Repos[testRepo].PullRequests[testPRID].Labels, 1)
				require.Equal(t, configs.MergeBlockLabel, gitfake.Repos[testRepo].PullRequests[testPRID].Labels[0].Name)
			},
		},
		"holdCancel": {
			command: chatops.Command{Type: "hold", Args: []string{"cancel"}},
			preFunc: func(wh *git.Webhook) {
				gitfake.Repos[testRepo].PullRequests[testPRID].Labels = append(gitfake.Repos[testRepo].PullRequests[testPRID].Labels, git.IssueLabel{Name: configs.MergeBlockLabel})
			},
			verifyFunc: func(t *testing.T) {
				require.Len(t, gitfake.Repos[testRepo].PullRequests[testPRID].Labels, 0)
			},
		},
		"failMalformed": {
			command: chatops.Command{Type: "hold", Args: []string{"cancellllll"}},
			preFunc: func(wh *git.Webhook) {},
			verifyFunc: func(t *testing.T) {
				require.Len(t, gitfake.Repos[testRepo].PullRequests[testPRID].Labels, 0)
				require.Len(t, gitfake.Repos[testRepo].Comments[testPRID], 1)
				require.Equal(t, "[HOLD ALERT]\n\nHold comment is malformed\n\nYou can hold or cancel hold the pull request by commenting...\n- `/hold`\n- `/hold cancel`\n", gitfake.Repos[testRepo].Comments[testPRID][0].Comment.Body)
			},
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			// Init fake git
			initFakeGit()

			// Initialize webhook
			wh := buildTestWebhookCommentHold()
			c.preFunc(wh)

			err := handler.HandleChatOps(c.command, wh, ic)
			require.NoError(t, err)
			c.verifyFunc(t)
		})
	}
}

func initFakeGit() {
	gitfake.Users = map[string]*git.User{
		testUserName: {ID: testUserID, Name: testUserName, Email: testUserEmail},
	}
	gitfake.Repos = map[string]*gitfake.Repo{
		testRepo: {
			PullRequests: map[int]*git.PullRequest{
				testPRID: {},
			},
			Comments: map[int][]git.IssueComment{
				testPRID: nil,
			},
		},
	}
}

func buildTestConfigForHold() *cicdv1.IntegrationConfig {
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

func buildTestWebhookCommentHold() *git.Webhook {
	return &git.Webhook{
		EventType: git.EventTypeIssueComment,
		Repo: git.Repository{
			Name: testRepo,
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
