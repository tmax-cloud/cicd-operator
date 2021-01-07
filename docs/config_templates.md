# Configuring Templates
`Report` and `Email` features use golang's `html/template` feature to render user-facing pages.

Refer to https://golang.org/pkg/html/template/ for the syntax.

## Configuring Report Templates
You can check and update the report template from the ConfigMap `report-template` in namespace `cicd-system`.

Following structure is passed to compile the template.
```go
type report struct {
	JobName    string
	JobJobName string
	JobStatus  *cicdv1.JobStatus
	Log        string
}
```

## Configuring Email Templates
You can check and update the email template from the ConfigMap `email-template` in namespace `cicd-system`.

`Approval` is passed to compile the template.
