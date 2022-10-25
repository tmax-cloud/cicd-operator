/*
 Copyright 2021 The CI/CD Operator Authors

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package pipelinemanager

import (
	"github.com/stretchr/testify/require"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"testing"
)

func Test_fetchCatalog(t *testing.T) {
	tc := map[string]struct {
		caName string
		caVer  string
		caUrl  string

		errorOccurs  bool
		taskStr      string
		expectedTask *tektonv1beta1.Task
	}{
		"invalidArgument": {
			caName:      "",
			caVer:       "",
			caUrl:       "",
			errorOccurs: true,
		},
		"CatalogWithOutUrl": {
			caName:       "golang-build",
			caVer:        "0.3",
			caUrl:        "",
			taskStr:      "apiVersion: tekton.dev/v1beta1\nkind: Task\nmetadata:\n  name: golang-build\n  labels:\n    app.kubernetes.io/version: \"0.3\"\n  annotations:\n    tekton.dev/pipelines.minVersion: \"0.12.1\"\n    tekton.dev/categories: Build Tools\n    tekton.dev/tags: build-tool\n    tekton.dev/displayName: \"golang build\"\n    tekton.dev/platforms: \"linux/amd64,linux/s390x,linux/ppc64le\"\nspec:\n  description: >-\n    This Task is Golang task to build Go projects.\n\n  params:\n  - name: package\n    description: base package to build in\n  - name: packages\n    description: \"packages to build (default: ./cmd/...)\"\n    default: \"./cmd/...\"\n  - name: version\n    description: golang version to use for builds\n    default: \"latest\"\n  - name: flags\n    description: flags to use for the test command\n    default: -v\n  - name: GOOS\n    description: \"running program's operating system target\"\n    default: linux\n  - name: GOARCH\n    description: \"running program's architecture target\"\n    default: amd64\n  - name: GO111MODULE\n    description: \"value of module support\"\n    default: auto\n  - name: GOCACHE\n    description: \"Go caching directory path\"\n    default: \"\"\n  - name: GOMODCACHE\n    description: \"Go mod caching directory path\"\n    default: \"\"\n  - name: CGO_ENABLED\n    description: \"Toggle cgo tool during Go build. Use value '0' to disable cgo (for static builds).\"\n    default: \"\"\n  - name: GOSUMDB\n    description: \"Go checksum database url. Use value 'off' to disable checksum validation.\"\n    default: \"\"\n  workspaces:\n  - name: source\n  steps:\n  - name: build\n    image: docker.io/library/golang:$(params.version)\n    workingDir: $(workspaces.source.path)\n    script: |\n      if [ ! -e $GOPATH/src/$(params.package)/go.mod ];then\n        SRC_PATH=\"$GOPATH/src/$(params.package)\"\n        mkdir -p $SRC_PATH\n        cp -R \"$(workspaces.source.path)\"/* $SRC_PATH\n        cd $SRC_PATH\n      fi\n      go build $(params.flags) $(params.packages)\n    env:\n    - name: GOOS\n      value: \"$(params.GOOS)\"\n    - name: GOARCH\n      value: \"$(params.GOARCH)\"\n    - name: GO111MODULE\n      value: \"$(params.GO111MODULE)\"\n    - name: GOCACHE\n      value: \"$(params.GOCACHE)\"\n    - name: GOMODCACHE\n      value: \"$(params.GOMODCACHE)\"\n    - name: CGO_ENABLED\n      value: \"$(params.CGO_ENABLED)\"\n    - name: GOSUMDB\n      value: \"$(params.GOSUMDB)\"",
			expectedTask: &tektonv1beta1.Task{},
		},
		"CatalogWithUrl": {
			caName:       "",
			caVer:        "",
			caUrl:        "https://raw.githubusercontent.com/tektoncd/catalog/main/task/golang-build/0.3/golang-build.yaml",
			taskStr:      "apiVersion: tekton.dev/v1beta1\nkind: Task\nmetadata:\n  name: golang-build\n  labels:\n    app.kubernetes.io/version: \"0.3\"\n  annotations:\n    tekton.dev/pipelines.minVersion: \"0.12.1\"\n    tekton.dev/categories: Build Tools\n    tekton.dev/tags: build-tool\n    tekton.dev/displayName: \"golang build\"\n    tekton.dev/platforms: \"linux/amd64,linux/s390x,linux/ppc64le\"\nspec:\n  description: >-\n    This Task is Golang task to build Go projects.\n\n  params:\n  - name: package\n    description: base package to build in\n  - name: packages\n    description: \"packages to build (default: ./cmd/...)\"\n    default: \"./cmd/...\"\n  - name: version\n    description: golang version to use for builds\n    default: \"latest\"\n  - name: flags\n    description: flags to use for the test command\n    default: -v\n  - name: GOOS\n    description: \"running program's operating system target\"\n    default: linux\n  - name: GOARCH\n    description: \"running program's architecture target\"\n    default: amd64\n  - name: GO111MODULE\n    description: \"value of module support\"\n    default: auto\n  - name: GOCACHE\n    description: \"Go caching directory path\"\n    default: \"\"\n  - name: GOMODCACHE\n    description: \"Go mod caching directory path\"\n    default: \"\"\n  - name: CGO_ENABLED\n    description: \"Toggle cgo tool during Go build. Use value '0' to disable cgo (for static builds).\"\n    default: \"\"\n  - name: GOSUMDB\n    description: \"Go checksum database url. Use value 'off' to disable checksum validation.\"\n    default: \"\"\n  workspaces:\n  - name: source\n  steps:\n  - name: build\n    image: docker.io/library/golang:$(params.version)\n    workingDir: $(workspaces.source.path)\n    script: |\n      if [ ! -e $GOPATH/src/$(params.package)/go.mod ];then\n        SRC_PATH=\"$GOPATH/src/$(params.package)\"\n        mkdir -p $SRC_PATH\n        cp -R \"$(workspaces.source.path)\"/* $SRC_PATH\n        cd $SRC_PATH\n      fi\n      go build $(params.flags) $(params.packages)\n    env:\n    - name: GOOS\n      value: \"$(params.GOOS)\"\n    - name: GOARCH\n      value: \"$(params.GOARCH)\"\n    - name: GO111MODULE\n      value: \"$(params.GO111MODULE)\"\n    - name: GOCACHE\n      value: \"$(params.GOCACHE)\"\n    - name: GOMODCACHE\n      value: \"$(params.GOMODCACHE)\"\n    - name: CGO_ENABLED\n      value: \"$(params.CGO_ENABLED)\"\n    - name: GOSUMDB\n      value: \"$(params.GOSUMDB)\"",
			expectedTask: &tektonv1beta1.Task{},
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			taskSpec, err := fetchCatalog(c.caName, c.caVer, c.caUrl)
			if c.errorOccurs {
				require.Error(t, err)
			} else {
				parseYAML(t, c.taskStr, c.expectedTask)
				require.NoError(t, err)
				require.Equal(t, &c.expectedTask.Spec, taskSpec)
			}
		})
	}
}

func parseYAML(t *testing.T, str string, i runtime.Object) {
	if _, _, err := scheme.Codecs.UniversalDeserializer().Decode([]byte(str), nil, i); err != nil {
		t.Fatalf("fail to parse string (%s): %v", str, err)
	}
}
