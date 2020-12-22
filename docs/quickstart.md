# Quick Start Guide

This guide leads you to start using CI/CD operator, with simple pull-request/push examples.
The contents are as follows.

* [You need](#you-need)
* [Create bot account and token](#create-bot-account-and-token)
* [Create IntegrationConfig](#create-integrationconfig)
* [Create Pull Request](#create-pull-request)
* [Merge Pull Request](#merge-pull-request)
* [Release](#release)

## You need...
- Git repository (either GitHub or GitLab)
- K8s cluster for the jobs to run

## Create bot account and token
1. Create account/token
- For GitHub  
  - Create a new bot account
  - Create an access token for the bot account  
    `https://github.com/settings/tokens > Generate a new token`  
    Scope:
    * repo
    * admin:repo_hook

- For GitLab
    - Create a new bot account
    - Create an access token for the bot account
      `https://gitlab.com/-/profile/personal_access_tokens`  
      Scope:
      * api

2. Copy generated token and store it as a secret
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: bot-token
stringData:
  token: <TOKEN>
```

## Create `IntegrationConfig`
```yaml
apiVersion: cicd.tmax.io/v1
kind: IntegrationConfig
metadata:
  name: tutorial-config
spec:
  git:
    type: github # If you are using gitlab, use 'gitlab'
    repository: my/repository
    token:
      valueFrom:
        secretKeyRef:
          name: bot-token
          key: token
  jobs:
    preSubmit: # Jobs to be executed on PullRequest
      - name: test
        image: golang:1.14
        script: go test ./...
    postSubmit: # Jobs to be executed on Push/Tag
      - name: image
        image: quay.io/buildah/stable
        script: |
          buildah bud --format docker --storage-driver=overlay -f ./Dockerfile -t $IMAGE_URL:${CI_HEAD_REF#refs/tags/} .
          buildah push --storage-driver=vfs $IMAGE_URL:${CI_HEAD_REF#refs/tags/} docker://$IMAGE_URL:${CI_HEAD_REF#refs/tags/}
        env:
          - name: IMAGE_URL
            value: my-repository/my-image
        securityContext:
          privileged: true
        when:
          branch:
            - master
```

Check if the webhook is registered for the repository.
Check if the conditions of the created `IntegrationConfig` object are true.

## Create Pull Request/Merge Request
1. Create a new branch from `master` branch
2. Create a new commit for the branch & push it
3. Create PullRequest (for GitHub) or MergeRequest (for GitLab) of the branch to master
4. Check `IntegrationJob` generation
   ```bash
   kubectl get integrationjob
   ```
5. Check PullRequest/MergeRequest details page to see if `test` job is running.

## Merge Pull Request/Merge Request
1. Merge the PullRequest/MergeRequest to master
2. Check `IntegrationJob` generation
   ```bash
   kubectl get integrationjob
   ```
3. Check commit details page to see if `image` job is running.
