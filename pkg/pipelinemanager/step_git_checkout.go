package pipelinemanager

import tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"

func gitCheckout() tektonv1beta1.Step {
	step := tektonv1beta1.Step{}

	step.Name = "git-clone"
	step.Image = "alpine/git"
	step.WorkingDir = DefaultWorkingDir
	step.Script = `#!/bin/sh
set -x
set -e

git config --global user.email "bot@cicd.tmax.io"
git config --global user.name "tmax-cicd-bot"

CHECKOUT_URL="$CI_SERVER_URL/$CI_REPOSITORY"
if [ "$CI_BASE_REF" = "" ]; then
    CHECKOUT_REF="$CI_HEAD_REF"
    CHECKOUT_SHA="$CI_HEAD_SHA"
else
    CHECKOUT_REF="$CI_BASE_REF"
    CHECKOUT_SHA="$CI_BASE_SHA"
fi
git init
git fetch "$CHECKOUT_URL" "$CHECKOUT_REF"
git checkout FETCH_HEAD
if [ "$CI_BASE_REF" != "" ]; then
    git fetch "$CHECKOUT_URL" "$CI_HEAD_REF"
    git merge --no-ff "$CI_HEAD_SHA"
fi
git submodule update --init --recursive
`

	return step
}
