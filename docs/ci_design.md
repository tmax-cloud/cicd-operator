# CI Module Design

## Custom Resources
- `IntegrationConfig`  
: CI configuration for each git repository. It specifies jobs to be run on `pull_request` event or `push` event arrives.
- `IntegrationJob`  
: Each `IntegrationJob` is an instance for the CI tasks. For example, when a developer creates a pull request, then an `IntegrationJob` is created, referring to the corresponding `IntegrationConfig`.

## Modules
- Git Webhook  
(for GitHub, GitLab, Gitea, Bitbucket)
- Git client  
(for GitHub, GitLab, Gitea, Bitbucket)
- `IntegrationConfigController`
- `IntegrationJobController`

## Procedure 0. Common
1. A project manager creates an `IntegrationConfig` for a GitHub(for instance) repository. He specifies jobs to be executed when pull request is created (might be unit test, e2e test, lint, ...), or/and when code is pushed (might be image build, ...).
2. `IntegrationConfigController` registers CI/CD operator's webhook server to the GitHub repository, using git client module.

## Procedure 1. PullRequest - Creation/Modification
1. Developers write code, commit to a branch and create a pull request.
2. The `pull_request` event is delivered to a webhook server.
3. Webhook body (GitHub specific body) is converted to a generic webhook body structure.
4. Webhook calls `Integrator` plugin, which is registered for the `pull_request` event.  
FYI. CI/CD operator implementation should consist of **only one webhook server** and **a number of plugins**(`Integrator` is one of the plugins) for scalability.
5. `Integrator` plugin creates an `IntegrationJob`, referring to the `IntegrationConfig`'s `.spec.jobs.preSubmit` field.
6. `IntegrationJobController` detects that the `IntegrationJob` is created and creates Tekton `PipelineRun`.  
*Further: Should have a work queue of `IntegrationJob`s, instead of creating `PipelineRun` instantly.*
7. `IntegrationJobController` watches the `PipelineRun` and whenever tasks end, it reports to GitHub repository by setting commit status, using git client module.

## Procedure 3. Push (PullRequest Merged)
1. Repository owner merges the pull request.  
*Further: Auto merge should be enabled, like `tide` in Prow*
2. The `push` event for the base branch is delivered to a webhook server.
3. Webhook body (GitHub specific body) is converted to a generic webhook body structure.
4. Webhook calls `Integrator` plugin, which is registered for the `push` event.  
5. `Integrator` plugin creates an `IntegrationJob`, referring to the `IntegrationConfig`'s `.spec.jobs.postSubmit` field.
6. Rest is same with `pull_request` event handling.

## Procedure 2. PullRequest - Comment
1. For an open pull request, a developer comments `/test <job name>`.  
(e.g., the test should be re-run because the base branch is updated.)
2. The `issue_comment` event is delivered to a webhook server.
3. Webhook body (GitHub specific body) is converted to a generic webhook body structure.
4. Webhook calls `ChatOps` plugin, which is registered for the `issue_comment` event.
5. `ChatOps` calls `Integrator` to create `IntegrationJob`.
6. Rest is same with `pull_request` event handling.

## Important Facts
- For pull request tasks, every tasks (e.g., unit test, e2e test) are executed with the merged codes. i.e., `git merge <pull request sha>` is executed for base branch (e.g., master) before all the tasks are carried on.
