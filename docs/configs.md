# Operator Configurations

This guide shows how to configure the operator. Contents are as follows.
- [System Configurations](#system-configurations)
  - [`maxPipelineRun`](#maxpipelinerun)
  - [`ingressClass`](#ingressclass)
  - [`externalHostName`](#externalhostname)
  - [`gitImage`](#gitimage)
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

### `ingressClass`
Ingress's class name to be used for the webhook/report server access.

### `externalHostName`
External host name for the ingress. It should be the address a user/git server can access. Default address is `cicd-webhook.INGRESS_IP.nip.io`

### `gitImage`
Git image to be used for `git-checkout` steps
> Default: docker.io/alpine/git:1.0.30

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
