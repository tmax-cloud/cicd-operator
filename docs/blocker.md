# [WIP] Blocker

Blocker is a helper module for the merge automation.
It literally blocks PullRequests from automatically merged, by setting a commit status `blocker` to the pull request.

After all the **simple merge conditions** (branch, author, and label conditions) are met, it checks the **full merge conditions** (no merge conflict, commit statuses are successful based on the recent SHA of the base branch).
If the test is successful, merger merges the PR automatically to the base branch.
If the test is successful but the test was not performed based on the recent base commit, it triggers the test again.

## Merge Pool
Merge pool is a pool of pull requests satisfies simple conditions for merge.
There are two separated list of PRs, `pending` and `success`. `pending` pool contains PRs with merge conflicts or unmet commit-status-conditions.
PRs in `success` pool are ready to be merged. (i.e., satisfies full merge condition)

## Pool Syncer
Pool syncer synchronizes (caches) pull requests and checks simple conditions of the PR to be merged.
The conditions are `author`, (base)`branch`, `labels`. If all the conditions are satisfied, the PR is added to a merge pool.
Otherwise, the pr is not included in the merge pool.

## Status Syncer
Status syncer checks full merge conditions for the PRs in the merge pool.
Also, status syncer reports `blocker` commit status (e.g., In merge pool, Not mergeable) to every PR, including those who are not in the merge pool.

## Merger
[WIP]
