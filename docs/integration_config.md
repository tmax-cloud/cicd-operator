# `IntegrationConfig` Spec

This guide shows how to configure `IntegrationConfig` in detail.

- [Configuring `git`](#configuring-git)
  - [`type`](#type)
  - [`apiUrl`](#apiurl)
  - [`repository`](#repository)
  - [`token`](#token)
    - [Token value](#token-value)
    - [Token from Secret](#token-from-secret)
- [Configuring `jobs`](#configuring-jobs)
  - [Category of jobs](#category-of-jobs)
  - [Configuring normal jobs](#configuring-normal-jobs)
  - [`skipCheckout`](#skipcheckout)
  - [`when`](#when)
  - [`after`](#after)
  - [Configuring `approval` jobs](#configuring-approval-jobs)
  - [Configuring `email` jobs](#configuring-email-jobs)
  - [Using Tekton Tasks](#using-tekton-tasks)
- [Configuring `secrets`](#configuring-secrets)
- [Configuring `workspaces`](#configuring-workspaces)

## Configuring `git`
For example,
```yaml
spec:
  git:
    type: gitlab
    apiUrl: http://gitlab.my.domain
    repository: root/my-repo
    token:
      value: woeifj93fjsikljv49wfjkfn12
```
### `type`
It is a type of git remote server.
> **Required**  
> Available values: github, gitlab

### `apiUrl`
API server url for self-served git servers. (e.g., http://gitlab.my.domain)
> Optional

### `repository`
> **Required**  
> Available value: < Owner >/< Repo >

### `token`
Access token for accessing the repository. (It registers webhook, commit statuses)
> **Required**

### Token value
Stores token value itself in the yaml. **Not recommended due to a security issue**

### Token from Secret
Takes token value from existing secret. **Recommended**
```yaml
spec:
  git:
    ...
    token:
      valueFrom:
        secretKeyRef:
          name: my-git-secret
          key: my-token-key
```

## Configuring `jobs`
### Category of jobs
- **Pre-submit jobs**  
  Pre-submit jobs are executed when a pull request (or merge request) is opened, reopened, or new commits are added.
- **Post-submit jobs**  
  Post-submit jobs are executed when commits are pushed. It includes when code is pushed, pull request is merged, and tag is pushed.
### Configuring normal jobs
Jobs are in same shape of Tekton's steps. Refer to https://github.com/tektoncd/pipeline/blob/master/docs/tasks.md#defining-steps
```yaml
spec:
  jobs:
    preSubmit:
      - name: test
        image: maven:3.6.3-openjdk-16-slim
        script: |
          mvn test
```

### `skipCheckout`
Whether to skip git checkout or not. If you don't need a git source for a job, you can set it as true.
> Optional  
> Available values: true, false  
> Default value: false
```yaml
spec:
  jobs:
    preSubmit:
      - name: test
        ...
        skipCheckout: true
```

### `when`
If you want this job to be executed only for specific branches or tags, you can specify here.

**All values for the fields should be in valid regular expression**  
**At most one category should be configured, among branch-related and tag-related**

> Optional  
> Available fields: branch, skipBranch, tag, skipTag
```yaml
spec:
  jobs:
    preSubmit:
      - name: test
        ...
        when:
          branch:
            - master
    postSubmit:
      - name: release
        ...
        when:
          skipTag:
            - test-.*
```

### `after`
If you want this job to be executed after specific jobs, you can specify here.
> Optional  
```yaml
spec:
  jobs:
    preSubmit:
      - name: pre-process
        ...
      - name: test
        ...
        after:
          - pre-process
```

### Configuring `approval` jobs
Refer to the [`Approval` guide](./approval.md)

### Configuring `email` jobs
Refer to the [`Email` guide](./email.md)

### Using Tekton Tasks
You can use (or reuse) tekton tasks, rather than writing scripts in `IntegrationCOnfig`.
* `params` are slightly different from the Tekton's spec. You should specify `name` field and one of `stringVal` field or `arrayVal` field, depending on the parameter's type.
* You can use `resources` just like using in `PipelineRun` or `TaskRun`, either using `resourceRef` or `resourceSpec`.
* You can use `workspaces`, just like using in `Pipeline`. Be sure you specified the workspace in `workspaces` field of `IntegrationConfig` (refer to [the link](#configuring-workspaces))
```yaml
spec:
  jobs:
    preSubmit:
    - name: test-1
      tektonTask:
        taskRef:
          local:
            name: curl
            kind: Task
        params:
          - name: BUILDER_IMAGE
            stringVal: tmaxcloudck/s2i-tomcat:latest
        resources:
          outputs:
            - name: image
              resourceSpec:
                type: image
                params:
                  - name: url
                    value: 172.22.11.2:30500/test:test
        workspaces:
          - name: source
            workspace: s2i
```
For users' convenience, we provide catalog reference. The catalog is fetched from the Tekton's catalog repository (https://github.com/tektoncd/catalog/tree/master/task).
Be sure to specify the catalog name as `[name]@[version]` (e.g., `s2i@0.1`)
```yaml
spec:
  jobs:
    - name: test-1
      tektonTask:
        taskRef:
          catalog: 's2i@0.1'
        params:
          - name: BUILDER_IMAGE
            stringVal: tmaxcloudck/s2i-tomcat:latest
        resources:
          outputs:
            - name: image
              resourceSpec:
                type: image
                params:
                  - name: url
                    value: 172.22.11.2:30500/test:test
        workspaces:
          - name: source
            workspace: s2i
```


## Configuring `secrets`
Secrets in this field are included in the service account, which is automatically generated.
Useful for configuring Docker hub secrets.
```yaml
spec:
  secrets:
    - my-docker-hub-secret
```

## Configuring `workspaces`
Workspaces are used for sharing data between jobs. The spec is same as Tekton's workspace.

If you specify workspaces here, every job has the volume mounted, with the specified name.
The path of the volume is `$(workspaces.NAME.path)`, just same as Tekton's spec.

Refer to https://github.com/tektoncd/pipeline/blob/master/docs/workspaces.md
```yaml
spec:
  workspaces:
    - name: s2i
      volumeClaimTemplate:
        spec:
          storageClassName: local-path
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
  jobs:
    - name: test
      ...
      script: |
        echo 'hi' >> $(workspaces.s2i.path)/hello-file
```

# Appendix
## All Available Fields
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
  workspaces:
    - name: <Name of workspace>
      ...
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
  workspaces:
    - name: s2i
      volumeClaimTemplate:
        spec:
          storageClassName: local-path
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
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
        tag:
        - v.*
```
