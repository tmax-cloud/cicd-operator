# [WIP] Blocker

Blocker is a helper module for the merge automation.

It literally blocks PullRequests from automatically merged, by setting a commit status `blocker` to the pull request.

After all the merge conditions are met (e.g., branch, author, and label conditions), it checks if all the commit status are successful based on the recent SHA of the base branch.

If the test is successful, it merges automatically to the base branch.

If the test is successful but it was not performed based on the recent base commit, it triggers the test again.
