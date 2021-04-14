# Merger

Merger is a merge automation module for PullRequests.

- [Configuring `mergeConfig`](#configuring-mergeconfig)

## Configuring `mergeConfig`
### `method`
`method` field specifies the method to merge the PR. 
> Optional  
> Available values: `squash`, `rebase`, `merge`  
> Default: `merge`

### `commitTemplate`
`commitTemplate` specifies the title template of the merge commit.
> Optional  
> Default: `{{ .Title }}({{ .ID }})`

### `query`
`query` is a selector of PRs to be merged. (i.e., conditions of PRs to be merged)
PRs are searched using the query and merged if all the CI checks are completed.
There are 9 kinds of queries. `labels`, `skipLabels`, `authors`, `skipAuthors`, `branches`, `skipBranches`, `checks`, `optionalChecks`, and `approveRequired`.
