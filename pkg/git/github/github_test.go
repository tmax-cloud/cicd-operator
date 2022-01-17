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
	sampleWebhooksList  = "[{\"type\":\"Repository\",\"id\":11111111,\"name\":\"web\",\"active\":true,\"events\":[\"*\"],\"config\":{\"content_type\":\"json\",\"insecure_ssl\":\"0\",\"secret\":\"********\",\"url\":\"http://asdasd/webhook/default/chatops-test\"},\"updated_at\":\"2021-04-08T02:31:42Z\",\"created_at\":\"2021-04-08T02:31:42Z\",\"url\":\"https://api.github.com/repos/vingsu/cicd-test/hooks/11111111\",\"test_url\":\"https://api.github.com/repos/vingsu/cicd-test/hooks/11111111/test\",\"ping_url\":\"https://api.github.com/repos/vingsu/cicd-test/hooks/11111111/pings\",\"last_response\":{\"code\":200,\"status\":\"active\",\"message\":\"OK\"}}]"
	sampleStatusesList  = "[{\"id\":1111111111,\"state\":\"success\",\"context\":\"test-1\",\"created_at\":\"2021-04-12T08:37:32Z\",\"updated_at\":\"2021-04-12T08:37:32Z\",\"creator\":{\"login\":\"sunghyunkim3\",\"id\":1111111,\"type\":\"User\",\"site_admin\":false}}]"
	samplePRList        = "[{\"url\":\"https://api.github.com/repos/vingsu/cicd-test/pulls/25\",\"id\":611161419,\"node_id\":\"MDExOlB1bGxSZXF1ZXN0NjExMTYxNDE5\",\"html_url\":\"https://github.com/vingsu/cicd-test/pull/25\",\"number\":25,\"state\":\"open\",\"locked\":false,\"title\":\"newnew\",\"user\":{\"login\":\"cqbqdd11519\",\"id\":6166781,\"node_id\":\"MDQ6VXNlcjYxNjY3ODE=\",\"avatar_url\":\"https://avatars.githubusercontent.com/u/6166781?v=4\",\"gravatar_id\":\"\",\"type\":\"User\",\"site_admin\":false},\"body\":\"\",\"created_at\":\"2021-04-08T02:35:17Z\",\"updated_at\":\"2021-04-13T04:54:16Z\",\"closed_at\":null,\"merged_at\":null,\"merge_commit_sha\":\"b6d9abd3254a6b3da35200f9cdbb307cea7db91a\",\"assignee\":null,\"assignees\":[],\"requested_reviewers\":[{\"login\":\"sunghyunkim3\",\"id\":66240202,\"node_id\":\"MDQ6VXNlcjY2MjQwMjAy\",\"avatar_url\":\"https://avatars.githubusercontent.com/u/66240202?v=4\",\"gravatar_id\":\"\",\"type\":\"User\",\"site_admin\":false}],\"requested_teams\":[],\"labels\":[{\"id\":2905890093,\"node_id\":\"MDU6TGFiZWwyOTA1ODkwMDkz\",\"url\":\"https://api.github.com/repos/vingsu/cicd-test/labels/kind/test\",\"name\":\"kind/test\",\"color\":\"CF61D3\",\"default\":false,\"description\":\"\"}],\"milestone\":null,\"draft\":false,\"head\":{\"label\":\"vingsu:newnew\",\"ref\":\"newnew\",\"sha\":\"3196ccc37bcae94852079b04fcbfaf928341d6e9\",\"user\":{\"login\":\"vingsu\",\"id\":71878727,\"node_id\":\"MDEyOk9yZ2FuaXphdGlvbjcxODc4NzI3\",\"avatar_url\":\"https://avatars.githubusercontent.com/u/71878727?v=4\",\"gravatar_id\":\"\",\"type\":\"Organization\",\"site_admin\":false},\"repo\":{\"id\":319253224,\"node_id\":\"MDEwOlJlcG9zaXRvcnkzMTkyNTMyMjQ=\",\"name\":\"cicd-test\",\"full_name\":\"vingsu/cicd-test\",\"private\":false,\"owner\":{\"login\":\"vingsu\",\"id\":71878727,\"node_id\":\"MDEyOk9yZ2FuaXphdGlvbjcxODc4NzI3\",\"avatar_url\":\"https://avatars.githubusercontent.com/u/71878727?v=4\",\"gravatar_id\":\"\",\"type\":\"Organization\",\"site_admin\":false},\"html_url\":\"https://github.com/vingsu/cicd-test\",\"description\":null,\"fork\":false,\"created_at\":\"2020-12-07T08:31:55Z\",\"updated_at\":\"2021-01-27T04:29:32Z\",\"pushed_at\":\"2021-04-09T04:46:39Z\",\"git_url\":\"git://github.com/vingsu/cicd-test.git\",\"ssh_url\":\"git@github.com:vingsu/cicd-test.git\",\"clone_url\":\"https://github.com/vingsu/cicd-test.git\",\"svn_url\":\"https://github.com/vingsu/cicd-test\",\"homepage\":null,\"size\":10,\"stargazers_count\":0,\"watchers_count\":0,\"language\":\"HTML\",\"has_issues\":true,\"has_projects\":true,\"has_downloads\":true,\"has_wiki\":true,\"has_pages\":false,\"forks_count\":0,\"mirror_url\":null,\"archived\":false,\"disabled\":false,\"open_issues_count\":1,\"license\":null,\"forks\":0,\"open_issues\":1,\"watchers\":0,\"default_branch\":\"master\"}},\"base\":{\"label\":\"vingsu:master\",\"ref\":\"master\",\"sha\":\"22ccae53032027186ba739dfaa473ee61a82b298\",\"user\":{\"login\":\"vingsu\",\"id\":71878727,\"node_id\":\"MDEyOk9yZ2FuaXphdGlvbjcxODc4NzI3\",\"avatar_url\":\"https://avatars.githubusercontent.com/u/71878727?v=4\",\"gravatar_id\":\"\",\"type\":\"Organization\",\"site_admin\":false},\"repo\":{\"id\":319253224,\"node_id\":\"MDEwOlJlcG9zaXRvcnkzMTkyNTMyMjQ=\",\"name\":\"cicd-test\",\"full_name\":\"vingsu/cicd-test\",\"private\":false,\"owner\":{\"login\":\"vingsu\",\"id\":71878727,\"node_id\":\"MDEyOk9yZ2FuaXphdGlvbjcxODc4NzI3\",\"avatar_url\":\"https://avatars.githubusercontent.com/u/71878727?v=4\",\"gravatar_id\":\"\",\"type\":\"Organization\",\"site_admin\":false},\"html_url\":\"https://github.com/vingsu/cicd-test\",\"description\":null,\"fork\":false,\"created_at\":\"2020-12-07T08:31:55Z\",\"updated_at\":\"2021-01-27T04:29:32Z\",\"pushed_at\":\"2021-04-09T04:46:39Z\",\"git_url\":\"git://github.com/vingsu/cicd-test.git\",\"ssh_url\":\"git@github.com:vingsu/cicd-test.git\",\"clone_url\":\"https://github.com/vingsu/cicd-test.git\",\"svn_url\":\"https://github.com/vingsu/cicd-test\",\"homepage\":null,\"size\":10,\"stargazers_count\":0,\"watchers_count\":0,\"language\":\"HTML\",\"has_issues\":true,\"has_projects\":true,\"has_downloads\":true,\"has_wiki\":true,\"has_pages\":false,\"forks_count\":0,\"mirror_url\":null,\"archived\":false,\"disabled\":false,\"open_issues_count\":1,\"license\":null,\"forks\":0,\"open_issues\":1,\"watchers\":0,\"default_branch\":\"master\"}},\"author_association\":\"CONTRIBUTOR\",\"auto_merge\":null,\"active_lock_reason\":null}]"
	samplePRFiles       = "[{\"filename\":\"Makefile\",\"additions\":1,\"deletions\":1,\"changes\":2,\"patch\":\"@@ -1,5 +1,5 @@\\n # Current Operator version\\n-VERSION ?= v0.3.0\\n+VERSION ?= v0.3.1\\n REGISTRY ?= tmaxcloudck\\n \\n # Image URL to use all building/pushing image targets\"},{\"filename\":\"config/release.yaml\",\"additions\":2,\"deletions\":2,\"changes\":4,\"patch\":\"@@ -82,7 +82,7 @@ spec:\\n       containers:\\n       - command:\\n         - /controller\\n-        image: tmaxcloudck/cicd-operator:v0.3.0\\n+        image: tmaxcloudck/cicd-operator:v0.3.1\\n         imagePullPolicy: Always\\n         name: manager\\n         resources:\\n@@ -145,7 +145,7 @@ spec:\\n       containers:\\n         - command:\\n             - /blocker\\n-          image: tmaxcloudck/cicd-blocker:v0.3.0\\n+          image: tmaxcloudck/cicd-blocker:v0.3.1\\n           imagePullPolicy: Always\\n           name: manager\\n           resources:\"},{\"filename\":\"docs/installation.md\",\"additions\":1,\"deletions\":1,\"changes\":2,\"patch\":\"@@ -12,7 +12,7 @@ This guides to install CI/CD operator. The contents are as follows.\\n ## Installing CI/CD Operator\\n 1. Run the following command to install CI/CD operator  \\n    ```bash\\n-   VERSION=v0.3.0\\n+   VERSION=v0.3.1\\n    kubectl apply -f https://raw.githubusercontent.com/tmax-cloud/cicd-operator/$VERSION/config/release.yaml\\n    ```\\n 2. Enable `CustomTask` feature, disable `Affinity Assistant`\"}]"
	samplePRCommits     = "[\n  {\n    \"sha\": \"bfa929712952e60d5ad5d3b73376f6ba392f8b50\",\n    \"commit\": {\n      \"author\": {\n        \"name\": \"Sunghyun Kim\",\n        \"email\": \"cqbqdd11519@gmail.com\",\n        \"date\": \"2021-08-24T07:16:13Z\"\n      },\n      \"committer\": {\n        \"name\": \"Sunghyun Kim\",\n        \"email\": \"cqbqdd11519@gmail.com\",\n        \"date\": \"2021-08-25T04:34:17Z\"\n      },\n      \"message\": \"[fix] Batch pull requests properly\\n\\nfix #270\\n\\n- Fix critical typo\\n- Remove a PR from the batch right away after merging it.\\n  This is to avoid an infinite error, when a PR is already merged, but\\n  is still in the CurrentBatch in the next loop (because of one of the\\n  next PRs fails to merge)\"\n    }\n  }\n]"
	sampleLabelLists    = "[\n  {\n    \"id\": 3048006488,\n    \"node_id\": \"MDU6TGFiZWwzMDQ4MDA2NDg4\",\n    \"url\": \"https://api.github.com/repos/tmax-cloud/cicd-operator/labels/approved\",\n    \"name\": \"approved\",\n    \"color\": \"ededed\",\n    \"default\": false,\n    \"description\": null\n  },\n  {\n    \"id\": 3187077209,\n    \"node_id\": \"MDU6TGFiZWwzMTg3MDc3MjA5\",\n    \"url\": \"https://api.github.com/repos/tmax-cloud/cicd-operator/labels/size/L\",\n    \"name\": \"size/L\",\n    \"color\": \"ededed\",\n    \"default\": false,\n    \"description\": null\n  }\n]"
	samplePRComments    = "[\n  {\n    \"url\": \"https://api.github.com/repos/tmax-cloud/cicd-operator/pulls/comments/771113606\",\n    \"pull_request_review_id\": 834849190,\n    \"id\": 771113606,\n    \"node_id\": \"PRRC_kwDOEm6Tx84t9kKG\",\n    \"diff_hunk\": \"@@ -20,89 +20,10 @@ import (\\n \\t\\\"testing\\\"\\n \\n \\t\\\"github.com/stretchr/testify/require\\\"\\n-\\ttektonv1beta1 \\\"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1\\\"\\n \\t\\\"github.com/tmax-cloud/cicd-operator/internal/configs\\\"\\n \\tmetav1 \\\"k8s.io/apimachinery/pkg/apis/meta/v1\\\"\\n )\\n \\n-func TestConvertToTektonParamSpecs(t *testing.T) {\",\n    \"path\": \"api/v1/integrationjob_types_test.go\",\n    \"position\": 9,\n    \"original_position\": 9,\n    \"commit_id\": \"d3b2006b7a2ab28268b248429bc215854a497d24\",\n    \"original_commit_id\": \"654761e79f45e62ef8ca4d94c47cf7adc1756122\",\n    \"user\": {\n      \"login\": \"eddy-kor-92\",\n      \"id\": 33279734,\n      \"node_id\": \"MDQ6VXNlcjMzMjc5NzM0\",\n      \"avatar_url\": \"https://avatars.githubusercontent.com/u/33279734?v=4\",\n      \"gravatar_id\": \"\",\n      \"url\": \"https://api.github.com/users/eddy-kor-92\",\n      \"html_url\": \"https://github.com/eddy-kor-92\",\n      \"followers_url\": \"https://api.github.com/users/eddy-kor-92/followers\",\n      \"following_url\": \"https://api.github.com/users/eddy-kor-92/following{/other_user}\",\n      \"gists_url\": \"https://api.github.com/users/eddy-kor-92/gists{/gist_id}\",\n      \"starred_url\": \"https://api.github.com/users/eddy-kor-92/starred{/owner}{/repo}\",\n      \"subscriptions_url\": \"https://api.github.com/users/eddy-kor-92/subscriptions\",\n      \"organizations_url\": \"https://api.github.com/users/eddy-kor-92/orgs\",\n      \"repos_url\": \"https://api.github.com/users/eddy-kor-92/repos\",\n      \"events_url\": \"https://api.github.com/users/eddy-kor-92/events{/privacy}\",\n      \"received_events_url\": \"https://api.github.com/users/eddy-kor-92/received_events\",\n      \"type\": \"User\",\n      \"site_admin\": false\n    },\n    \"body\": \"이 Test 함수가 원래 integrationconfig_types_test에 있는게 맞는거죠? 그래서 옮기신거죠?\",\n    \"created_at\": \"2021-12-17T05:29:08Z\",\n    \"updated_at\": \"2021-12-17T05:31:38Z\",\n    \"html_url\": \"https://github.com/tmax-cloud/cicd-operator/pull/324#discussion_r771113606\",\n    \"pull_request_url\": \"https://api.github.com/repos/tmax-cloud/cicd-operator/pulls/324\",\n    \"author_association\": \"NONE\",\n    \"_links\": {\n      \"self\": {\n        \"href\": \"https://api.github.com/repos/tmax-cloud/cicd-operator/pulls/comments/771113606\"\n      },\n      \"html\": {\n        \"href\": \"https://github.com/tmax-cloud/cicd-operator/pull/324#discussion_r771113606\"\n      },\n      \"pull_request\": {\n        \"href\": \"https://api.github.com/repos/tmax-cloud/cicd-operator/pulls/324\"\n      }\n    },\n    \"reactions\": {\n      \"url\": \"https://api.github.com/repos/tmax-cloud/cicd-operator/pulls/comments/771113606/reactions\",\n      \"total_count\": 0,\n      \"+1\": 0,\n      \"-1\": 0,\n      \"laugh\": 0,\n      \"hooray\": 0,\n      \"confused\": 0,\n      \"heart\": 0,\n      \"rocket\": 0,\n      \"eyes\": 0\n    },\n    \"start_line\": null,\n    \"original_start_line\": null,\n    \"start_side\": null,\n    \"line\": 28,\n    \"original_line\": 28,\n    \"side\": \"LEFT\"\n  },\n  {\n    \"url\": \"https://api.github.com/repos/tmax-cloud/cicd-operator/pulls/comments/771114018\",\n    \"pull_request_review_id\": 834849190,\n    \"id\": 771114018,\n    \"node_id\": \"PRRC_kwDOEm6Tx84t9kQi\",\n    \"diff_hunk\": \"@@ -127,18 +130,33 @@ func (p *pipelineManager) Generate(job *cicdv1.IntegrationJob) (*tektonv1beta1.P\\n \\t\\t\\t\\tResources:  specResources,\\n \\t\\t\\t\\tTasks:      tasks,\\n \\t\\t\\t\\tWorkspaces: workspaceDefs,\\n-\\t\\t\\t\\tParams:     cicdv1.ConvertToTektonParamSpecs(job.Spec.ParamConfig.ParamDefine),\\n+\\t\\t\\t\\tParams:     paramDefine,\\n \\t\\t\\t},\\n \\t\\t\\tPodTemplate: job.Spec.PodTemplate,\\n \\t\\t\\tWorkspaces:  job.Spec.Workspaces,\\n \\t\\t\\tTimeout: &metav1.Duration{\\n \\t\\t\\t\\tDuration: job.Spec.Timeout.Duration,\\n \\t\\t\\t},\\n-\\t\\t\\tParams: cicdv1.ConvertToTektonParams(job.Spec.ParamConfig.ParamValue),\\n+\\t\\t\\tParams: paramValue,\\n \\t\\t},\\n \\t}, nil\\n }\\n \\n+func getParams(job *cicdv1.IntegrationJob) ([]tektonv1beta1.ParamSpec, []tektonv1beta1.Param) {\",\n    \"path\": \"pkg/pipelinemanager/pipelinemanager.go\",\n    \"position\": 28,\n    \"original_position\": 28,\n    \"commit_id\": \"d3b2006b7a2ab28268b248429bc215854a497d24\",\n    \"original_commit_id\": \"654761e79f45e62ef8ca4d94c47cf7adc1756122\",\n    \"user\": {\n      \"login\": \"eddy-kor-92\",\n      \"id\": 33279734,\n      \"node_id\": \"MDQ6VXNlcjMzMjc5NzM0\",\n      \"avatar_url\": \"https://avatars.githubusercontent.com/u/33279734?v=4\",\n      \"gravatar_id\": \"\",\n      \"url\": \"https://api.github.com/users/eddy-kor-92\",\n      \"html_url\": \"https://github.com/eddy-kor-92\",\n      \"followers_url\": \"https://api.github.com/users/eddy-kor-92/followers\",\n      \"following_url\": \"https://api.github.com/users/eddy-kor-92/following{/other_user}\",\n      \"gists_url\": \"https://api.github.com/users/eddy-kor-92/gists{/gist_id}\",\n      \"starred_url\": \"https://api.github.com/users/eddy-kor-92/starred{/owner}{/repo}\",\n      \"subscriptions_url\": \"https://api.github.com/users/eddy-kor-92/subscriptions\",\n      \"organizations_url\": \"https://api.github.com/users/eddy-kor-92/orgs\",\n      \"repos_url\": \"https://api.github.com/users/eddy-kor-92/repos\",\n      \"events_url\": \"https://api.github.com/users/eddy-kor-92/events{/privacy}\",\n      \"received_events_url\": \"https://api.github.com/users/eddy-kor-92/received_events\",\n      \"type\": \"User\",\n      \"site_admin\": false\n    },\n    \"body\": \"nil 체크를 하는게 이 함수의 목적인거 같은데, parameter를 직접 사용하는 함수에서 parameter validation을 하는게 더 낫지 않을까요? ConvertToTektonParamSpecs랑 ConvertToTektonParams 함수에서요.\",\n    \"created_at\": \"2021-12-17T05:30:31Z\",\n    \"updated_at\": \"2021-12-17T05:31:38Z\",\n    \"html_url\": \"https://github.com/tmax-cloud/cicd-operator/pull/324#discussion_r771114018\",\n    \"pull_request_url\": \"https://api.github.com/repos/tmax-cloud/cicd-operator/pulls/324\",\n    \"author_association\": \"NONE\",\n    \"_links\": {\n      \"self\": {\n        \"href\": \"https://api.github.com/repos/tmax-cloud/cicd-operator/pulls/comments/771114018\"\n      },\n      \"html\": {\n        \"href\": \"https://github.com/tmax-cloud/cicd-operator/pull/324#discussion_r771114018\"\n      },\n      \"pull_request\": {\n        \"href\": \"https://api.github.com/repos/tmax-cloud/cicd-operator/pulls/324\"\n      }\n    },\n    \"reactions\": {\n      \"url\": \"https://api.github.com/repos/tmax-cloud/cicd-operator/pulls/comments/771114018/reactions\",\n      \"total_count\": 0,\n      \"+1\": 0,\n      \"-1\": 0,\n      \"laugh\": 0,\n      \"hooray\": 0,\n      \"confused\": 0,\n      \"heart\": 0,\n      \"rocket\": 0,\n      \"eyes\": 0\n    },\n    \"start_line\": null,\n    \"original_start_line\": null,\n    \"start_side\": null,\n    \"line\": 145,\n    \"original_line\": 145,\n    \"side\": \"RIGHT\"\n  },\n  {\n    \"url\": \"https://api.github.com/repos/tmax-cloud/cicd-operator/pulls/comments/771115644\",\n    \"pull_request_review_id\": 834851875,\n    \"id\": 771115644,\n    \"node_id\": \"PRRC_kwDOEm6Tx84t9kp8\",\n    \"diff_hunk\": \"@@ -20,89 +20,10 @@ import (\\n \\t\\\"testing\\\"\\n \\n \\t\\\"github.com/stretchr/testify/require\\\"\\n-\\ttektonv1beta1 \\\"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1\\\"\\n \\t\\\"github.com/tmax-cloud/cicd-operator/internal/configs\\\"\\n \\tmetav1 \\\"k8s.io/apimachinery/pkg/apis/meta/v1\\\"\\n )\\n \\n-func TestConvertToTektonParamSpecs(t *testing.T) {\",\n    \"path\": \"api/v1/integrationjob_types_test.go\",\n    \"position\": 9,\n    \"original_position\": 9,\n    \"commit_id\": \"d3b2006b7a2ab28268b248429bc215854a497d24\",\n    \"original_commit_id\": \"654761e79f45e62ef8ca4d94c47cf7adc1756122\",\n    \"user\": {\n      \"login\": \"changjjjjjjj\",\n      \"id\": 56624551,\n      \"node_id\": \"MDQ6VXNlcjU2NjI0NTUx\",\n      \"avatar_url\": \"https://avatars.githubusercontent.com/u/56624551?v=4\",\n      \"gravatar_id\": \"\",\n      \"url\": \"https://api.github.com/users/changjjjjjjj\",\n      \"html_url\": \"https://github.com/changjjjjjjj\",\n      \"followers_url\": \"https://api.github.com/users/changjjjjjjj/followers\",\n      \"following_url\": \"https://api.github.com/users/changjjjjjjj/following{/other_user}\",\n      \"gists_url\": \"https://api.github.com/users/changjjjjjjj/gists{/gist_id}\",\n      \"starred_url\": \"https://api.github.com/users/changjjjjjjj/starred{/owner}{/repo}\",\n      \"subscriptions_url\": \"https://api.github.com/users/changjjjjjjj/subscriptions\",\n      \"organizations_url\": \"https://api.github.com/users/changjjjjjjj/orgs\",\n      \"repos_url\": \"https://api.github.com/users/changjjjjjjj/repos\",\n      \"events_url\": \"https://api.github.com/users/changjjjjjjj/events{/privacy}\",\n      \"received_events_url\": \"https://api.github.com/users/changjjjjjjj/received_events\",\n      \"type\": \"User\",\n      \"site_admin\": false\n    },\n    \"body\": \"네 잘못 들어가있어서 옮겼습니다\",\n    \"created_at\": \"2021-12-17T05:36:07Z\",\n    \"updated_at\": \"2021-12-17T05:36:07Z\",\n    \"html_url\": \"https://github.com/tmax-cloud/cicd-operator/pull/324#discussion_r771115644\",\n    \"pull_request_url\": \"https://api.github.com/repos/tmax-cloud/cicd-operator/pulls/324\",\n    \"author_association\": \"COLLABORATOR\",\n    \"_links\": {\n      \"self\": {\n        \"href\": \"https://api.github.com/repos/tmax-cloud/cicd-operator/pulls/comments/771115644\"\n      },\n      \"html\": {\n        \"href\": \"https://github.com/tmax-cloud/cicd-operator/pull/324#discussion_r771115644\"\n      },\n      \"pull_request\": {\n        \"href\": \"https://api.github.com/repos/tmax-cloud/cicd-operator/pulls/324\"\n      }\n    },\n    \"reactions\": {\n      \"url\": \"https://api.github.com/repos/tmax-cloud/cicd-operator/pulls/comments/771115644/reactions\",\n      \"total_count\": 0,\n      \"+1\": 0,\n      \"-1\": 0,\n      \"laugh\": 0,\n      \"hooray\": 0,\n      \"confused\": 0,\n      \"heart\": 0,\n      \"rocket\": 0,\n      \"eyes\": 0\n    },\n    \"start_line\": null,\n    \"original_start_line\": null,\n    \"start_side\": null,\n    \"line\": 28,\n    \"original_line\": 28,\n    \"side\": \"LEFT\",\n    \"in_reply_to_id\": 771113606\n  },\n  {\n    \"url\": \"https://api.github.com/repos/tmax-cloud/cicd-operator/pulls/comments/771122149\",\n    \"pull_request_review_id\": 834860063,\n    \"id\": 771122149,\n    \"node_id\": \"PRRC_kwDOEm6Tx84t9mPl\",\n    \"diff_hunk\": \"@@ -127,18 +130,33 @@ func (p *pipelineManager) Generate(job *cicdv1.IntegrationJob) (*tektonv1beta1.P\\n \\t\\t\\t\\tResources:  specResources,\\n \\t\\t\\t\\tTasks:      tasks,\\n \\t\\t\\t\\tWorkspaces: workspaceDefs,\\n-\\t\\t\\t\\tParams:     cicdv1.ConvertToTektonParamSpecs(job.Spec.ParamConfig.ParamDefine),\\n+\\t\\t\\t\\tParams:     paramDefine,\\n \\t\\t\\t},\\n \\t\\t\\tPodTemplate: job.Spec.PodTemplate,\\n \\t\\t\\tWorkspaces:  job.Spec.Workspaces,\\n \\t\\t\\tTimeout: &metav1.Duration{\\n \\t\\t\\t\\tDuration: job.Spec.Timeout.Duration,\\n \\t\\t\\t},\\n-\\t\\t\\tParams: cicdv1.ConvertToTektonParams(job.Spec.ParamConfig.ParamValue),\\n+\\t\\t\\tParams: paramValue,\\n \\t\\t},\\n \\t}, nil\\n }\\n \\n+func getParams(job *cicdv1.IntegrationJob) ([]tektonv1beta1.ParamSpec, []tektonv1beta1.Param) {\",\n    \"path\": \"pkg/pipelinemanager/pipelinemanager.go\",\n    \"position\": 28,\n    \"original_position\": 28,\n    \"commit_id\": \"d3b2006b7a2ab28268b248429bc215854a497d24\",\n    \"original_commit_id\": \"654761e79f45e62ef8ca4d94c47cf7adc1756122\",\n    \"user\": {\n      \"login\": \"changjjjjjjj\",\n      \"id\": 56624551,\n      \"node_id\": \"MDQ6VXNlcjU2NjI0NTUx\",\n      \"avatar_url\": \"https://avatars.githubusercontent.com/u/56624551?v=4\",\n      \"gravatar_id\": \"\",\n      \"url\": \"https://api.github.com/users/changjjjjjjj\",\n      \"html_url\": \"https://github.com/changjjjjjjj\",\n      \"followers_url\": \"https://api.github.com/users/changjjjjjjj/followers\",\n      \"following_url\": \"https://api.github.com/users/changjjjjjjj/following{/other_user}\",\n      \"gists_url\": \"https://api.github.com/users/changjjjjjjj/gists{/gist_id}\",\n      \"starred_url\": \"https://api.github.com/users/changjjjjjjj/starred{/owner}{/repo}\",\n      \"subscriptions_url\": \"https://api.github.com/users/changjjjjjjj/subscriptions\",\n      \"organizations_url\": \"https://api.github.com/users/changjjjjjjj/orgs\",\n      \"repos_url\": \"https://api.github.com/users/changjjjjjjj/repos\",\n      \"events_url\": \"https://api.github.com/users/changjjjjjjj/events{/privacy}\",\n      \"received_events_url\": \"https://api.github.com/users/changjjjjjjj/received_events\",\n      \"type\": \"User\",\n      \"site_admin\": false\n    },\n    \"body\": \"paramConfig nil 은 체크해야 해서 함수는 남겨뒀고 생각해보니까 paramDefine이랑 paramValue는  getParams에서 nil 체크 안해도 돼서 삭제했습니다.\",\n    \"created_at\": \"2021-12-17T05:57:08Z\",\n    \"updated_at\": \"2021-12-17T05:57:08Z\",\n    \"html_url\": \"https://github.com/tmax-cloud/cicd-operator/pull/324#discussion_r771122149\",\n    \"pull_request_url\": \"https://api.github.com/repos/tmax-cloud/cicd-operator/pulls/324\",\n    \"author_association\": \"COLLABORATOR\",\n    \"_links\": {\n      \"self\": {\n        \"href\": \"https://api.github.com/repos/tmax-cloud/cicd-operator/pulls/comments/771122149\"\n      },\n      \"html\": {\n        \"href\": \"https://github.com/tmax-cloud/cicd-operator/pull/324#discussion_r771122149\"\n      },\n      \"pull_request\": {\n        \"href\": \"https://api.github.com/repos/tmax-cloud/cicd-operator/pulls/324\"\n      }\n    },\n    \"reactions\": {\n      \"url\": \"https://api.github.com/repos/tmax-cloud/cicd-operator/pulls/comments/771122149/reactions\",\n      \"total_count\": 0,\n      \"+1\": 0,\n      \"-1\": 0,\n      \"laugh\": 0,\n      \"hooray\": 0,\n      \"confused\": 0,\n      \"heart\": 0,\n      \"rocket\": 0,\n      \"eyes\": 0\n    },\n    \"start_line\": null,\n    \"original_start_line\": null,\n    \"start_side\": null,\n    \"line\": 145,\n    \"original_line\": 145,\n    \"side\": \"RIGHT\",\n    \"in_reply_to_id\": 771114018\n  }\n]"
	samplePRReviews     = "[\n  {\n    \"id\": 834849190,\n    \"node_id\": \"PRR_kwDOEm6Tx84xwsmm\",\n    \"user\": {\n      \"login\": \"eddy-kor-92\",\n      \"id\": 33279734,\n      \"node_id\": \"MDQ6VXNlcjMzMjc5NzM0\",\n      \"avatar_url\": \"https://avatars.githubusercontent.com/u/33279734?u=bed3bf0df30f21a34b1d88dac4bdea053d2edafa&v=4\",\n      \"gravatar_id\": \"\",\n      \"url\": \"https://api.github.com/users/eddy-kor-92\",\n      \"html_url\": \"https://github.com/eddy-kor-92\",\n      \"followers_url\": \"https://api.github.com/users/eddy-kor-92/followers\",\n      \"following_url\": \"https://api.github.com/users/eddy-kor-92/following{/other_user}\",\n      \"gists_url\": \"https://api.github.com/users/eddy-kor-92/gists{/gist_id}\",\n      \"starred_url\": \"https://api.github.com/users/eddy-kor-92/starred{/owner}{/repo}\",\n      \"subscriptions_url\": \"https://api.github.com/users/eddy-kor-92/subscriptions\",\n      \"organizations_url\": \"https://api.github.com/users/eddy-kor-92/orgs\",\n      \"repos_url\": \"https://api.github.com/users/eddy-kor-92/repos\",\n      \"events_url\": \"https://api.github.com/users/eddy-kor-92/events{/privacy}\",\n      \"received_events_url\": \"https://api.github.com/users/eddy-kor-92/received_events\",\n      \"type\": \"User\",\n      \"site_admin\": false\n    },\n    \"body\": \"\",\n    \"state\": \"COMMENTED\",\n    \"html_url\": \"https://github.com/tmax-cloud/cicd-operator/pull/324#pullrequestreview-834849190\",\n    \"pull_request_url\": \"https://api.github.com/repos/tmax-cloud/cicd-operator/pulls/324\",\n    \"author_association\": \"NONE\",\n    \"_links\": {\n      \"html\": {\n        \"href\": \"https://github.com/tmax-cloud/cicd-operator/pull/324#pullrequestreview-834849190\"\n      },\n      \"pull_request\": {\n        \"href\": \"https://api.github.com/repos/tmax-cloud/cicd-operator/pulls/324\"\n      }\n    },\n    \"submitted_at\": \"2021-12-17T05:31:38Z\",\n    \"commit_id\": \"654761e79f45e62ef8ca4d94c47cf7adc1756122\"\n  },\n  {\n    \"id\": 834851875,\n    \"node_id\": \"PRR_kwDOEm6Tx84xwtQj\",\n    \"user\": {\n      \"login\": \"changjjjjjjj\",\n      \"id\": 56624551,\n      \"node_id\": \"MDQ6VXNlcjU2NjI0NTUx\",\n      \"avatar_url\": \"https://avatars.githubusercontent.com/u/56624551?v=4\",\n      \"gravatar_id\": \"\",\n      \"url\": \"https://api.github.com/users/changjjjjjjj\",\n      \"html_url\": \"https://github.com/changjjjjjjj\",\n      \"followers_url\": \"https://api.github.com/users/changjjjjjjj/followers\",\n      \"following_url\": \"https://api.github.com/users/changjjjjjjj/following{/other_user}\",\n      \"gists_url\": \"https://api.github.com/users/changjjjjjjj/gists{/gist_id}\",\n      \"starred_url\": \"https://api.github.com/users/changjjjjjjj/starred{/owner}{/repo}\",\n      \"subscriptions_url\": \"https://api.github.com/users/changjjjjjjj/subscriptions\",\n      \"organizations_url\": \"https://api.github.com/users/changjjjjjjj/orgs\",\n      \"repos_url\": \"https://api.github.com/users/changjjjjjjj/repos\",\n      \"events_url\": \"https://api.github.com/users/changjjjjjjj/events{/privacy}\",\n      \"received_events_url\": \"https://api.github.com/users/changjjjjjjj/received_events\",\n      \"type\": \"User\",\n      \"site_admin\": false\n    },\n    \"body\": \"\",\n    \"state\": \"COMMENTED\",\n    \"html_url\": \"https://github.com/tmax-cloud/cicd-operator/pull/324#pullrequestreview-834851875\",\n    \"pull_request_url\": \"https://api.github.com/repos/tmax-cloud/cicd-operator/pulls/324\",\n    \"author_association\": \"COLLABORATOR\",\n    \"_links\": {\n      \"html\": {\n        \"href\": \"https://github.com/tmax-cloud/cicd-operator/pull/324#pullrequestreview-834851875\"\n      },\n      \"pull_request\": {\n        \"href\": \"https://api.github.com/repos/tmax-cloud/cicd-operator/pulls/324\"\n      }\n    },\n    \"submitted_at\": \"2021-12-17T05:36:07Z\",\n    \"commit_id\": \"654761e79f45e62ef8ca4d94c47cf7adc1756122\"\n  },\n  {\n    \"id\": 834860063,\n    \"node_id\": \"PRR_kwDOEm6Tx84xwvQf\",\n    \"user\": {\n      \"login\": \"changjjjjjjj\",\n      \"id\": 56624551,\n      \"node_id\": \"MDQ6VXNlcjU2NjI0NTUx\",\n      \"avatar_url\": \"https://avatars.githubusercontent.com/u/56624551?v=4\",\n      \"gravatar_id\": \"\",\n      \"url\": \"https://api.github.com/users/changjjjjjjj\",\n      \"html_url\": \"https://github.com/changjjjjjjj\",\n      \"followers_url\": \"https://api.github.com/users/changjjjjjjj/followers\",\n      \"following_url\": \"https://api.github.com/users/changjjjjjjj/following{/other_user}\",\n      \"gists_url\": \"https://api.github.com/users/changjjjjjjj/gists{/gist_id}\",\n      \"starred_url\": \"https://api.github.com/users/changjjjjjjj/starred{/owner}{/repo}\",\n      \"subscriptions_url\": \"https://api.github.com/users/changjjjjjjj/subscriptions\",\n      \"organizations_url\": \"https://api.github.com/users/changjjjjjjj/orgs\",\n      \"repos_url\": \"https://api.github.com/users/changjjjjjjj/repos\",\n      \"events_url\": \"https://api.github.com/users/changjjjjjjj/events{/privacy}\",\n      \"received_events_url\": \"https://api.github.com/users/changjjjjjjj/received_events\",\n      \"type\": \"User\",\n      \"site_admin\": false\n    },\n    \"body\": \"\",\n    \"state\": \"COMMENTED\",\n    \"html_url\": \"https://github.com/tmax-cloud/cicd-operator/pull/324#pullrequestreview-834860063\",\n    \"pull_request_url\": \"https://api.github.com/repos/tmax-cloud/cicd-operator/pulls/324\",\n    \"author_association\": \"COLLABORATOR\",\n    \"_links\": {\n      \"html\": {\n        \"href\": \"https://github.com/tmax-cloud/cicd-operator/pull/324#pullrequestreview-834860063\"\n      },\n      \"pull_request\": {\n        \"href\": \"https://api.github.com/repos/tmax-cloud/cicd-operator/pulls/324\"\n      }\n    },\n    \"submitted_at\": \"2021-12-17T05:57:08Z\",\n    \"commit_id\": \"d3b2006b7a2ab28268b248429bc215854a497d24\"\n  },\n  {\n    \"id\": 834871251,\n    \"node_id\": \"PRR_kwDOEm6Tx84xwx_T\",\n    \"user\": {\n      \"login\": \"yxzzzxh\",\n      \"id\": 36444454,\n      \"node_id\": \"MDQ6VXNlcjM2NDQ0NDU0\",\n      \"avatar_url\": \"https://avatars.githubusercontent.com/u/36444454?u=bbc82e004d2e79434274c1fc4ac97c1d2b6f249e&v=4\",\n      \"gravatar_id\": \"\",\n      \"url\": \"https://api.github.com/users/yxzzzxh\",\n      \"html_url\": \"https://github.com/yxzzzxh\",\n      \"followers_url\": \"https://api.github.com/users/yxzzzxh/followers\",\n      \"following_url\": \"https://api.github.com/users/yxzzzxh/following{/other_user}\",\n      \"gists_url\": \"https://api.github.com/users/yxzzzxh/gists{/gist_id}\",\n      \"starred_url\": \"https://api.github.com/users/yxzzzxh/starred{/owner}{/repo}\",\n      \"subscriptions_url\": \"https://api.github.com/users/yxzzzxh/subscriptions\",\n      \"organizations_url\": \"https://api.github.com/users/yxzzzxh/orgs\",\n      \"repos_url\": \"https://api.github.com/users/yxzzzxh/repos\",\n      \"events_url\": \"https://api.github.com/users/yxzzzxh/events{/privacy}\",\n      \"received_events_url\": \"https://api.github.com/users/yxzzzxh/received_events\",\n      \"type\": \"User\",\n      \"site_admin\": false\n    },\n    \"body\": \"/approve\",\n    \"state\": \"COMMENTED\",\n    \"html_url\": \"https://github.com/tmax-cloud/cicd-operator/pull/324#pullrequestreview-834871251\",\n    \"pull_request_url\": \"https://api.github.com/repos/tmax-cloud/cicd-operator/pulls/324\",\n    \"author_association\": \"CONTRIBUTOR\",\n    \"_links\": {\n      \"html\": {\n        \"href\": \"https://github.com/tmax-cloud/cicd-operator/pull/324#pullrequestreview-834871251\"\n      },\n      \"pull_request\": {\n        \"href\": \"https://api.github.com/repos/tmax-cloud/cicd-operator/pulls/324\"\n      }\n    },\n    \"submitted_at\": \"2021-12-17T06:21:13Z\",\n    \"commit_id\": \"d3b2006b7a2ab28268b248429bc215854a497d24\"\n  }\n]"
	sampleIssueComments = "[\n  {\n    \"url\": \"https://api.github.com/repos/tmax-cloud/cicd-operator/issues/comments/996468306\",\n    \"html_url\": \"https://github.com/tmax-cloud/cicd-operator/pull/324#issuecomment-996468306\",\n    \"issue_url\": \"https://api.github.com/repos/tmax-cloud/cicd-operator/issues/324\",\n    \"id\": 996468306,\n    \"node_id\": \"IC_kwDOEm6Tx847ZOZS\",\n    \"user\": {\n      \"login\": \"tmax-cloud-bot\",\n      \"id\": 76757421,\n      \"node_id\": \"MDQ6VXNlcjc2NzU3NDIx\",\n      \"avatar_url\": \"https://avatars.githubusercontent.com/u/76757421?v=4\",\n      \"gravatar_id\": \"\",\n      \"url\": \"https://api.github.com/users/tmax-cloud-bot\",\n      \"html_url\": \"https://github.com/tmax-cloud-bot\",\n      \"followers_url\": \"https://api.github.com/users/tmax-cloud-bot/followers\",\n      \"following_url\": \"https://api.github.com/users/tmax-cloud-bot/following{/other_user}\",\n      \"gists_url\": \"https://api.github.com/users/tmax-cloud-bot/gists{/gist_id}\",\n      \"starred_url\": \"https://api.github.com/users/tmax-cloud-bot/starred{/owner}{/repo}\",\n      \"subscriptions_url\": \"https://api.github.com/users/tmax-cloud-bot/subscriptions\",\n      \"organizations_url\": \"https://api.github.com/users/tmax-cloud-bot/orgs\",\n      \"repos_url\": \"https://api.github.com/users/tmax-cloud-bot/repos\",\n      \"events_url\": \"https://api.github.com/users/tmax-cloud-bot/events{/privacy}\",\n      \"received_events_url\": \"https://api.github.com/users/tmax-cloud-bot/received_events\",\n      \"type\": \"User\",\n      \"site_admin\": false\n    },\n    \"created_at\": \"2021-12-17T06:21:16Z\",\n    \"updated_at\": \"2021-12-17T06:21:16Z\",\n    \"author_association\": \"NONE\",\n    \"body\": \"[APPROVE ALERT]\\n\\nUser `yxzzzxh` approved this pull request!\",\n    \"reactions\": {\n      \"url\": \"https://api.github.com/repos/tmax-cloud/cicd-operator/issues/comments/996468306/reactions\",\n      \"total_count\": 0,\n      \"+1\": 0,\n      \"-1\": 0,\n      \"laugh\": 0,\n      \"hooray\": 0,\n      \"confused\": 0,\n      \"heart\": 0,\n      \"rocket\": 0,\n      \"eyes\": 0\n    },\n    \"performed_via_github_app\": null\n  }\n]"
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

func TestClient_ListComments(t *testing.T) {
	c, err := testEnv()
	if err != nil {
		t.Fatal(err)
	}

	comments, err := c.ListComments(5)
	require.NoError(t, err)
	require.Len(t, comments, 9)
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

func TestClient_ListLabels(t *testing.T) {
	c, err := testEnv()
	if err != nil {
		t.Fatal(err)
	}

	labels, err := c.ListLabels(5)
	require.NoError(t, err)
	require.Len(t, labels, 2)
	require.Equal(t, "approved", labels[0].Name)
	require.Equal(t, "size/L", labels[1].Name)
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
	r.HandleFunc("/repos/{org}/{repo}/issues/{id}/labels", func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte(sampleLabelLists))
	})
	r.HandleFunc("/repos/{org}/{repo}/pulls/{id}/comments", func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte(samplePRComments))
	})
	r.HandleFunc("/repos/{org}/{repo}/pulls/{id}/reviews", func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte(samplePRReviews))
	})
	r.HandleFunc("/repos/{org}/{repo}/issues/{id}/comments", func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte(sampleIssueComments))
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
