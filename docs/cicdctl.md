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
cicdctl run pre -n default ic-test --head-branch test --base-branch master

# Running postSubmit jobs
cicdctl run post -n default ic-test --branch master
```

### Approve
`Approve` command approves an `Approval`
#### Command
`cicdctl approve [Approval Name] [Reason]`
#### Examples
```bash
cicdctl approve -n default approval-test 'just because'
```

### Reject
`Reject` command rejects an `Approval`
#### Command
`cicdctl reject [Approval Name] [Reason]`
#### Examples
```bash
cicdctl reject -n default approval-test 'just because'
```

