# Environmental Variables

Following variables are set to the job containers.

|Key|Description|
|:-----------------:|---|
|`CI`               | Always set to true |
|`CI_CONFIG_NAME`   | The name of the `IntegrationConfig` |
|`CI_JOB_ID`        | The id of the `IntegrationJob` |
|`CI_REPOSITORY`    | Repository name. e.g., tmax-cloud/cicd-operator |
|`CI_EVENT_TYPE`    | The type of webhook event (`PreSubmit` or `PostSubmit`) |
|`CI_WORKSPACE`     | Working directory, where the repository is cloned |
|`CI_HEAD_SHA`      | The commit SHA which triggered the job. For Multiple PRs, it would set to a single string seperated by white-spaces(" ") |
|`CI_HEAD_REF`      | The branch or tag ref which triggered the job. For Multiple PRs, it would set to a single string seperated by white-spaces(" ") |
|`CI_BASE_SHA`      | Only set for forked repository / pull request |
|`CI_BASE_REF`      | Only set for forked repository / pull request |
|`CI_SERVER_URL`    | Server URL. e.g., https://github.com |
|`CI_SENDER_NAME`   | Event sender's name |
|`CI_SENDER_EMAIL`  | Event sender's email |
