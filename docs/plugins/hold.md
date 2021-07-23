## `Hold` ChatOps-Plugin

Hold chat-ops plugin makes it possible to hold a pull request from being merged by commenting on the pull request.
Anyone can hold the pull request by commenting `/hold` and anyone can cancel the hold by commenting `/hold cancel`.

Label value can be configured via ConfigMap `blocker-config`'s `mergeBlockLabel`
> **Default Label**  
> ci/hold
