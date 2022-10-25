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

package v1

// Labels for IntegrationJobs or PipelineRuns
const (
	// JobLabelPrefix is a prefix of every labels
	JobLabelPrefix = "cicd.tmax.io/"

	JobLabelConfig      = JobLabelPrefix + "integration-config"
	JobLabelType        = JobLabelPrefix + "integration-type"
	JobLabelID          = JobLabelPrefix + "integration-id"
	JobLabelRepository  = JobLabelPrefix + "repository"
	JobLabelPullRequest = JobLabelPrefix + "pull-request"

	RunLabelJob            = JobLabelPrefix + "integration-job"
	RunLabelJobID          = JobLabelPrefix + "integration-job-id"
	RunLabelRepository     = JobLabelRepository
	RunLabelPullRequest    = JobLabelPullRequest
	RunLabelPullRequestSha = JobLabelPrefix + "pull-request-sha"
	RunLabelSender         = JobLabelPrefix + "sender"
)
