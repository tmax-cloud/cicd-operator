# `cicdctl`

`cicdctl` is a command-line tool for calling extension APIs for CI/CD operator.

## Building `cicdctl` binary
```bash
make cicdctl
```

## Supported commands
- [Run](#run)
- [Approve](#approve)
- [Reject](#reject)
- [Webhook](#webhook)

### Run
`Run` command triggers jobs of an `IntegrationConfig`.
#### Command
`cicdctl run [pre|post] [IntegrationConfig Name]`
#### Options
|Name|Description|
|---|---|
|`head-branch`| Head branch of the git repository|
|`base-branch`| Base branch of the git repository|
#### Examples
```bash
# Running preSubmit jobs
$ cicdctl run pre -n default ic-test --head-branch test --base-branch master
Triggered pre jobs for IntegrationConfig default/ic-test

# Running postSubmit jobs
$ cicdctl run post -n default ic-test --branch master
Triggered post jobs for IntegrationConfig default/ic-test
```

### Approve
`Approve` command approves an `Approval`
#### Command
`cicdctl approve [Approval Name] [Reason]`
#### Examples
```bash
$ cicdctl approve -n default approval-test 'just because'
Approved Approval default/approval-test
```

### Reject
`Reject` command rejects an `Approval`
#### Command
`cicdctl reject [Approval Name] [Reason]`
#### Examples
```bash
$ cicdctl reject -n default approval-test 'just because'
Rejected Approval default/approval-test
```


### Webhook
`Webhook` command gets webhook server information from an IntegrationConfig
#### Command
`cicdctl webhook [IntegrationConfig Name]`
#### Examples
```bash
$ cicdctl webhook -n default ic-test
Webhook URL     : http://my-webhook.com/webhook/default/ic-test
Webhook Secret  : xxxxxxxxxxxxx
```

