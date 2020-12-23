# `IntegrationConfig` Spec

## Available Fields
```yaml
apiVersion: cicd.tmax.io/v1
kind: IntegrationConfig
metadata:
  name: <Name>
spec:
  git:
    type: [github|gitlab|gitea|bitbucket]
    repository: <org>/<repo> (e.g., tmax-cloud/cicd-operator)
    apiUrl: <API server URL>
    token:
      value: <Token value>
      valueFrom:
        secretKeyRef:
          name: <Token secret name>
          key: <Token secret key>
  secrets:
    - name: <Secret name to be included in a service account>
  jobs:
    preSubmit:
    - name: <Job name>
      image: <Job image>
      command:
      - <Command>
      script: <Script>
      resources:
        requests:
          memory: "64Mi"
          cpu: "250m"
        limits:
          memory: "128Mi"
          cpu: "500m"
      env:
      - name: TEST
        value: val
      when:
        branch:
        - <RegExp>
        skipBranch:
        - <RegExp>
        tag:
        - <RegExp>
        skipTag:
        - <RegExp>
        ref:
        - <RegExp>
        skipRef:
        - <RegExp>
      after:
      - <Job Name>
      approval:
        approvers:
        - <List (comma-seperated user names)>
        approversConfigMap:
          name: <ConfigMap name>
    postSubmit:
    - <Same as preSubmit>
status:
  secrets: <Webhook secret>
  conditions:
  - type: WebhookRegistered
    status: [True|False]
    reason: <Reason of the condition status>
    message: <Message for the condition status>
```

## Sample YAML
```yaml
apiVersion: cicd.tmax.io/v1
kind: IntegrationConfig
metadata:
  name: sample-config
  namespace: default
spec:
  git:
    type: github 
    repository: tmax-cloud/cicd-operator
    token:
      valueFrom:
        secretKeyRef:
          name: tmax-cloud-bot-credential
          key: token
  secrets:
    - name: tmax-cloud-hub
  jobs:
    preSubmit:
    - name: test-unit
      image: golang:1.14
      command:
      - go test -v ./pkg/...
      when:
        branch:
        - master
    - name: test-lint
      image: golangci/golangci-lint:v1.32
      command:
      - golangci-lint run ./... -v -E gofmt --timeout 1h0m0s
      when:
        branch:
        - master
    postSubmit:
    - name: build-push-image
      image: quay.io/buildah/stable
      command:
      - buildah bud --format docker --storage-driver=vfs -f ./Dockerfile -t $IMAGE_URL .
      - buildah push --storage-driver=vfs --creds=$CRED $IMAGE_URL docker://$IMAGE_URL
      env:
      - name: IMAGE_URL
        value: tmaxcloudck/cicd-operator:recent
      - name: CRED
        valueFrom:
          secretKeyRef:
             name: tmaxcloudck-hub-credential
             key: .dockerconfigjson
      privileged: true
      when:
        branch:
        - master
```
