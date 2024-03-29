# Operator Configurations

This guide shows how to configure the operator. Contents are as follows.
- [System Configurations](#system-configurations)
  - [`maxPipelineRun`](#maxpipelinerun)
  - [`exposeMode`](#exposemode)
  - [`ingressClass`](#ingressclass)
  - [`ingressHost`](#ingresshost)
  - [`externalHostName`](#externalhostname)
  - [`gitImage`](#gitimage)
  - [`gitCheckoutStepCPURequest`](#gitcheckoutstepcpurequest)
  - [`gitCheckoutStepMemRequest`](#gitcheckoutstepmemrequest)
  - [`reportRedirectUriTemplate`](#reportredirecturitemplate)
- [Email Configurations](#email-configurations)
  - [`enableMail`](#enablemail)
  - [`smtpHost`](#smtphost)
  - [`smtpUserSecret`](#smtpusersecret)
- [Garbage Collector Configurations](#garbage-collector-configurations)
  - [`collectPeriod`](#collectperiod)
  - [`integrationJobTTL`](#integrationjobttl)

You can check and update the configuration values from the ConfigMap `cicd-config` in namespace `cicd-system`.
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cicd-config
  namespace: cicd-system
data:
  maxPipelineRun: "5"
  externalHostName: ""
  enableMail: "false"
  smtpHost: ""
  smtpUserSecret: ""
  collectPeriod: "120"
  integrationJobTTL: "120"
  ingressClass: ""
```

## System Configurations
### `maxPipelineRun`
Maximum number of PipelineRuns which can run in same time.
> Default: 5

### `exposeMode`
ExposeMode is a mode to be used for exposing the webhook server (Ingress/LoadBalancer/ClusterIP)
> Default: Ingress

### `ingressClass`
Ingress's class name to be used for the webhook/report server access.

### `ingressHost`
Ingress's host name for the webhook/report server access.
> Default: cicd-webhook.{Ingress Controller IP}.nip.io

### `externalHostName`
External host name for the ingress. It should be the address a user/git server can access. Default address is `cicd-webhook.INGRESS_IP.nip.io`

### `gitImage`
Git image to be used for `git-checkout` steps
> Default: docker.io/alpine/git:1.0.30

### `gitCheckoutStepCPURequest`
Resource (CPU) requirement for git checkout step
> Default: 30m

### `gitCheckoutStepMemRequest`
Resource (Memory) requirement for git checkout step
> Default: 100Mi

### `reportRedirectUriTemplate`
Url template of commit status's detail page, which is compiled using `IntegrationJob` struct. If it's empty, it uses default report page.

## Email Configurations
### `enableMail`
Whether to enable email feature. If it's true, `smtpHost` and `smtpUserSecret` should be configured.
> Default: false
### `smtpHost`
SMTP server host (e.g., `smtp.gmail.com:465`)
### `smtpUserSecret`
Secret name for SMTP user credential. The secret's kind should be `kubernetes.io/basic-auth`

## Garbage Collector Configurations
Garbage collector deletes outdated `IntegrationJobs`.
### `collectPeriod`
Garbage collection period (in hours)
> Default: 120

### `integrationJobTTL`
TTL of `IntegrationJob`s (in hours). `IntegrationJobs` after the TTL would be collected.
> Default: 120
