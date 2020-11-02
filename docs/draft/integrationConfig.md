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
    url: <Git URL>
    token:
      value: <Token value>
      valueFrom:
        secretKeyRef:
          name: <Token secret name>
          key: <Token secret key>
  jobs:
    preSubmit:
    - name: <Job name>
      image: <Job image>
      command:
      - <Command>
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
      privileged: false
      tektonTask:
        taskRef:
          name: <Tekton Task name>
          catalog: <Tekton Catalog name>
        params:
        - name: <Param name>
          value: <Param value>
        resources:
          inputs:
          - <Refer to tekton input>
          outputs:
          - <Refer to tekton input>
        workspaces:
        - <Refer to tekton workspace>
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
          list: <List (comma-seperated user names)>
          listFrom:
            configMapKeyRef:
              name: <ConfigMap name>
              key: <ConfigMap key>
      mailNotification:
        server: <Mail-notifier server address>
        subject: <Mail subject>
        content: <Mail content>
        list: <Mail recepient list (comma-seperated user emails)>
        listFrom:
          configMapKeyRef:
            name: <ConfigMap Name>
            key: <ConfigMap Key>
    postSubmit:
    - <Same as preSubmit>
  merge: // TODO
status:
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
    url: https://github.com/tmax-cloud/cicd-operator
    token:
      valueFrom:
        secretKeyRef:
          name: tmax-cloud-bot-credential
          key: token
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
