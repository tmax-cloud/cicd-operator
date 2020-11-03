# Environmental Variables

|Key|Description|
|:-----------------:|---|
|`CI`               | Always set to true |
|`CI_CONFIG_NAME`   | The name of the `IntegrationConfig` |
|`CI_JOB_ID`        | The id of the `IntegrationJob` |
|`CI_REPOSITORY`    | Repository name. e.g., tmax-cloud/cicd-operator |
|`CI_EVENT_TYPE`    | The type of webhook event |
|`CI_EVENT_PATH`    | The path of the file with webhook body |
|`CI_WORKSPACE`     | Working directory, where the repository is cloned |
|`CI_SHA`           | The commit SHA which triggered the job |
|`CI_REF`           | The branch or tag ref which triggered the job |
|`CI_HEAD_REF`      | Only set for forked repository |
|`CI_BASE_REF`      | Only set for forked repository |
