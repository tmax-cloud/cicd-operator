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
  - [`notification`](#notification)
  - [`tektonWhen`](#tektonwhen)
  - [`results`](#results)
  - [Configuring `approval` jobs](#configuring-approval-jobs)
  - [Configuring Notification jobs](#configuring-notification-jobs)
  - [Using Tekton Tasks](#using-tekton-tasks)
- [Configuring `secrets`](#configuring-secrets)
- [Configuring `workspaces`](#configuring-workspaces)
- [Configuring `podTemplate`](#configuring-podtemplate)
- [Configuring `mergeConfig`](#configuring-mergeconfig)
    - [`method`](#method)
    - [`commitTemplate`](#committemplate)
    - [`query`](#query)
- [Configuring `ijManageSpec`](#configuring-ijmanagespec)
- [Configuring `paramConfig`](#configuring-paramconfig)
    - [`paramDefine`](#paramdefine)
    - [`paramValue`](#paramvalue)
- [Configuring `TLSConfig`](#configuring-tlsconfig)
- [Triggering jobs](#triggering-jobs)
  - [Option.1 Using `cicdctl`](#option1-using-cicdctl)
  - [Option.2 Using `curl`](#option2-using-curl)

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
**This should NOT contain repository path (e.g., tmax-cloud/cicd-operator)**
> Optional

### `repository`
> **Required**  
> Available value: < Owner >/< Repo >

### `token`
Access token for accessing the repository. (It registers webhook, commit statuses)
> Optional

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

### `notification`
If you want to send notification when the job succeeded/failed, you can specify it in `notification` field.
The field's spec is same as [Notification Jobs](./notification-jobs.md)
> Optional  
```yaml
spec:
  jobs:
    preSubmit:
      - name: pre-process
        ...
      - name: test
        ...
        notification:
          onSuccess:
            email:
              receivers:
                - sunghyun_kim3@tmax.co.kr
              title: IntegrationJob $INTEGRATION_JOB_NAME Succeeded
              isHtml: true
              content: |
                <hr>
                <b>IntegrationJob: $INTEGRATION_JOB_NAME</b>
                <i>Job: $JOb_NAME</i>
          onFailure:
            slack:
              url: https://hooks.slack.com/services/....
              message: IntegrationJob($INTEGRATION_JOB_NAME)'s job($JOB_NAME) failed
```

### `tektonWhen`
You can use TektonWhen field to execute jobs conditionally.
The components of `when` expressions are `input`, `operator` and `values`:
- `input` is the input for the `when` expression which can be static inputs or variables ([`Parameters`](#configuring-paramconfig) or [`Results`](#results)).
- `operator` represents an `input`'s relationship to a set of `values`. A valid `operator` must be provided, which can be either `in` or `notin`.
- `values` is an array of string values. The `values` array must be provided and be non-empty. It can contain static values or variables ([`Parameters`](#configuring-paramconfig), [`Results`](#results) ).

> Optional  
```yaml
# Conditional execution using task results
spec:
  jobs:
    preSubmit:
      - name: test
        ...
        tektonWhen:
        - input: "$(tasks.<task-name>.results.<result-name>)"
          operator: in
          values: ["true"]

# Conditional execution using parameters
spec:
  jobs:
    preSubmit:
      - name: test
        ...
        tektonWhen:
        - input: "$(params.<param-name>)"
          operator: in
          values: ["true"]
```

### `results`
You can use results to pass task's results to [`Parameters`](#configuring-paramconfig) or [`TektonWhen`](#tektonWhen)
Define results and use `$(results.<task-name>.path)` form to emit task's results.
To get emitted results, use `$(tasks.<task-name>.results.<result-name>)` form.
> Optional  
```yaml
spec:
  jobs:
    preSubmit:
      - name: test-result
        image: bash:latest
        script: |
          #!/usr/bin/env bash
          echo -n "true"  | tee -a $(results.test-result.path)
        results:
        - name: test-result
          description: test result
```


### Configuring `approval` jobs
Refer to the [`Approval` guide](./approval.md)

### Configuring Notification jobs
Refer to the [`Notifications` guide](./notification-jobs.md)

### Using Tekton Tasks
You can use (or reuse) tekton tasks, rather than writing scripts in `IntegrationConfig`.
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
    - name: my-docker-hub-secret
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

## Configuring `podTemplate`
You can specify pod's additional spec for running the jobs. It is just same as tekton's `podTemplate`, so please refer to https://github.com/tektoncd/pipeline/blob/master/docs/podtemplates.md
```yaml
spec:
  jobs:
    - name: test
      ...
  podTemplate:
    securityContext:
      runAsNonRoot: true
      runAsUser: 1001
    imagePullSecrets:
      - name: pull-secret-1
```

## Configuring `mergeConfig`
*Currently, an ALPHA feature*

Merge automation can be configured using `mergeConfig`.
### `method`
`method` field specifies the method to merge the PR.
> Optional  
> Available values: `squash`, `merge`  
> Default: `merge`

### `commitTemplate`
`commitTemplate` specifies the title template of the merge commit. It should be a form of [golang template](https://pkg.go.dev/text/template).
The template is compiled using a structure [`blocker.PullRequest`](../pkg/blocker/blocker.go)
> Optional  
> Default: `{{ .Title }}({{ .ID }})`

### `query`
`query` is a selector of PRs to be merged. (i.e., conditions of PRs to be merged)
PRs are searched using the query and merged if all the CI checks are completed.
There are 9 kinds of queries. `labels`, `blockLabels`, `authors`, `skipAuthors`, `branches`, `skipBranches`, `checks`, `optionalChecks`, and `approveRequired`.

## Configuring `ijManageSpec`
IJManageSpec is used to define parameters to manage integration jobs. 
Currently provide timeout spec for garbage collection.
Timeout should be formed as [duration string](https://golang.org/pkg/time/#ParseDuration).

```yaml
spec:
  jobs:
    - name: test
      ...
  ijManageSpec:
    timeout: "2h"
```

## Configuring `paramConfig`
Parameters can be configured by `paramConfig`. 
Defined parameters are converted to tekton param when PipelineRun is created.
You can use parameter values in form of `$(params.<param-name>)`
### `paramDefine`
`paramDefine` field can define parameter's name, description & default values.
default values can be defined in form of string array or string
### `paramValue`
`paramValue` field can specify values of parameters in string or string array
```yaml
spec:
  jobs:
    - name: test
      ...
  paramConfig: 
    paramDefine:
    - name: "test-param"
    paramValue:
    - name: "test-param"
      stringVal: "true"
```

## Configuring `tlsConfig`
TLSConfig is used to define parameters for TLS. 
Currently provide InsecureSkipVerify flag.
Set true if you want to accept any certificate.

```yaml
spec:
  jobs:
    - name: test
      ...
  tlsConfig:
    insecureSkipVerify: true
```


## Triggering jobs
Although the jobs are triggered via git event, you can manually trigger them by calling API request.
### Option.1 Using `cicdctl`
```bash
cicdctl run -n <Namespace> post <IntegrationConfig Name>
```
### Option.2 Using `curl`
1. Find the user's token.
   If you are using ServiceAccount, you can find your token with following command
   ```bash
   SERVICE_ACCOUNT=<Name of the service account>
   kubectl get secret $(kubectl get serviceaccount $SERVICE_ACCOUNT -o jsonpath='{.secrets[].name}') -o jsonpath='{.data.token}' | base64 -d
   ```
2. Run API call to Kubernetes API server
   1. Trigger PullRequest event
   ```bash
   KUBERNETES_API_SERVER=<Kubernetes api server host:port>
   TOKEN=<Token got from 1.>

   INTEGRATION_CONFIG=<Name of the IntegrationConfig object>
   NAMESPACE=<Namespace where the IntegrationConfig exists>

   BASE_BRANCH="master"
   HEAD_BRANCH="feat/new-feat"

   curl -k -X POST \
   -H "Authorization: Bearer $TOKEN" \
   -d "{\"base_branch\": \"$BASE_BRANCH\", \"head_branch\": \"$HEAD_BRANCH\"}"
   "$KUBERNETES_API_SERVER/apis/cicdapi.tmax.io/v1/namespaces/$NAMESPACE/integrationconfigs/$INTEGRATION_CONFIG/runpre"
   ```

   2. Trigger Push event
   ```bash
   KUBERNETES_API_SERVER=<Kubernetes api server host:port>
   TOKEN=<Token got from 1.>

   INTEGRATION_CONFIG=<Name of the IntegrationConfig object>
   NAMESPACE=<Namespace where the IntegrationConfig exists>

   BRANCH="master"

   curl -k -X POST \
   -H "Authorization: Bearer $TOKEN" \
   -d "{\"branch\": \"$BRANCH\"}"
   "$KUBERNETES_API_SERVER/apis/cicdapi.tmax.io/v1/namespaces/$NAMESPACE/integrationconfigs/$INTEGRATION_CONFIG/runpost"
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
        - name: <User name>
          email: <User email>
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
