package pipelinemanager

import (
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	gitCheckoutCPUReq = "30m"
	gitCheckoutMemReq = "50Mi"
)

func gitCheckout() tektonv1beta1.Step {
	step := tektonv1beta1.Step{}

	step.Name = "git-clone"
	step.Image = configs.GitImage
	step.WorkingDir = DefaultWorkingDir
	step.Script = `#!/bin/sh
set -x
set -e

git config --global user.email "bot@cicd.tmax.io"
git config --global user.name "tmax-cicd-bot"
git init

CHECKOUT_URL="$CI_SERVER_URL/$CI_REPOSITORY"
CI_HEAD_REF_ARRAY="$CI_HEAD_REF"

if [ "$CI_BASE_REF" = "" ]; then 
  # Push Event
  CHECKOUT_REF="$CI_HEAD_REF"
else 
  # Pull Request Event
  CHECKOUT_REF="$CI_BASE_REF"
fi

git fetch "$CHECKOUT_URL" "$CHECKOUT_REF"
git checkout FETCH_HEAD

if [ "$CI_BASE_REF" != "" ]; then
  # Pull request event
  for ci_head_ref in $CI_HEAD_REF_ARRAY; do 
    git fetch "$CHECKOUT_URL" "$ci_head_ref"
    git merge --no-ff FETCH_HEAD
  done
fi

git submodule update --init --recursive
`
	resources := corev1.ResourceList{
		"cpu":    resource.MustParse(gitCheckoutCPUReq),
		"memory": resource.MustParse(gitCheckoutMemReq),
	}
	step.Resources = corev1.ResourceRequirements{
		Limits:   resources,
		Requests: resources,
	}

	return step
}
