# Notification Jobs
This guide lets you know how to use Notification jobs (not jobs running in a pod)

* [Email](#email)
* [Slack](#slack)
* [Webhook](#webhook)

## `Email`
Add following 'email' job in `IntegrationConfig`

You can use following variables as substitutions in the title, content.
- `$INTEGRATION_JOB_NAME`: IntegrationJob's name
- `$JOB_NAME`: Job's name

Also, title and content are compiled using `IntegrationJob` struct.
```yaml
- name: email
  email:
     receivers:
        - sunghyun_kim3@tmax.co.kr
        - cqbqdd11519@gmail.com
     title: HTML Email title
     isHtml: true
     content: |
        <hr>
        <b>sdiofjsiodfjsdofj</b>
        <i>sdiofjsiodfjsdofj</i>
```

## `Slack`
Add following 'email' job in `IntegrationConfig`.

You can use following variables as substitutions in the message.
- `$INTEGRATION_JOB_NAME`: IntegrationJob's name
- `$JOB_NAME`: Job's name

Also, message is compiled using `IntegrationJob` struct.
```yaml
- name: slack
  slack:
    url: https://hooks.slack.com/services/....
    message: IntegrationJob($INTEGRATION_JOB_NAME)'s job($JOB_NAME) is running!
```

## `Webhook`
Add following 'webhook' job in `IntegrationConfig`.

You can use following variables as substitutions in the message.
- `$INTEGRATION_JOB_NAME`: IntegrationJob's name
- `$JOB_NAME`: Job's name

Also, message is compiled using `IntegrationJob` struct.
```yaml
- name: webhook
  webhook:
    url: https://webhookUrl
    body: IntegrationJob($INTEGRATION_JOB_NAME)'s job($JOB_NAME) is running!
```
