package blocker

import (
	"context"
	"github.com/bmizerany/assert"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	gitfake "github.com/tmax-cloud/cicd-operator/pkg/git/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"testing"
)

const (
	testICName      = "test-ic"
	testICNamespace = "default"
)

func TestSortPullRequestByID(t *testing.T) {
	prs := map[int]*PullRequest{
		13: {PullRequest: git.PullRequest{ID: 13}},
		6:  {PullRequest: git.PullRequest{ID: 6}},
		72: {PullRequest: git.PullRequest{ID: 72}},
	}

	sorted := sortPullRequestByID(prs)

	assert.Equal(t, 3, len(sorted), "Length")
	assert.Equal(t, 6, sorted[0].ID, "1st PR")
	assert.Equal(t, 13, sorted[1].ID, "2nd PR")
	assert.Equal(t, 72, sorted[2].ID, "3rd PR")
}

func TestBlocker_retestAndMergeOnePool(t *testing.T) {
	_, cli := mergeTestConfig()
	b := New(cli)
	pool := NewPRPool(testICNamespace, testICName)

	baseBranch := "master"
	baseSHA := "22ccae53032027186ba739dfaa473ee61a82b298"
	newBaseSHA := "32cd89e8d07e37ab26d8c735090ae763884283db"
	headBranch := "newnew"
	headSHA := "3196ccc37bcae94852079b04fcbfaf928341d6e9"
	prID := 12

	pr := &PullRequest{
		PullRequest: git.PullRequest{
			ID: prID,
			Base: git.Base{
				Ref: baseBranch,
				Sha: baseSHA,
			},
			Head: git.Head{
				Ref: headBranch,
				Sha: headSHA,
			},
		},
		BlockerStatus: git.CommitStatusStateSuccess,
		Statuses: map[string]git.CommitStatus{
			"test-1": {Context: "test-1", State: git.CommitStatusStateSuccess, Description: "Job is successful    BaseSHA:" + baseSHA},
		},
	}
	pool.PullRequests[prID] = pr
	pool.MergePool.Add(pr)

	gitfake.Branches = map[string]*git.Branch{
		baseBranch: {CommitID: baseSHA},
	}

	b.retestAndMergeOnePool(pool)
	ijList := &cicdv1.IntegrationJobList{}
	_ = cli.List(context.Background(), ijList)
	assert.Equal(t, 0, len(ijList.Items), "No IntegrationJob created")
	assert.Equal(t, false, pool.CurrentBatch != nil, "Current batch not created")

	// TEST 2 - not based on the latest sha
	gitfake.Branches[baseBranch].CommitID = newBaseSHA
	b.retestAndMergeOnePool(pool)
	ijList = &cicdv1.IntegrationJobList{}
	_ = cli.List(context.Background(), ijList)

	assert.Equal(t, 1, len(ijList.Items), "IntegrationJob created")
	assert.Equal(t, true, pool.CurrentBatch != nil, "Current batch created")

	// TEST 3 - after batch test completion
	ij := ijList.Items[0]
	ij.Status.State = cicdv1.IntegrationJobStateCompleted
	_ = cli.Update(context.Background(), &ij)
	b.retestAndMergeOnePool(pool)

	assert.Equal(t, 1, len(ijList.Items), "IntegrationJob not created")
	assert.Equal(t, false, pool.CurrentBatch != nil, "Current batch cleared")
}

func TestBlocker_handleBatch(t *testing.T) {
	ic, cli := mergeTestConfig()
	gitCli, _ := utils.GetGitCli(ic, cli)
	b := New(cli)
	pool := NewPRPool(testICNamespace, testICName)

	// TEST 1 - Batch IJ fails
	t.Run("batch_ij_fails", func(t *testing.T) {
		ij1 := &cicdv1.IntegrationJob{
			ObjectMeta: metav1.ObjectMeta{Name: "test-ij-1", Namespace: testICNamespace},
		}
		_ = cli.Create(context.Background(), ij1)
		ij1.Status.State = cicdv1.IntegrationJobStateFailed
		_ = cli.Status().Update(context.Background(), ij1)

		pool.CurrentBatch = &Batch{
			PRs: []*PullRequest{{PullRequest: git.PullRequest{ID: 12}}},
			Job: types.NamespacedName{Name: "test-ij-1", Namespace: testICNamespace},
		}
		if err := b.handleBatch(pool, ic, gitCli); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, true, pool.CurrentBatch == nil, "CurrentBatch cleared")
	})

	// TEST 2 - Batch IJ Succeeds
	t.Run("batch_ij_succeeds", func(t *testing.T) {
		ij2 := &cicdv1.IntegrationJob{
			ObjectMeta: metav1.ObjectMeta{Name: "test-ij-2", Namespace: testICNamespace},
		}
		_ = cli.Create(context.Background(), ij2)
		ij2.Status.State = cicdv1.IntegrationJobStateCompleted
		_ = cli.Status().Update(context.Background(), ij2)

		pool.CurrentBatch = &Batch{
			PRs: []*PullRequest{{PullRequest: git.PullRequest{ID: 12}}},
			Job: types.NamespacedName{Name: "test-ij-2", Namespace: testICNamespace},
		}
		if err := b.handleBatch(pool, ic, gitCli); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, true, pool.CurrentBatch == nil, "CurrentBatch cleared")
	})
}

type getMergeMethodTestCase struct {
	Labels         []git.IssueLabel
	ICMethod       git.MergeMethod
	ExpectedMethod git.MergeMethod
}

func TestGetMergeMethod(t *testing.T) {
	tc := map[string]getMergeMethodTestCase{
		"followICMerge": {
			Labels:         []git.IssueLabel{},
			ICMethod:       git.MergeMethodMerge,
			ExpectedMethod: git.MergeMethodMerge,
		},
		"followICSquash": {
			Labels:         []git.IssueLabel{},
			ICMethod:       git.MergeMethodSquash,
			ExpectedMethod: git.MergeMethodSquash,
		},
		"globalMerge": {
			Labels:         []git.IssueLabel{{Name: "global/merge-merge"}},
			ICMethod:       git.MergeMethodMerge,
			ExpectedMethod: git.MergeMethodMerge,
		},
		"globalMerge2": {
			Labels:         []git.IssueLabel{{Name: "global/merge-merge"}},
			ICMethod:       git.MergeMethodSquash,
			ExpectedMethod: git.MergeMethodMerge,
		},
		"globalSquash": {
			Labels:         []git.IssueLabel{{Name: "global/merge-squash"}},
			ICMethod:       git.MergeMethodSquash,
			ExpectedMethod: git.MergeMethodSquash,
		},
		"globalSquash2": {
			Labels:         []git.IssueLabel{{Name: "global/merge-squash"}},
			ICMethod:       git.MergeMethodMerge,
			ExpectedMethod: git.MergeMethodSquash,
		},
	}

	configs.MergeKindMergeLabel = "global/merge-merge"
	configs.MergeKindSquashLabel = "global/merge-squash"

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			pr := &PullRequest{}
			pr.Labels = c.Labels
			ic := &cicdv1.IntegrationConfig{}
			ic.Spec.MergeConfig = &cicdv1.MergeConfig{Method: c.ICMethod}

			method := getMergeMethod(pr, ic)

			assert.Equal(t, string(c.ExpectedMethod), string(method))
		})
	}
}

func mergeTestConfig() (*cicdv1.IntegrationConfig, client.Client) {
	if _, exist := os.LookupEnv("CI"); !exist {
		ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	}

	s := runtime.NewScheme()
	utilruntime.Must(cicdv1.AddToScheme(s))
	ic := &cicdv1.IntegrationConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testICName,
			Namespace: testICNamespace,
		},
		Spec: cicdv1.IntegrationConfigSpec{
			Git: cicdv1.GitConfig{
				Type:       cicdv1.GitTypeFake,
				Repository: testRepo,
				Token:      &cicdv1.GitToken{Value: "dummy"},
			},
			MergeConfig: &cicdv1.MergeConfig{
				Query: cicdv1.MergeQuery{
					Labels:          []string{},
					ApproveRequired: true,
					Checks:          []string{"test-1"},
				},
			},
			Jobs: cicdv1.IntegrationConfigJobs{
				PreSubmit: []cicdv1.Job{
					{Container: corev1.Container{Name: "test-1"}},
				},
			},
		},
	}
	return ic, fake.NewFakeClientWithScheme(s, ic)
}
