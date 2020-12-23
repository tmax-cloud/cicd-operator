package configs

// Configs to be configured by command line arguments

var (
	// MaxPipelineRun is the number of PipelineRuns that can run simultaneously
	MaxPipelineRun int

	// ExternalHostName to be used for webhook server (default is ingress host name)
	ExternalHostName string

	// EnableMail
	EnableMail bool

	// SMTPHost string
	SMTPHost string

	// SMTPUserSecret string
	SMTPUserSecret string

	// Email templates
	ApprovalRequestMailTitle   string
	ApprovalRequestMailContent string

	ApprovalResultMailTitle   string
	ApprovalResultMailContent string
)
