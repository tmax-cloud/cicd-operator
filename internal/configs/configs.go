package configs

// Configs to be configured by command line arguments

var (
	// MaxPipelineRun is the number of PipelineRuns that can run simultaneously
	MaxPipelineRun int

	// ExternalHostName to be used for webhook server (default is ingress host name)
	ExternalHostName string

	// ReportRedirectURITemplate is a uri template for report page redirection
	ReportRedirectURITemplate string

	// CollectPeriod is a garbage collection period (in hour)
	CollectPeriod int

	// IntegrationJobTTL is a garbage collection threshold (in hour).
	// If IntegrationJob's .status.completionTime + TTL < now, it's collected
	IntegrationJobTTL int

	// EnableMail is whether to enable mail feature or not
	EnableMail bool

	// SMTPHost is a host (IP:PORT) of the SMTP server
	SMTPHost string

	// SMTPUserSecret is a credential secret for the SMTP server (should be basic type)
	SMTPUserSecret string

	// ApprovalRequestMailTitle is a title for the approval request mail
	ApprovalRequestMailTitle string
	// ApprovalRequestMailContent is a content of the approval request mail
	ApprovalRequestMailContent string

	// ApprovalResultMailTitle is a title for the approval result mail
	ApprovalResultMailTitle string
	// ApprovalResultMailContent is a content of the approval result mail
	ApprovalResultMailContent string

	// IngressClass is a class for ingress instance
	IngressClass string

	// IngressHost is a host for ingress instance
	IngressHost string
)
