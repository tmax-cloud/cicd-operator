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
      sha: <SHA of base repo.>
      link: <Link of base repo.>
    pull:
      id: <Pull request ID>
      sha: <SHA of the pull request commit>
      link: <Link of the pull request>
      author: 
        name: <Author name>
        link: <Author link>
status:
  state: [pending | running | completed | failed]
  taskStatus:
  - <Tekton Task Status>
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
    cicd.tmax.io/pull-request: 32
    cicd.tmax.io/integration-id: fudsfh389s234fasdf323df3fxf5df
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
      sha: 9d9a5bf747c6f7327ad420458e8bfb8bbe340b43
      link: https://github.com/tmax-cloud/cicd-operator
    pull:
      id: 32
      sha: 48f6ce4dd655a64f723de28695e3322502183c79
      link: https://github.com/tmax-cloud/cicd-operator/pull/32
      author: 
        name: cqbqdd11519
        link: https://github.com/cqbqdd11519
status:
  state: running
  taskStatus:
  - <Tekton Task Status>
```
