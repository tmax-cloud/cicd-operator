package v1

const (
	JobLabelPrefix      = "cicd.tmax.io/"
	JobLabelConfig      = JobLabelPrefix + "integration-config"
	JobLabelType        = JobLabelPrefix + "integration-type"
	JobLabelId          = JobLabelPrefix + "integration-id"
	JobLabelRepository  = JobLabelPrefix + "repository"
	JobLabelPullRequest = JobLabelPrefix + "pull-request"
	JobLabelStatus      = JobLabelPrefix + "status"
)
