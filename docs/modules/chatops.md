# ChatOps

ChatOps is a module that enables git-users to trigger jobs by commenting to the issues/pull requests.
This is a kind of plugin registered to the webhook server.

Handles following events
- `issue_comment` for GitHub
- `pull_request_review` for GitHub
- `pull_request_review_comment` for GitHub
- `Note Hook` for GitLab
