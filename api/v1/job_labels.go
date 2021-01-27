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
