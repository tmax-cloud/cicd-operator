package gitlab

import (
	"github.com/bmizerany/assert"
	"github.com/gorilla/mux"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

const (
	sampleMRList = "[{\"id\":95464030,\"iid\":3,\"project_id\":25815215,\"title\":\"Newnew\",\"description\":\"\",\"state\":\"opened\",\"created_at\":\"2021-04-12T05:07:00.660Z\",\"updated_at\":\"2021-04-13T04:53:14.489Z\",\"merged_by\":null,\"merged_at\":null,\"closed_by\":null,\"closed_at\":null,\"target_branch\":\"master\",\"source_branch\":\"newnew\",\"user_notes_count\":2,\"upvotes\":0,\"downvotes\":0,\"author\":{\"id\":7169076,\"name\":\"Sunghyun Kim\",\"username\":\"cqbqdd11519\",\"state\":\"active\",\"avatar_url\":\"https://secure.gravatar.com/avatar/4021c3aaa995c31bd117cb7800005e85?s=80\\u0026d=identicon\",\"web_url\":\"https://gitlab.com/cqbqdd11519\"},\"assignees\":[],\"assignee\":null,\"reviewers\":[],\"source_project_id\":25815215,\"target_project_id\":25815215,\"labels\":[],\"work_in_progress\":false,\"milestone\":null,\"merge_when_pipeline_succeeds\":false,\"merge_status\":\"can_be_merged\",\"sha\":\"5f065c6de7dacb91aa5929a5c0ab71ecba5456b0\",\"merge_commit_sha\":null,\"squash_commit_sha\":null,\"discussion_locked\":null,\"should_remove_source_branch\":null,\"force_remove_source_branch\":false,\"reference\":\"!3\",\"references\":{\"short\":\"!3\",\"relative\":\"!3\",\"full\":\"cqbqdd11519/cicd-test!3\"},\"web_url\":\"https://gitlab.com/cqbqdd11519/cicd-test/-/merge_requests/3\",\"time_stats\":{\"time_estimate\":0,\"total_time_spent\":0,\"human_time_estimate\":null,\"human_total_time_spent\":null},\"squash\":false,\"task_completion_status\":{\"count\":0,\"completed_count\":0},\"has_conflicts\":false,\"blocking_discussions_resolved\":true,\"approvals_before_merge\":null},{\"id\":95463922,\"iid\":2,\"project_id\":25815215,\"title\":\"Newnew\",\"description\":\"\",\"state\":\"closed\",\"created_at\":\"2021-04-12T05:05:06.339Z\",\"updated_at\":\"2021-04-12T05:05:42.049Z\",\"merged_by\":null,\"merged_at\":null,\"closed_by\":{\"id\":7169076,\"name\":\"Sunghyun Kim\",\"username\":\"cqbqdd11519\",\"state\":\"active\",\"avatar_url\":\"https://secure.gravatar.com/avatar/4021c3aaa995c31bd117cb7800005e85?s=80\\u0026d=identicon\",\"web_url\":\"https://gitlab.com/cqbqdd11519\"},\"closed_at\":\"2021-04-12T05:05:42.070Z\",\"target_branch\":\"master\",\"source_branch\":\"newnew\",\"user_notes_count\":0,\"upvotes\":0,\"downvotes\":0,\"author\":{\"id\":7169076,\"name\":\"Sunghyun Kim\",\"username\":\"cqbqdd11519\",\"state\":\"active\",\"avatar_url\":\"https://secure.gravatar.com/avatar/4021c3aaa995c31bd117cb7800005e85?s=80\\u0026d=identicon\",\"web_url\":\"https://gitlab.com/cqbqdd11519\"},\"assignees\":[],\"assignee\":null,\"reviewers\":[],\"source_project_id\":25815215,\"target_project_id\":25815215,\"labels\":[],\"work_in_progress\":false,\"milestone\":null,\"merge_when_pipeline_succeeds\":false,\"merge_status\":\"can_be_merged\",\"sha\":\"dace98c2d0437f6ccacd8b9c8094f4dde9162214\",\"merge_commit_sha\":null,\"squash_commit_sha\":null,\"discussion_locked\":null,\"should_remove_source_branch\":null,\"force_remove_source_branch\":false,\"reference\":\"!2\",\"references\":{\"short\":\"!2\",\"relative\":\"!2\",\"full\":\"cqbqdd11519/cicd-test!2\"},\"web_url\":\"https://gitlab.com/cqbqdd11519/cicd-test/-/merge_requests/2\",\"time_stats\":{\"time_estimate\":0,\"total_time_spent\":0,\"human_time_estimate\":null,\"human_total_time_spent\":null},\"squash\":false,\"task_completion_status\":{\"count\":0,\"completed_count\":0},\"has_conflicts\":false,\"blocking_discussions_resolved\":true,\"approvals_before_merge\":null},{\"id\":95462727,\"iid\":1,\"project_id\":25815215,\"title\":\"newnew\",\"description\":\"\",\"state\":\"closed\",\"created_at\":\"2021-04-12T04:42:18.407Z\",\"updated_at\":\"2021-04-12T04:58:53.632Z\",\"merged_by\":null,\"merged_at\":null,\"closed_by\":{\"id\":7169076,\"name\":\"Sunghyun Kim\",\"username\":\"cqbqdd11519\",\"state\":\"active\",\"avatar_url\":\"https://secure.gravatar.com/avatar/4021c3aaa995c31bd117cb7800005e85?s=80\\u0026d=identicon\",\"web_url\":\"https://gitlab.com/cqbqdd11519\"},\"closed_at\":\"2021-04-12T04:58:53.649Z\",\"target_branch\":\"master\",\"source_branch\":\"newnew\",\"user_notes_count\":0,\"upvotes\":0,\"downvotes\":0,\"author\":{\"id\":7169076,\"name\":\"Sunghyun Kim\",\"username\":\"cqbqdd11519\",\"state\":\"active\",\"avatar_url\":\"https://secure.gravatar.com/avatar/4021c3aaa995c31bd117cb7800005e85?s=80\\u0026d=identicon\",\"web_url\":\"https://gitlab.com/cqbqdd11519\"},\"assignees\":[{\"id\":7169076,\"name\":\"Sunghyun Kim\",\"username\":\"cqbqdd11519\",\"state\":\"active\",\"avatar_url\":\"https://secure.gravatar.com/avatar/4021c3aaa995c31bd117cb7800005e85?s=80\\u0026d=identicon\",\"web_url\":\"https://gitlab.com/cqbqdd11519\"}],\"assignee\":{\"id\":7169076,\"name\":\"Sunghyun Kim\",\"username\":\"cqbqdd11519\",\"state\":\"active\",\"avatar_url\":\"https://secure.gravatar.com/avatar/4021c3aaa995c31bd117cb7800005e85?s=80\\u0026d=identicon\",\"web_url\":\"https://gitlab.com/cqbqdd11519\"},\"reviewers\":[],\"source_project_id\":25815215,\"target_project_id\":25815215,\"labels\":[\"kind/test\"],\"work_in_progress\":false,\"milestone\":null,\"merge_when_pipeline_succeeds\":false,\"merge_status\":\"unchecked\",\"sha\":\"e703f64f722f33c4fbb1f326aed08edc81053b0b\",\"merge_commit_sha\":null,\"squash_commit_sha\":null,\"discussion_locked\":null,\"should_remove_source_branch\":null,\"force_remove_source_branch\":false,\"reference\":\"!1\",\"references\":{\"short\":\"!1\",\"relative\":\"!1\",\"full\":\"cqbqdd11519/cicd-test!1\"},\"web_url\":\"https://gitlab.com/cqbqdd11519/cicd-test/-/merge_requests/1\",\"time_stats\":{\"time_estimate\":0,\"total_time_spent\":0,\"human_time_estimate\":null,\"human_total_time_spent\":null},\"squash\":false,\"task_completion_status\":{\"count\":0,\"completed_count\":0},\"has_conflicts\":false,\"blocking_discussions_resolved\":true,\"approvals_before_merge\":null}]"
)

func TestClient_ListPullRequests(t *testing.T) {
	c, err := testEnv()
	if err != nil {
		t.Fatal(err)
	}

	prs, err := c.ListPullRequests(false)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 3, len(prs), "Number of prs")
	assert.Equal(t, 95464030, prs[0].ID, "PR ID")
	assert.Equal(t, "Newnew", prs[0].Title, "PR Title")
	assert.Equal(t, 95463922, prs[1].ID, "PR ID")
	assert.Equal(t, "Newnew", prs[1].Title, "PR Title")
	assert.Equal(t, 95462727, prs[2].ID, "PR ID")
	assert.Equal(t, "newnew", prs[2].Title, "PR Title")
}

func testEnv() (*Client, error) {
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(req.URL.String()))
	})
	r.HandleFunc("/api/v4/projects/{org}/{repo}/merge_requests", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(sampleMRList))
	})
	testSrv := httptest.NewServer(r)

	s := runtime.NewScheme()
	utilruntime.Must(cicdv1.AddToScheme(s))

	ic := &cicdv1.IntegrationConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ic",
			Namespace: "default",
		},
		Spec: cicdv1.IntegrationConfigSpec{
			Git: cicdv1.GitConfig{
				Type:       "gitlab",
				Repository: "tmax-cloud/cicd-test",
				APIUrl:     testSrv.URL,
				Token:      &cicdv1.GitToken{Value: "dummy"},
			},
		},
	}

	c := &Client{
		IntegrationConfig: ic,
		K8sClient:         fake.NewFakeClientWithScheme(s, ic),
	}
	if err := c.Init(); err != nil {
		return nil, err
	}

	return c, nil
}
