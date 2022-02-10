# Current Operator version
VERSION ?= v0.4.2-release2
REGISTRY ?= tmaxcloudck

# Image URL to use all building/pushing image targets
IMG_CONTROLLER ?= $(REGISTRY)/cicd-operator:$(VERSION)
IMG_BLOCKER ?= $(REGISTRY)/cicd-blocker:$(VERSION)
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: controller cicdctl

# Run tests
test: test-crd test-gen test-verify test-unit test-lint test-coverage

# Build controller binary
controller: generate fmt vet
	go build -o bin/controller cmd/controller/main.go

# Build cicdctl binary
cicdctl:
	go build -o bin/cicdctl cmd/cicdctl/main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go $(RUN_ARGS)

run-cicdctl:
	go run cmd/cicdctl/main.go $(RUN_ARGS)

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=cicd-manager-role webhook paths="./..." output:crd:artifacts:config=config/crd
	./hack/release-manifest.sh $(VERSION)

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
.PHONY: docker-build
docker-build: docker-build-controller docker-build-blocker

docker-build-controller:
	docker build . -f build/controller/Dockerfile -t ${IMG_CONTROLLER}

docker-build-blocker:
	docker build . -f build/blocker/Dockerfile -t ${IMG_BLOCKER}

# Push the docker image
.PHONY: docker-push
docker-push: docker-push-controller docker-push-blocker

docker-push-controller:
	docker push ${IMG_CONTROLLER}

docker-push-blocker:
	docker push ${IMG_BLOCKER}

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.1 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

# Custom targets for CI/CD operator
.PHONY: test-gen test-crd test-verify test-lint test-unit

# Test if zz_generated.deepcopy.go file is generated
test-gen: save-sha-gen generate compare-sha-gen

# Test if crd yaml files are generated
test-crd: save-sha-crd manifests compare-sha-crd

# Verify if go.sum is valid
test-verify: save-sha-mod verify compare-sha-mod

# Test code lint
test-lint:
	golangci-lint run ./... -v

# Unit test
test-unit:
	go test -v ./...

# Check coverage
test-coverage:
	go test -v -coverpkg=./... -coverprofile=profile.cov.tmp ./...
	cat profile.cov.tmp | grep -v "_generated.deepcopy.go" > profile.cov
	go tool cover -func profile.cov
	rm -f profile.cov profile.cov.tmp

save-sha-gen:
	$(eval GENSHA=$(shell sha512sum api/v1/zz_generated.deepcopy.go))

compare-sha-gen:
	$(eval GENSHA_AFTER=$(shell sha512sum api/v1/zz_generated.deepcopy.go))
	@if [ "${GENSHA_AFTER}" = "${GENSHA}" ]; then echo "zz_generated.deepcopy.go is not changed"; else echo "zz_generated.deepcopy.go file is changed"; exit 1; fi

save-sha-crd:
	$(eval CRDSHA1=$(shell sha512sum config/crd/cicd.tmax.io_integrationconfigs.yaml))
	$(eval CRDSHA2=$(shell sha512sum config/crd/cicd.tmax.io_integrationjobs.yaml))
	$(eval CRDSHA3=$(shell sha512sum config/crd/cicd.tmax.io_approvals.yaml))
	$(eval CRDSHA4=$(shell sha512sum config/release.yaml))

compare-sha-crd:
	$(eval CRDSHA1_AFTER=$(shell sha512sum config/crd/cicd.tmax.io_integrationconfigs.yaml))
	$(eval CRDSHA2_AFTER=$(shell sha512sum config/crd/cicd.tmax.io_integrationjobs.yaml))
	$(eval CRDSHA3_AFTER=$(shell sha512sum config/crd/cicd.tmax.io_approvals.yaml))
	$(eval CRDSHA4_AFTER=$(shell sha512sum config/release.yaml))
	@if [ "${CRDSHA1_AFTER}" = "${CRDSHA1}" ]; then echo "cicd.tmax.io_integrationconfigs.yaml is not changed"; else echo "cicd.tmax.io_integrationconfigs.yaml file is changed"; exit 1; fi
	@if [ "${CRDSHA2_AFTER}" = "${CRDSHA2}" ]; then echo "cicd.tmax.io_integrationjobs.yaml is not changed"; else echo "cicd.tmax.io_integrationjobs.yaml file is changed"; exit 1; fi
	@if [ "${CRDSHA3_AFTER}" = "${CRDSHA3}" ]; then echo "cicd.tmax.io_approvals.yaml is not changed"; else echo "cicd.tmax.io_approvals.yaml file is changed"; exit 1; fi
	@if [ "${CRDSHA4_AFTER}" = "${CRDSHA4}" ]; then echo "config/release.yaml is not changed"; else echo "config/release.yaml file is changed"; exit 1; fi

save-sha-mod:
	$(eval MODSHA=$(shell sha512sum go.mod))
	$(eval SUMSHA=$(shell sha512sum go.sum))

verify:
	go mod verify

compare-sha-mod:
	$(eval MODSHA_AFTER=$(shell sha512sum go.mod))
	$(eval SUMSHA_AFTER=$(shell sha512sum go.sum))
	@if [ "${MODSHA_AFTER}" = "${MODSHA}" ]; then echo "go.mod is not changed"; else echo "go.mod file is changed"; exit 1; fi
	@if [ "${SUMSHA_AFTER}" = "${SUMSHA}" ]; then echo "go.sum is not changed"; else echo "go.sum file is changed"; exit 1; fi
