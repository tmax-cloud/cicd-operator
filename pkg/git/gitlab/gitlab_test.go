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

package gitlab

import (
	"fmt"

	"github.com/bmizerany/assert"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
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
	sampleWebhooksList = "[{\"id\":7194623,\"url\":\"http://asdasd/webhook/default/chatops-test-gitlab\",\"created_at\":\"2021-04-12T04:35:27.210Z\",\"push_events\":true,\"tag_push_events\":true,\"merge_requests_events\":true,\"repository_update_events\":false,\"enable_ssl_verification\":false,\"project_id\":25815215,\"issues_events\":true,\"confidential_issues_events\":true,\"note_events\":true,\"confidential_note_events\":true,\"pipeline_events\":true,\"wiki_page_events\":true,\"deployment_events\":true,\"job_events\":true,\"releases_events\":false,\"push_events_branch_filter\":null}]"
	sampleStatusesList = "[{\"id\":1170837740,\"sha\":\"5f065c6de7dacb91aa5929a5c0ab71ecba5456b0\",\"ref\":\"newnew\",\"status\":\"running\",\"name\":\"blocker\",\"target_url\":\"http://a\",\"description\":\"PR does not meet all conditions. Label lgtm is required. Checks [blocker] are not met. \",\"created_at\":\"2021-04-12T05:40:07.995Z\",\"started_at\":\"2021-04-12T05:40:08.028Z\",\"finished_at\":null,\"allow_failure\":false,\"coverage\":null,\"author\":{\"id\":7169076,\"name\":\"Sunghyun Kim\",\"username\":\"cqbqdd11519\",\"state\":\"active\",\"avatar_url\":\"https://secure.gravatar.com/avatar/4021c3aaa995c31bd117cb7800005e85?s=80\\u0026d=identicon\",\"web_url\":\"https://gitlab.com/cqbqdd11519\"}},{\"id\":1171264736,\"sha\":\"5f065c6de7dacb91aa5929a5c0ab71ecba5456b0\",\"ref\":\"newnew\",\"status\":\"success\",\"name\":\"test-1\",\"target_url\":\"http://cicd-local.vingsu.com:8080/report/default/chatops-test-gitlab-5f065-cmiyw/test-1\",\"description\":\"All Steps have completed executing\",\"created_at\":\"2021-04-12T08:38:29.773Z\",\"started_at\":\"2021-04-12T08:38:29.819Z\",\"finished_at\":\"2021-04-12T08:38:51.996Z\",\"allow_failure\":false,\"coverage\":null,\"author\":{\"id\":7169076,\"name\":\"Sunghyun Kim\",\"username\":\"cqbqdd11519\",\"state\":\"active\",\"avatar_url\":\"https://secure.gravatar.com/avatar/4021c3aaa995c31bd117cb7800005e85?s=80\\u0026d=identicon\",\"web_url\":\"https://gitlab.com/cqbqdd11519\"}}]"
	sampleMRList       = "[{\"id\":95464030,\"iid\":3,\"project_id\":25815215,\"title\":\"Newnew\",\"description\":\"\",\"state\":\"opened\",\"created_at\":\"2021-04-12T05:07:00.660Z\",\"updated_at\":\"2021-04-13T04:53:14.489Z\",\"merged_by\":null,\"merged_at\":null,\"closed_by\":null,\"closed_at\":null,\"target_branch\":\"master\",\"source_branch\":\"newnew\",\"user_notes_count\":2,\"upvotes\":0,\"downvotes\":0,\"author\":{\"id\":7169076,\"name\":\"Sunghyun Kim\",\"username\":\"cqbqdd11519\",\"state\":\"active\",\"avatar_url\":\"https://secure.gravatar.com/avatar/4021c3aaa995c31bd117cb7800005e85?s=80\\u0026d=identicon\",\"web_url\":\"https://gitlab.com/cqbqdd11519\"},\"assignees\":[],\"assignee\":null,\"reviewers\":[],\"source_project_id\":25815215,\"target_project_id\":25815215,\"labels\":[],\"work_in_progress\":false,\"milestone\":null,\"merge_when_pipeline_succeeds\":false,\"merge_status\":\"can_be_merged\",\"sha\":\"5f065c6de7dacb91aa5929a5c0ab71ecba5456b0\",\"merge_commit_sha\":null,\"squash_commit_sha\":null,\"discussion_locked\":null,\"should_remove_source_branch\":null,\"force_remove_source_branch\":false,\"reference\":\"!3\",\"references\":{\"short\":\"!3\",\"relative\":\"!3\",\"full\":\"cqbqdd11519/cicd-test!3\"},\"web_url\":\"https://gitlab.com/cqbqdd11519/cicd-test/-/merge_requests/3\",\"time_stats\":{\"time_estimate\":0,\"total_time_spent\":0,\"human_time_estimate\":null,\"human_total_time_spent\":null},\"squash\":false,\"task_completion_status\":{\"count\":0,\"completed_count\":0},\"has_conflicts\":false,\"blocking_discussions_resolved\":true,\"approvals_before_merge\":null},{\"id\":95463922,\"iid\":2,\"project_id\":25815215,\"title\":\"Newnew\",\"description\":\"\",\"state\":\"closed\",\"created_at\":\"2021-04-12T05:05:06.339Z\",\"updated_at\":\"2021-04-12T05:05:42.049Z\",\"merged_by\":null,\"merged_at\":null,\"closed_by\":{\"id\":7169076,\"name\":\"Sunghyun Kim\",\"username\":\"cqbqdd11519\",\"state\":\"active\",\"avatar_url\":\"https://secure.gravatar.com/avatar/4021c3aaa995c31bd117cb7800005e85?s=80\\u0026d=identicon\",\"web_url\":\"https://gitlab.com/cqbqdd11519\"},\"closed_at\":\"2021-04-12T05:05:42.070Z\",\"target_branch\":\"master\",\"source_branch\":\"newnew\",\"user_notes_count\":0,\"upvotes\":0,\"downvotes\":0,\"author\":{\"id\":7169076,\"name\":\"Sunghyun Kim\",\"username\":\"cqbqdd11519\",\"state\":\"active\",\"avatar_url\":\"https://secure.gravatar.com/avatar/4021c3aaa995c31bd117cb7800005e85?s=80\\u0026d=identicon\",\"web_url\":\"https://gitlab.com/cqbqdd11519\"},\"assignees\":[],\"assignee\":null,\"reviewers\":[],\"source_project_id\":25815215,\"target_project_id\":25815215,\"labels\":[],\"work_in_progress\":false,\"milestone\":null,\"merge_when_pipeline_succeeds\":false,\"merge_status\":\"can_be_merged\",\"sha\":\"dace98c2d0437f6ccacd8b9c8094f4dde9162214\",\"merge_commit_sha\":null,\"squash_commit_sha\":null,\"discussion_locked\":null,\"should_remove_source_branch\":null,\"force_remove_source_branch\":false,\"reference\":\"!2\",\"references\":{\"short\":\"!2\",\"relative\":\"!2\",\"full\":\"cqbqdd11519/cicd-test!2\"},\"web_url\":\"https://gitlab.com/cqbqdd11519/cicd-test/-/merge_requests/2\",\"time_stats\":{\"time_estimate\":0,\"total_time_spent\":0,\"human_time_estimate\":null,\"human_total_time_spent\":null},\"squash\":false,\"task_completion_status\":{\"count\":0,\"completed_count\":0},\"has_conflicts\":false,\"blocking_discussions_resolved\":true,\"approvals_before_merge\":null},{\"id\":95462727,\"iid\":1,\"project_id\":25815215,\"title\":\"newnew\",\"description\":\"\",\"state\":\"closed\",\"created_at\":\"2021-04-12T04:42:18.407Z\",\"updated_at\":\"2021-04-12T04:58:53.632Z\",\"merged_by\":null,\"merged_at\":null,\"closed_by\":{\"id\":7169076,\"name\":\"Sunghyun Kim\",\"username\":\"cqbqdd11519\",\"state\":\"active\",\"avatar_url\":\"https://secure.gravatar.com/avatar/4021c3aaa995c31bd117cb7800005e85?s=80\\u0026d=identicon\",\"web_url\":\"https://gitlab.com/cqbqdd11519\"},\"closed_at\":\"2021-04-12T04:58:53.649Z\",\"target_branch\":\"master\",\"source_branch\":\"newnew\",\"user_notes_count\":0,\"upvotes\":0,\"downvotes\":0,\"author\":{\"id\":7169076,\"name\":\"Sunghyun Kim\",\"username\":\"cqbqdd11519\",\"state\":\"active\",\"avatar_url\":\"https://secure.gravatar.com/avatar/4021c3aaa995c31bd117cb7800005e85?s=80\\u0026d=identicon\",\"web_url\":\"https://gitlab.com/cqbqdd11519\"},\"assignees\":[{\"id\":7169076,\"name\":\"Sunghyun Kim\",\"username\":\"cqbqdd11519\",\"state\":\"active\",\"avatar_url\":\"https://secure.gravatar.com/avatar/4021c3aaa995c31bd117cb7800005e85?s=80\\u0026d=identicon\",\"web_url\":\"https://gitlab.com/cqbqdd11519\"}],\"assignee\":{\"id\":7169076,\"name\":\"Sunghyun Kim\",\"username\":\"cqbqdd11519\",\"state\":\"active\",\"avatar_url\":\"https://secure.gravatar.com/avatar/4021c3aaa995c31bd117cb7800005e85?s=80\\u0026d=identicon\",\"web_url\":\"https://gitlab.com/cqbqdd11519\"},\"reviewers\":[],\"source_project_id\":25815215,\"target_project_id\":25815215,\"labels\":[\"kind/test\"],\"work_in_progress\":false,\"milestone\":null,\"merge_when_pipeline_succeeds\":false,\"merge_status\":\"unchecked\",\"sha\":\"e703f64f722f33c4fbb1f326aed08edc81053b0b\",\"merge_commit_sha\":null,\"squash_commit_sha\":null,\"discussion_locked\":null,\"should_remove_source_branch\":null,\"force_remove_source_branch\":false,\"reference\":\"!1\",\"references\":{\"short\":\"!1\",\"relative\":\"!1\",\"full\":\"cqbqdd11519/cicd-test!1\"},\"web_url\":\"https://gitlab.com/cqbqdd11519/cicd-test/-/merge_requests/1\",\"time_stats\":{\"time_estimate\":0,\"total_time_spent\":0,\"human_time_estimate\":null,\"human_total_time_spent\":null},\"squash\":false,\"task_completion_status\":{\"count\":0,\"completed_count\":0},\"has_conflicts\":false,\"blocking_discussions_resolved\":true,\"approvals_before_merge\":null}]"
	sampleMRChange     = `{"id":104830956,"iid":5,"project_id":25815215,"title":"Newnew","state":"opened","created_at":"2021-06-18T07:11:01.715Z","updated_at":"2021-07-13T01:05:33.877Z","target_branch":"master","source_branch":"newnew","source_project_id":25815215,"target_project_id":25815215,"sha":"5f065c6de7dacb91aa5929a5c0ab71ecba5456b0","changes":[{"old_path":"src/main/webapp/index.html","new_path":"src/main/webapp/index.html","a_mode":"100644","b_mode":"100644","new_file":false,"renamed_file":false,"deleted_file":false,"diff":"@@ -1,7 +1,7 @@\n \u003c!DOCTYPE html\u003e\n \u003chtml\u003e\n     \u003chead\u003e\n-        \u003ctitle\u003eTomcatMavenApp\u003c/title\u003e\n+        \u003ctitle\u003eTomcatMavenAppaaaa - add commit3\u003c/title\u003e\n         \u003cmeta http-equiv=\"Content-Type\" content=\"text/html; charset=UTF-8\"\u003e\n     \u003c/head\u003e\n     \u003cbody\u003e\n"}]}`
	sampleMRCommits    = "[\n    {\n        \"id\":\"5f065c6de7dacb91aa5929a5c0ab71ecba5456b0\",\n        \"created_at\":\"2021-04-12T05:07:48.000Z\",\n        \"title\":\"Update index.html\",\n        \"message\":\"Update index.html\",\n        \"author_name\":\"Sunghyun Kim\",\n        \"author_email\":\"cqbqdd11519@gmail.com\",\n        \"authored_date\":\"2021-04-12T05:07:48.000Z\",\n        \"committer_name\":\"Sunghyun Kim\",\n        \"committer_email\":\"cqbqdd11519@gmail.com\",\n        \"committed_date\":\"2021-04-12T05:07:48.000Z\"\n    },\n    {\n        \"id\":\"dace98c2d0437f6ccacd8b9c8094f4dde9162214\",\n        \"created_at\":\"2021-04-12T05:04:54.000Z\",\n        \"title\":\"Update index.html\",\n        \"message\":\"Update index.html\",\n        \"author_name\":\"Sunghyun Kim\",\n        \"author_email\":\"cqbqdd11519@gmail.com\",\n        \"authored_date\":\"2021-04-12T05:04:54.000Z\",\n        \"committer_name\":\"Sunghyun Kim\",\n        \"committer_email\":\"cqbqdd11519@gmail.com\",\n        \"committed_date\":\"2021-04-12T05:04:54.000Z\"\n    },\n    {\n        \"id\":\"e703f64f722f33c4fbb1f326aed08edc81053b0b\",\n        \"created_at\":\"2021-04-12T04:50:34.000Z\",\n        \"title\":\"Update index.html\",\n        \"message\":\"Update index.html\",\n        \"author_name\":\"Sunghyun Kim\",\n        \"author_email\":\"cqbqdd11519@gmail.com\",\n        \"authored_date\":\"2021-04-12T04:50:34.000Z\",\n        \"committer_name\":\"Sunghyun Kim\",\n        \"committer_email\":\"cqbqdd11519@gmail.com\",\n        \"committed_date\":\"2021-04-12T04:50:34.000Z\"\n    },\n    {\n        \"id\":\"3196ccc37bcae94852079b04fcbfaf928341d6e9\",\n        \"created_at\":\"2021-01-22T03:25:50.000Z\",\n        \"title\":\"newnew\",\n        \"message\":\"newnew\\n\",\n        \"author_name\":\"Sunghyun Kim\",\n        \"author_email\":\"cqbqdd11519@gmail.com\",\n        \"authored_date\":\"2021-01-22T03:25:50.000Z\",\n        \"committer_name\":\"Sunghyun Kim\",\n        \"committer_email\":\"cqbqdd11519@gmail.com\",\n        \"committed_date\":\"2021-01-22T03:25:50.000Z\"\n    }\n]"
	sampleMR           = "{\"id\":133148669,\"iid\":1,\"project_id\":31228574,\"title\":\"Child directory test\",\"description\":\"\",\"state\":\"opened\",\"created_at\":\"2021-12-30T06:58:09.077Z\",\"updated_at\":\"2021-12-30T07:18:33.391Z\",\"merged_by\":null,\"merged_at\":null,\"closed_by\":null,\"closed_at\":null,\"target_branch\":\"main\",\"source_branch\":\"child-directory-test\",\"user_notes_count\":1,\"upvotes\":0,\"downvotes\":0,\"author\":{\"id\":10192010,\"username\":\"changjjjjjjj\",\"name\":\"Changju Kim\",\"state\":\"active\",\"avatar_url\":\"https://secure.gravatar.com/avatar/c9995fef2d5a47e133b9461fea8cf3d3?s=80\\u0026d=identicon\",\"web_url\":\"https://gitlab.com/changjjjjjjj\"},\"assignees\":[],\"assignee\":null,\"reviewers\":[],\"source_project_id\":31228574,\"target_project_id\":31228574,\"labels\":[\"approved\"],\"draft\":false,\"work_in_progress\":false,\"milestone\":null,\"merge_when_pipeline_succeeds\":false,\"merge_status\":\"can_be_merged\",\"sha\":\"d84e251bf2d84b74e2e5161bcf693cdbb7130f23\",\"merge_commit_sha\":null,\"squash_commit_sha\":null,\"discussion_locked\":null,\"should_remove_source_branch\":null,\"force_remove_source_branch\":true,\"reference\":\"!1\",\"references\":{\"short\":\"!1\",\"relative\":\"!1\",\"full\":\"changjjjjjjj/cd-example-apps!1\"},\"web_url\":\"https://gitlab.com/changjjjjjjj/cd-example-apps/-/merge_requests/1\",\"time_stats\":{\"time_estimate\":0,\"total_time_spent\":0,\"human_time_estimate\":null,\"human_total_time_spent\":null},\"squash\":false,\"task_completion_status\":{\"count\":0,\"completed_count\":0},\"has_conflicts\":false,\"blocking_discussions_resolved\":true,\"approvals_before_merge\":null,\"subscribed\":true,\"changes_count\":\"2\",\"latest_build_started_at\":null,\"latest_build_finished_at\":null,\"first_deployed_to_production_at\":null,\"pipeline\":null,\"head_pipeline\":null,\"diff_refs\":{\"base_sha\":\"e1eb6f3829eee63f55e77fdf6cf2b332d3a91ae0\",\"head_sha\":\"d84e251bf2d84b74e2e5161bcf693cdbb7130f23\",\"start_sha\":\"c37271972e2bb9fe7ada89e2e7ae7045da4fffcb\"},\"merge_error\":null,\"first_contribution\":false,\"user\":{\"can_merge\":true}}"
	sampleMRNotes      = "[{\"id\":797962489,\"type\":null,\"body\":\"test\",\"attachment\":null,\"author\":{\"id\":10192010,\"username\":\"changjjjjjjj\",\"name\":\"Changju Kim\",\"state\":\"active\",\"avatar_url\":\"https://secure.gravatar.com/avatar/c9995fef2d5a47e133b9461fea8cf3d3?s=80\\u0026d=identicon\",\"web_url\":\"https://gitlab.com/changjjjjjjj\"},\"created_at\":\"2021-12-30T06:58:52.936Z\",\"updated_at\":\"2021-12-30T06:58:52.936Z\",\"system\":false,\"noteable_id\":133148669,\"noteable_type\":\"MergeRequest\",\"resolvable\":false,\"confidential\":false,\"noteable_iid\":1,\"commands_changes\":{}}]"
)

var serverURL string

func TestClient_ListWebhook(t *testing.T) {
	c, err := testEnv()
	if err != nil {
		t.Fatal(err)
	}

	wh, err := c.ListWebhook()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 2, len(wh))
	assert.Equal(t, "http://asdasd/webhook/default/chatops-test-gitlab", wh[0].URL)
	assert.Equal(t, "http://asdasd/webhook/default/chatops-test-gitlab", wh[1].URL)
}

func TestClient_ListCommitStatuses(t *testing.T) {
	c, err := testEnv()
	if err != nil {
		t.Fatal(err)
	}

	sha := "5f065c6de7dacb91aa5929a5c0ab71ecba5456b0"
	statuses, err := c.ListCommitStatuses(sha)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 4, len(statuses))
	assert.Equal(t, "blocker", statuses[0].Context)
	assert.Equal(t, "pending", string(statuses[0].State))
	assert.Equal(t, "test-1", statuses[1].Context)
	assert.Equal(t, "success", string(statuses[1].State))
	assert.Equal(t, "blocker", statuses[2].Context)
	assert.Equal(t, "pending", string(statuses[2].State))
	assert.Equal(t, "test-1", statuses[3].Context)
	assert.Equal(t, "success", string(statuses[3].State))
}

func TestClient_ListLabels(t *testing.T) {
	c, err := testEnv()
	if err != nil {
		t.Fatal(err)
	}

	labels, err := c.ListLabels(5)
	require.NoError(t, err)
	require.Len(t, labels, 1)
	require.Equal(t, "approved", labels[0].Name)
}

func TestClient_ListPullRequests(t *testing.T) {
	c, err := testEnv()
	if err != nil {
		t.Fatal(err)
	}

	prs, err := c.ListPullRequests(false)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 6, len(prs), "Number of prs")
	assert.Equal(t, 3, prs[0].ID, "PR ID")
	assert.Equal(t, "Newnew", prs[0].Title, "PR Title")
	assert.Equal(t, 2, prs[1].ID, "PR ID")
	assert.Equal(t, "Newnew", prs[1].Title, "PR Title")
	assert.Equal(t, 1, prs[2].ID, "PR ID")
	assert.Equal(t, "newnew", prs[2].Title, "PR Title")
	assert.Equal(t, 3, prs[3].ID, "PR ID")
	assert.Equal(t, "Newnew", prs[3].Title, "PR Title")
	assert.Equal(t, 2, prs[4].ID, "PR ID")
	assert.Equal(t, "Newnew", prs[4].Title, "PR Title")
	assert.Equal(t, 1, prs[5].ID, "PR ID")
	assert.Equal(t, "newnew", prs[5].Title, "PR Title")
}

func TestClient_GetPullRequestDiff(t *testing.T) {
	c, err := testEnv()
	if err != nil {
		t.Fatal(err)
	}

	diff, err := c.GetPullRequestDiff(5)
	require.NoError(t, err)
	require.Len(t, diff.Changes, 1)
	require.Equal(t, "src/main/webapp/index.html", diff.Changes[0].Filename)
	require.Equal(t, "src/main/webapp/index.html", diff.Changes[0].OldFilename)
	require.Equal(t, 1, diff.Changes[0].Additions)
	require.Equal(t, 1, diff.Changes[0].Deletions)
	require.Equal(t, 2, diff.Changes[0].Changes)
}

func TestClient_ListComments(t *testing.T) {
	c, err := testEnv()
	if err != nil {
		t.Fatal(err)
	}
	comments, err := c.ListComments(5)
	require.NoError(t, err)
	require.Len(t, comments, 1)
	require.Equal(t, "test", comments[0].Comment.Body)
}

func TestClient_ListPullRequestCommits(t *testing.T) {
	c, err := testEnv()
	if err != nil {
		t.Fatal(err)
	}

	commits, err := c.ListPullRequestCommits(5)
	require.NoError(t, err)
	require.Len(t, commits, 4)
	require.Equal(t, "5f065c6de7dacb91aa5929a5c0ab71ecba5456b0", commits[0].SHA)
	require.Equal(t, "Update index.html", commits[0].Message)
	require.Equal(t, "Sunghyun Kim", commits[0].Author.Name)
	require.Equal(t, "cqbqdd11519@gmail.com", commits[0].Author.Email)
	require.Equal(t, "Sunghyun Kim", commits[0].Committer.Name)
	require.Equal(t, "cqbqdd11519@gmail.com", commits[0].Committer.Email)
}

func testEnv() (*Client, error) {
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(req.URL.String()))
	})
	r.HandleFunc("/api/v4/projects/{org}/{repo}/hooks", func(w http.ResponseWriter, req *http.Request) {
		page := req.URL.Query().Get("page")
		if page == "" || page == "1" {
			w.Header().Set("Link", fmt.Sprintf("<%s/%s?state=all&per_page=100&page=2>; rel=\"next\", <%s/%s?state=all&per_page=100&page=3>; rel=\"last\"", serverURL, req.URL.Path, serverURL, req.URL.Path))
		}
		_, _ = w.Write([]byte(sampleWebhooksList))
	})
	r.HandleFunc("/api/v4/projects/{org}/{repo}/repository/commits/{sha}/statuses", func(w http.ResponseWriter, req *http.Request) {
		page := req.URL.Query().Get("page")
		if page == "" || page == "1" {
			w.Header().Set("Link", fmt.Sprintf("<%s/%s?state=all&per_page=100&page=2>; rel=\"next\", <%s/%s?state=all&per_page=100&page=3>; rel=\"last\"", serverURL, req.URL.Path, serverURL, req.URL.Path))
		}
		_, _ = w.Write([]byte(sampleStatusesList))
	})
	r.HandleFunc("/api/v4/projects/{org}/{repo}/merge_requests", func(w http.ResponseWriter, req *http.Request) {
		page := req.URL.Query().Get("page")
		if page == "" || page == "1" {
			w.Header().Set("Link", fmt.Sprintf("<%s/%s?state=all&per_page=100&page=2>; rel=\"next\", <%s/%s?state=all&per_page=100&page=3>; rel=\"last\"", serverURL, req.URL.Path, serverURL, req.URL.Path))
		}
		_, _ = w.Write([]byte(sampleMRList))
	})
	r.HandleFunc("/api/v4/projects/{org}/{repo}/merge_requests/{iid}/changes", func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte(sampleMRChange))
	})
	r.HandleFunc("/api/v4/projects/{org}/{repo}/merge_requests/{iid}/commits", func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte(sampleMRCommits))
	})
	r.HandleFunc("/api/v4/projects/{org}/{repo}/merge_requests/{iid}", func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte(sampleMR))
	})
	r.HandleFunc("/api/v4/projects/{org}/{repo}/merge_requests/{iid}/notes", func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte(sampleMRNotes))
	})

	testSrv := httptest.NewServer(r)
	serverURL = testSrv.URL

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
				APIUrl:     serverURL,
				Token:      &cicdv1.GitToken{Value: "dummy"},
			},
		},
	}

	c := &Client{
		IntegrationConfig: ic,
		K8sClient:         fake.NewClientBuilder().WithScheme(s).WithObjects(ic).Build(),
	}
	if err := c.Init(); err != nil {
		return nil, err
	}

	return c, nil
}
