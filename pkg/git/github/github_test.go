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

package github

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/bmizerany/assert"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	sampleWebhooksList = "[{\"type\":\"Repository\",\"id\":11111111,\"name\":\"web\",\"active\":true,\"events\":[\"*\"],\"config\":{\"content_type\":\"json\",\"insecure_ssl\":\"0\",\"secret\":\"********\",\"url\":\"http://asdasd/webhook/default/chatops-test\"},\"updated_at\":\"2021-04-08T02:31:42Z\",\"created_at\":\"2021-04-08T02:31:42Z\",\"url\":\"https://api.github.com/repos/vingsu/cicd-test/hooks/11111111\",\"test_url\":\"https://api.github.com/repos/vingsu/cicd-test/hooks/11111111/test\",\"ping_url\":\"https://api.github.com/repos/vingsu/cicd-test/hooks/11111111/pings\",\"last_response\":{\"code\":200,\"status\":\"active\",\"message\":\"OK\"}}]"
	sampleStatusesList = "[{\"id\":1111111111,\"state\":\"success\",\"context\":\"test-1\",\"created_at\":\"2021-04-12T08:37:32Z\",\"updated_at\":\"2021-04-12T08:37:32Z\",\"creator\":{\"login\":\"sunghyunkim3\",\"id\":1111111,\"type\":\"User\",\"site_admin\":false}}]"
	samplePRList       = "[{\"url\":\"https://api.github.com/repos/vingsu/cicd-test/pulls/25\",\"id\":611161419,\"node_id\":\"MDExOlB1bGxSZXF1ZXN0NjExMTYxNDE5\",\"html_url\":\"https://github.com/vingsu/cicd-test/pull/25\",\"number\":25,\"state\":\"open\",\"locked\":false,\"title\":\"newnew\",\"user\":{\"login\":\"cqbqdd11519\",\"id\":6166781,\"node_id\":\"MDQ6VXNlcjYxNjY3ODE=\",\"avatar_url\":\"https://avatars.githubusercontent.com/u/6166781?v=4\",\"gravatar_id\":\"\",\"type\":\"User\",\"site_admin\":false},\"body\":\"\",\"created_at\":\"2021-04-08T02:35:17Z\",\"updated_at\":\"2021-04-13T04:54:16Z\",\"closed_at\":null,\"merged_at\":null,\"merge_commit_sha\":\"b6d9abd3254a6b3da35200f9cdbb307cea7db91a\",\"assignee\":null,\"assignees\":[],\"requested_reviewers\":[{\"login\":\"sunghyunkim3\",\"id\":66240202,\"node_id\":\"MDQ6VXNlcjY2MjQwMjAy\",\"avatar_url\":\"https://avatars.githubusercontent.com/u/66240202?v=4\",\"gravatar_id\":\"\",\"type\":\"User\",\"site_admin\":false}],\"requested_teams\":[],\"labels\":[{\"id\":2905890093,\"node_id\":\"MDU6TGFiZWwyOTA1ODkwMDkz\",\"url\":\"https://api.github.com/repos/vingsu/cicd-test/labels/kind/test\",\"name\":\"kind/test\",\"color\":\"CF61D3\",\"default\":false,\"description\":\"\"}],\"milestone\":null,\"draft\":false,\"head\":{\"label\":\"vingsu:newnew\",\"ref\":\"newnew\",\"sha\":\"3196ccc37bcae94852079b04fcbfaf928341d6e9\",\"user\":{\"login\":\"vingsu\",\"id\":71878727,\"node_id\":\"MDEyOk9yZ2FuaXphdGlvbjcxODc4NzI3\",\"avatar_url\":\"https://avatars.githubusercontent.com/u/71878727?v=4\",\"gravatar_id\":\"\",\"type\":\"Organization\",\"site_admin\":false},\"repo\":{\"id\":319253224,\"node_id\":\"MDEwOlJlcG9zaXRvcnkzMTkyNTMyMjQ=\",\"name\":\"cicd-test\",\"full_name\":\"vingsu/cicd-test\",\"private\":false,\"owner\":{\"login\":\"vingsu\",\"id\":71878727,\"node_id\":\"MDEyOk9yZ2FuaXphdGlvbjcxODc4NzI3\",\"avatar_url\":\"https://avatars.githubusercontent.com/u/71878727?v=4\",\"gravatar_id\":\"\",\"type\":\"Organization\",\"site_admin\":false},\"html_url\":\"https://github.com/vingsu/cicd-test\",\"description\":null,\"fork\":false,\"created_at\":\"2020-12-07T08:31:55Z\",\"updated_at\":\"2021-01-27T04:29:32Z\",\"pushed_at\":\"2021-04-09T04:46:39Z\",\"git_url\":\"git://github.com/vingsu/cicd-test.git\",\"ssh_url\":\"git@github.com:vingsu/cicd-test.git\",\"clone_url\":\"https://github.com/vingsu/cicd-test.git\",\"svn_url\":\"https://github.com/vingsu/cicd-test\",\"homepage\":null,\"size\":10,\"stargazers_count\":0,\"watchers_count\":0,\"language\":\"HTML\",\"has_issues\":true,\"has_projects\":true,\"has_downloads\":true,\"has_wiki\":true,\"has_pages\":false,\"forks_count\":0,\"mirror_url\":null,\"archived\":false,\"disabled\":false,\"open_issues_count\":1,\"license\":null,\"forks\":0,\"open_issues\":1,\"watchers\":0,\"default_branch\":\"master\"}},\"base\":{\"label\":\"vingsu:master\",\"ref\":\"master\",\"sha\":\"22ccae53032027186ba739dfaa473ee61a82b298\",\"user\":{\"login\":\"vingsu\",\"id\":71878727,\"node_id\":\"MDEyOk9yZ2FuaXphdGlvbjcxODc4NzI3\",\"avatar_url\":\"https://avatars.githubusercontent.com/u/71878727?v=4\",\"gravatar_id\":\"\",\"type\":\"Organization\",\"site_admin\":false},\"repo\":{\"id\":319253224,\"node_id\":\"MDEwOlJlcG9zaXRvcnkzMTkyNTMyMjQ=\",\"name\":\"cicd-test\",\"full_name\":\"vingsu/cicd-test\",\"private\":false,\"owner\":{\"login\":\"vingsu\",\"id\":71878727,\"node_id\":\"MDEyOk9yZ2FuaXphdGlvbjcxODc4NzI3\",\"avatar_url\":\"https://avatars.githubusercontent.com/u/71878727?v=4\",\"gravatar_id\":\"\",\"type\":\"Organization\",\"site_admin\":false},\"html_url\":\"https://github.com/vingsu/cicd-test\",\"description\":null,\"fork\":false,\"created_at\":\"2020-12-07T08:31:55Z\",\"updated_at\":\"2021-01-27T04:29:32Z\",\"pushed_at\":\"2021-04-09T04:46:39Z\",\"git_url\":\"git://github.com/vingsu/cicd-test.git\",\"ssh_url\":\"git@github.com:vingsu/cicd-test.git\",\"clone_url\":\"https://github.com/vingsu/cicd-test.git\",\"svn_url\":\"https://github.com/vingsu/cicd-test\",\"homepage\":null,\"size\":10,\"stargazers_count\":0,\"watchers_count\":0,\"language\":\"HTML\",\"has_issues\":true,\"has_projects\":true,\"has_downloads\":true,\"has_wiki\":true,\"has_pages\":false,\"forks_count\":0,\"mirror_url\":null,\"archived\":false,\"disabled\":false,\"open_issues_count\":1,\"license\":null,\"forks\":0,\"open_issues\":1,\"watchers\":0,\"default_branch\":\"master\"}},\"author_association\":\"CONTRIBUTOR\",\"auto_merge\":null,\"active_lock_reason\":null}]"
	samplePRFiles      = "[{\"filename\":\"Makefile\",\"additions\":1,\"deletions\":1,\"changes\":2,\"patch\":\"@@ -1,5 +1,5 @@\\n # Current Operator version\\n-VERSION ?= v0.3.0\\n+VERSION ?= v0.3.1\\n REGISTRY ?= tmaxcloudck\\n \\n # Image URL to use all building/pushing image targets\"},{\"filename\":\"config/release.yaml\",\"additions\":2,\"deletions\":2,\"changes\":4,\"patch\":\"@@ -82,7 +82,7 @@ spec:\\n       containers:\\n       - command:\\n         - /controller\\n-        image: tmaxcloudck/cicd-operator:v0.3.0\\n+        image: tmaxcloudck/cicd-operator:v0.3.1\\n         imagePullPolicy: Always\\n         name: manager\\n         resources:\\n@@ -145,7 +145,7 @@ spec:\\n       containers:\\n         - command:\\n             - /blocker\\n-          image: tmaxcloudck/cicd-blocker:v0.3.0\\n+          image: tmaxcloudck/cicd-blocker:v0.3.1\\n           imagePullPolicy: Always\\n           name: manager\\n           resources:\"},{\"filename\":\"docs/installation.md\",\"additions\":1,\"deletions\":1,\"changes\":2,\"patch\":\"@@ -12,7 +12,7 @@ This guides to install CI/CD operator. The contents are as follows.\\n ## Installing CI/CD Operator\\n 1. Run the following command to install CI/CD operator  \\n    ```bash\\n-   VERSION=v0.3.0\\n+   VERSION=v0.3.1\\n    kubectl apply -f https://raw.githubusercontent.com/tmax-cloud/cicd-operator/$VERSION/config/release.yaml\\n    ```\\n 2. Enable `CustomTask` feature, disable `Affinity Assistant`\"}]"
	samplePRCommits    = "[\n  {\n    \"sha\": \"bfa929712952e60d5ad5d3b73376f6ba392f8b50\",\n    \"commit\": {\n      \"author\": {\n        \"name\": \"Sunghyun Kim\",\n        \"email\": \"cqbqdd11519@gmail.com\",\n        \"date\": \"2021-08-24T07:16:13Z\"\n      },\n      \"committer\": {\n        \"name\": \"Sunghyun Kim\",\n        \"email\": \"cqbqdd11519@gmail.com\",\n        \"date\": \"2021-08-25T04:34:17Z\"\n      },\n      \"message\": \"[fix] Batch pull requests properly\\n\\nfix #270\\n\\n- Fix critical typo\\n- Remove a PR from the batch right away after merging it.\\n  This is to avoid an infinite error, when a PR is already merged, but\\n  is still in the CurrentBatch in the next loop (because of one of the\\n  next PRs fails to merge)\"\n    }\n  }\n]"
)

var serverURL string

func TestClient_CheckRateLimit(t *testing.T) {
	req, _ := http.NewRequest("GET", "", nil)
	testTime := strconv.FormatInt(time.Now().Unix(), 10)

	// GitHub test
	msg := "API rate limit exceeded"
	req.Header.Add("X-RateLimit-Reset", testTime)
	isRateLimit, unixTime := CheckRateLimit(msg, req.Header)
	require.Equal(t, isRateLimit, true)
	require.Equal(t, unixTime, testTime)

	// Non error
	msg = ""
	isRateLimit, unixTime = CheckRateLimit(msg, req.Header)
	require.Equal(t, isRateLimit, false)
	require.Equal(t, unixTime, "")
}

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
	assert.Equal(t, "http://asdasd/webhook/default/chatops-test", wh[0].URL)
	assert.Equal(t, "http://asdasd/webhook/default/chatops-test", wh[1].URL)
}

func TestClient_ListCommitStatuses(t *testing.T) {
	c, err := testEnv()
	if err != nil {
		t.Fatal(err)
	}

	sha := "3196ccc37bcae94852079b04fcbfaf928341d6e9"
	statuses, err := c.ListCommitStatuses(sha)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 1, len(statuses))
	assert.Equal(t, "test-1", statuses[0].Context)
	assert.Equal(t, "success", string(statuses[0].State))
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

	assert.Equal(t, 2, len(prs), "Length")
	assert.Equal(t, 25, prs[0].ID, "ID")
	assert.Equal(t, 25, prs[1].ID, "ID")
	assert.Equal(t, "newnew", prs[0].Title, "Title")
	assert.Equal(t, "newnew", prs[1].Title, "Title")
}

func TestClient_GetPullRequestDiff(t *testing.T) {
	c, err := testEnv()
	if err != nil {
		t.Fatal(err)
	}

	diff, err := c.GetPullRequestDiff(5)
	require.NoError(t, err)
	require.Len(t, diff.Changes, 3)
	require.Equal(t, "Makefile", diff.Changes[0].Filename)
	require.Equal(t, "Makefile", diff.Changes[0].OldFilename)
	require.Equal(t, 1, diff.Changes[0].Additions)
	require.Equal(t, 1, diff.Changes[0].Deletions)
	require.Equal(t, 2, diff.Changes[0].Changes)
	require.Equal(t, "config/release.yaml", diff.Changes[1].Filename)
	require.Equal(t, "config/release.yaml", diff.Changes[1].OldFilename)
	require.Equal(t, 2, diff.Changes[1].Additions)
	require.Equal(t, 2, diff.Changes[1].Deletions)
	require.Equal(t, 4, diff.Changes[1].Changes)
	require.Equal(t, "docs/installation.md", diff.Changes[2].Filename)
	require.Equal(t, "docs/installation.md", diff.Changes[2].OldFilename)
	require.Equal(t, 1, diff.Changes[2].Additions)
	require.Equal(t, 1, diff.Changes[2].Deletions)
	require.Equal(t, 2, diff.Changes[2].Changes)
}

func TestClient_ListPullRequestCommits(t *testing.T) {
	c, err := testEnv()
	if err != nil {
		t.Fatal(err)
	}

	commits, err := c.ListPullRequestCommits(5)
	require.NoError(t, err)
	require.Len(t, commits, 1)
	require.Equal(t, "bfa929712952e60d5ad5d3b73376f6ba392f8b50", commits[0].SHA)
	require.Equal(t, "[fix] Batch pull requests properly\n\nfix #270\n\n- Fix critical typo\n- Remove a PR from the batch right away after merging it.\n  This is to avoid an infinite error, when a PR is already merged, but\n  is still in the CurrentBatch in the next loop (because of one of the\n  next PRs fails to merge)", commits[0].Message)
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
	r.HandleFunc("/repos/{org}/{repo}/hooks", func(w http.ResponseWriter, req *http.Request) {
		page := req.URL.Query().Get("page")
		if page == "" || page == "1" {
			w.Header().Set("Link", fmt.Sprintf("<%s/%s?state=all&per_page=100&page=2>; rel=\"next\", <%s/%s?state=all&per_page=100&page=3>; rel=\"last\"", serverURL, req.URL.Path, serverURL, req.URL.Path))
		}
		_, _ = w.Write([]byte(sampleWebhooksList))
	})
	r.HandleFunc("/repos/{org}/{repo}/commits/{sha}/statuses", func(w http.ResponseWriter, req *http.Request) {
		page := req.URL.Query().Get("page")
		if page == "" || page == "1" {
			w.Header().Set("Link", fmt.Sprintf("<%s/%s?state=all&per_page=100&page=2>; rel=\"next\", <%s/%s?state=all&per_page=100&page=3>; rel=\"last\"", serverURL, req.URL.Path, serverURL, req.URL.Path))
		}
		_, _ = w.Write([]byte(sampleStatusesList))
	})
	r.HandleFunc("/repos/{org}/{repo}/pulls", func(w http.ResponseWriter, req *http.Request) {
		page := req.URL.Query().Get("page")
		if page == "" || page == "1" {
			w.Header().Set("Link", fmt.Sprintf("<%s/%s?state=all&per_page=100&page=2>; rel=\"next\", <%s/%s?state=all&per_page=100&page=3>; rel=\"last\"", serverURL, req.URL.Path, serverURL, req.URL.Path))
		}
		_, _ = w.Write([]byte(samplePRList))
	})
	r.HandleFunc("/repos/{org}/{repo}/pulls/{id}/files", func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte(samplePRFiles))
	})
	r.HandleFunc("/repos/{org}/{repo}/pulls/{id}/commits", func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte(samplePRCommits))
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
				Type:       "github",
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
