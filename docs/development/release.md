## Release Procedure

### Before Release
1. Update `Makefile`
    - Update `VERSION ?= <to-be version>`
2. Re-generate manifests
    - `make manifest`
3. Bump the version via Pull Request  
   (Refer to [the link](https://github.com/tmax-cloud/cicd-operator/pull/152))

### Release
1. Make GitHub Release with version name `v[0-9]+.[0-9]+.[0-9]+`
2. Refer to the previous release notes!

### After Release
1. Wait until post-release CI pipelines succeed
