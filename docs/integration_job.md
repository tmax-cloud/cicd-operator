# `IntegrationJob` Spec

## Available Fields
```yaml
apiVersion: cicd.tmax.io/v1
kind: IntegrationJob
metadata:
  name: <Name>
spec:
  configRef:
    name: <IntegrationConfig Name>
    type: [presubmit|postsubmit]
  id: <Rand String>
  jobs:
  - <Same with IntegrationConfig spec.jobs.[preSubmit|postSubmit]>
  refs:
    repository: <e.g., tamx-cloud>/<e.g., cicd-operator>
    link: <e.g., https://github.com/tmax-cloud/cicd-operator>
    base:
      ref: <e.g., master>
      sha: <SHA of base commit>
      link: <Link of base repo.>
    pulls:
    - id: <Pull request ID>
      sha: <SHA of the pull request commit>
      link: <Link of the pull request>
      author: 
        name: <Author name>
status:
  state: [pending | running | completed | failed]
  startTime: <Started timestamp>
  completionTime: <Completed timestamp>
  jobs:
  - name: <job's name>
    startTime: <Started timestamp>
    completionTime: <Completed timestamp>
    state: [success | failure | error | pending]
    message: <Message of the job>
    podName: <Pod's name where the job is running>
    containers:
      - <Container status>
```

## Sample YAML
```yaml
apiVersion: cicd.tmax.io/v1
kind: IntegrationJob
metadata:
  name: sample-job
  namespace: default
  labels:
    cicd.tmax.io/integration-config: sample-config 
    cicd.tmax.io/integration-type: presubmit
    cicd.tmax.io/integration-id: fudsfh389s234fasdf323df3fxf5df
    cicd.tmax.io/pull-request: 32
spec:
  configRef:
    name: sample-config
    type: presubmit
  id: fudsfh389s234fasdf323df3fxf5df
  jobs:
  - name: test-unit
    image: golang:1.14
    command:
    - go test -v ./pkg/...
  - name: test-lint
    image: golangci/golangci-lint:v1.32
    command:
    - golangci-lint run ./... -v -E gofmt --timeout 1h0m0s
  refs:
    repository: tamx-cloud/cicd-operator
    link: https://github.com/tmax-cloud/cicd-operator
    base:
      ref: master
      sha: wef89weaf8wefje8cn2jd8c83nd9322502183c79
      link: https://github.com/tmax-cloud/cicd-operator
    pulls:
    - id: 32
      sha: 48f6ce4dd655a64f723de28695e3322502183c79
      link: https://github.com/tmax-cloud/cicd-operator/pull/32
      author: 
        name: sunghyunkim3
```
