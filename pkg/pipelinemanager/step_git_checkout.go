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
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	gitCheckoutCPUReqDefault = "30m"
	gitCheckoutMemReqDefault = "100Mi"
)

const defaultScript = `#!/bin/sh
set -x
set -e

git config --global user.email "bot@cicd.tmax.io"
git config --global user.name "tmax-cicd-bot"
git init

CHECKOUT_URL="$CI_SERVER_URL/$CI_REPOSITORY"
CI_HEAD_REF_ARRAY="$CI_HEAD_REF"

if [ "$CI_BASE_REF" = "" -a "$CI_HEAD_REF" != "" ]; then 
  # Push Event
  CHECKOUT_REF="$CI_HEAD_REF"
elif [ "$CI_BASE_REF" = "" -a "CI_HEAD_REF" = "" ]; then
  # Commit comment
  CHECKOUT_REF="$CI_HEAD_SHA"
else 
  # Pull Request Event
  CHECKOUT_REF="$CI_BASE_REF"
fi

git -c http.sslVerify=false fetch "$CHECKOUT_URL" "$CHECKOUT_REF"
git -c http.sslVerify=false checkout FETCH_HEAD

if [ "$CI_BASE_REF" != "" ]; then
  # Pull request event
  for ci_head_ref in $CI_HEAD_REF_ARRAY; do 
    git -c http.sslVerify=false fetch "$CHECKOUT_URL" "$ci_head_ref"
    git -c http.sslVerify=false merge --no-ff FETCH_HEAD
  done
fi

git -c http.sslVerify=false submodule update --init --recursive
`

func gitCheckout() tektonv1beta1.Step {
	step := tektonv1beta1.Step{}

	step.Name = "git-clone"
	step.Image = configs.GitImage
	step.WorkingDir = DefaultWorkingDir
	step.Script = defaultScript

	cpuReq, err := resource.ParseQuantity(configs.GitCheckoutStepCPURequest)
	if err != nil {
		cpuReq = resource.MustParse(gitCheckoutCPUReqDefault)
	}
	memReq, err := resource.ParseQuantity(configs.GitCheckoutStepMemRequest)
	if err != nil {
		memReq = resource.MustParse(gitCheckoutMemReqDefault)
	}
	resources := corev1.ResourceList{
		"cpu":    cpuReq,
		"memory": memReq,
	}
	step.Resources = corev1.ResourceRequirements{
		Limits:   resources,
		Requests: resources,
	}

	return step
}
