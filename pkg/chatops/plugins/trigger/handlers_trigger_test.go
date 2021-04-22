package trigger

import (
	"context"
	"github.com/bmizerany/assert"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/chatops"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
	"time"
)

const (
	testNamespace  = "default"
	testConfigName = "test-ic"

	testUserID    = 32
	testUserName  = "test-user"
	testUserEmail = "test@test.com"
)

func TestChatOps_handleTrigger(t *testing.T) {
	s := runtime.NewScheme()
	utilruntime.Must(cicdv1.AddToScheme(s))

	ic := buildTestJobs()
	wh := buildTestWebhookForTrigger()

	fakeCli := fake.NewFakeClientWithScheme(s, ic)

	handler := &Handler{Client: fakeCli}

	// /retest
	testJobTrigger(t, handler, fakeCli, wh, ic, chatops.Command{Type: "retest"}, func(ij *cicdv1.IntegrationJob) {
		assert.Equal(t, 6, len(ij.Spec.Jobs))
	})

	// /test
	testJobTrigger(t, handler, fakeCli, wh, ic, chatops.Command{Type: "test"}, func(ij *cicdv1.IntegrationJob) {
		assert.Equal(t, 6, len(ij.Spec.Jobs))
	})

	// /test a-1
	testJobTrigger(t, handler, fakeCli, wh, ic, chatops.Command{Type: "test", Args: []string{"a-1"}}, func(ij *cicdv1.IntegrationJob) {
		assert.Equal(t, 1, len(ij.Spec.Jobs))
		assert.Equal(t, "a-1", ij.Spec.Jobs[0].Name)
	})

	// /test a-2
	testJobTrigger(t, handler, fakeCli, wh, ic, chatops.Command{Type: "test", Args: []string{"a-2"}}, func(ij *cicdv1.IntegrationJob) {
		assert.Equal(t, 2, len(ij.Spec.Jobs))
		assert.Equal(t, "a-1", ij.Spec.Jobs[0].Name)
		assert.Equal(t, "a-2", ij.Spec.Jobs[1].Name)
		assert.Equal(t, 1, len(ij.Spec.Jobs[1].After))
		assert.Equal(t, "a-1", ij.Spec.Jobs[1].After[0])
	})

	// /test b-2
	testJobTrigger(t, handler, fakeCli, wh, ic, chatops.Command{Type: "test", Args: []string{"b-2"}}, func(ij *cicdv1.IntegrationJob) {
		assert.Equal(t, 2, len(ij.Spec.Jobs))
		assert.Equal(t, "b-1", ij.Spec.Jobs[0].Name)
		assert.Equal(t, "b-2", ij.Spec.Jobs[1].Name)
		assert.Equal(t, 1, len(ij.Spec.Jobs[1].After))
		assert.Equal(t, "b-1", ij.Spec.Jobs[1].After[0])
	})

	// /test a-4
	testJobTrigger(t, handler, fakeCli, wh, ic, chatops.Command{Type: "test", Args: []string{"a-4"}}, func(ij *cicdv1.IntegrationJob) {
		assert.Equal(t, 4, len(ij.Spec.Jobs))
		assert.Equal(t, "a-1", ij.Spec.Jobs[0].Name)
		assert.Equal(t, "a-2", ij.Spec.Jobs[1].Name)
		assert.Equal(t, 1, len(ij.Spec.Jobs[1].After))
		assert.Equal(t, "a-1", ij.Spec.Jobs[1].After[0])
		assert.Equal(t, "a-3", ij.Spec.Jobs[2].Name)
		assert.Equal(t, 1, len(ij.Spec.Jobs[2].After))
		assert.Equal(t, "a-1", ij.Spec.Jobs[2].After[0])
		assert.Equal(t, "a-4", ij.Spec.Jobs[3].Name)
		assert.Equal(t, 2, len(ij.Spec.Jobs[3].After))
		assert.Equal(t, "a-2", ij.Spec.Jobs[3].After[0])
		assert.Equal(t, "a-3", ij.Spec.Jobs[3].After[1])
	})
}

type testTriggerVerifier func(ij *cicdv1.IntegrationJob)

func testJobTrigger(t *testing.T, handler *Handler, fakeCli client.Client, wh *git.Webhook, ic *cicdv1.IntegrationConfig, command chatops.Command, verifyFunc testTriggerVerifier) {
	var ijList cicdv1.IntegrationJobList
	if err := handler.HandleChatOps(command, wh, ic); err != nil {
		t.Fatal(err)
	}
	if err := fakeCli.List(context.Background(), &ijList); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(ijList.Items))
	verifyFunc(&ijList.Items[0])
	if err := fakeCli.Delete(context.Background(), &ijList.Items[0]); err != nil {
		t.Fatal(err)
	}
}

func buildTestJobs() *cicdv1.IntegrationConfig {
	// Complicated jobs
	/*
		a-1  -->  a-2  --> a-4
		     \->  a-3  -/

		b-1  -->  b-2
	*/
	ic := &cicdv1.IntegrationConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testConfigName,
			Namespace: testNamespace,
		},
	}

	jobA1 := cicdv1.Job{}
	jobA1.Name = "a-1"
	ic.Spec.Jobs.PreSubmit = append(ic.Spec.Jobs.PreSubmit, jobA1)

	jobA2 := cicdv1.Job{}
	jobA2.Name = "a-2"
	jobA2.After = []string{"a-1"}
	ic.Spec.Jobs.PreSubmit = append(ic.Spec.Jobs.PreSubmit, jobA2)

	jobA3 := cicdv1.Job{}
	jobA3.Name = "a-3"
	jobA3.After = []string{"a-1"}
	ic.Spec.Jobs.PreSubmit = append(ic.Spec.Jobs.PreSubmit, jobA3)

	jobA4 := cicdv1.Job{}
	jobA4.Name = "a-4"
	jobA4.After = []string{"a-2", "a-3"}
	ic.Spec.Jobs.PreSubmit = append(ic.Spec.Jobs.PreSubmit, jobA4)

	jobB1 := cicdv1.Job{}
	jobB1.Name = "b-1"
	ic.Spec.Jobs.PreSubmit = append(ic.Spec.Jobs.PreSubmit, jobB1)

	jobB2 := cicdv1.Job{}
	jobB2.Name = "b-2"
	jobB2.After = []string{"b-1"}
	ic.Spec.Jobs.PreSubmit = append(ic.Spec.Jobs.PreSubmit, jobB2)

	return ic
}

func buildTestWebhookForTrigger() *git.Webhook {
	return &git.Webhook{
		EventType: git.EventTypeIssueComment,
		Repo: git.Repository{
			Name: "tmax-cloud/cicd-operator",
			URL:  "https://github.com/tmax-cloud/cicd-operator",
		},
		IssueComment: &git.IssueComment{
			Comment: git.Comment{
				CreatedAt: &metav1.Time{Time: time.Now()},
			},
			Issue: git.Issue{
				PullRequest: &git.PullRequest{
					ID:    0,
					Title: "test-pull-request",
					State: git.PullRequestStateOpen,
					Sender: git.User{
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
			Sender: git.User{
				ID:    testUserID,
				Name:  testUserName,
				Email: testUserEmail,
			},
		},
	}
}
