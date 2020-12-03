package v1

const (
	JobLabelPrefix = "cicd.tmax.io/"

	JobLabelConfig      = JobLabelPrefix + "integration-config"
	JobLabelType        = JobLabelPrefix + "integration-type"
	JobLabelId          = JobLabelPrefix + "integration-id"
	JobLabelRepository  = JobLabelPrefix + "repository"
	JobLabelPullRequest = JobLabelPrefix + "pull-request"

	RunLabelJob            = JobLabelPrefix + "integration-job"
	RunLabelJobId          = JobLabelPrefix + "integration-job-id"
	RunLabelRepository     = JobLabelRepository
	RunLabelPullRequest    = JobLabelPullRequest
	RunLabelPullRequestSha = JobLabelPrefix + "pull-request-sha"
	RunLabelSender         = JobLabelPrefix + "sender"
)
