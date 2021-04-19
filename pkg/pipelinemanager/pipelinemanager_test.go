package pipelinemanager

import (
	"github.com/bmizerany/assert"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	"testing"
)

func TestAppendBaseShaToDescription(t *testing.T) {
	desc := "test description"
	sha := git.FakeSha

	appended := appendBaseShaToDescription(desc, sha)
	assert.Equal(t, desc, appended[:len(desc)], "Description")
	assert.Equal(t, statusDescriptionBaseSHAKey+git.FakeSha, appended[len(appended)-len(statusDescriptionBaseSHAKey+git.FakeSha):], "BaseSHA")

	desc = "description which is very longlonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglong"
	msgLen := statusDescriptionMaxLength - len(statusDescriptionBaseSHAKey) - len(git.FakeSha) - len(statusDescriptionEllipse)
	appended = appendBaseShaToDescription(desc, sha)
	assert.Equal(t, desc[:msgLen], appended[:len(desc[:msgLen])], "Description")
	assert.Equal(t, statusDescriptionBaseSHAKey+git.FakeSha, appended[len(appended)-len(statusDescriptionBaseSHAKey+git.FakeSha):], "BaseSHA")

	sha = ""
	appended = appendBaseShaToDescription(desc, sha)
	assert.Equal(t, desc[:statusDescriptionMaxLength], appended, "Description")
}
