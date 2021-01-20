package configs

// Configs to be configured by command line arguments

var (
	// MaxPipelineRun is the number of PipelineRuns that can run simultaneously
	MaxPipelineRun int

	// ExternalHostName to be used for webhook server (default is ingress host name)
	ExternalHostName string

	// ReportRedirectUriTemplate is a uri template for report page redirection
	ReportRedirectUriTemplate string

	// CollectPeriod is a garbage collection period (in hour)
	CollectPeriod int

	// JobTTL is a garbage collection threshold (in hour).
	// If IntegrationJob's .status.completionTime + TTL < now, it's collected
	IntegrationJobTTL int

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

	// IngressClass is a class for ingress instance
	IngressClass string
)
