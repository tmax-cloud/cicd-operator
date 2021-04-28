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

func TestParseBaseFromDescription(t *testing.T) {
	fullDesc := "Job is running... BaseSHA:2641c89aac959fb804ec6f2a4a22e129f4ac4900"
	sha := ParseBaseFromDescription(fullDesc)
	assert.Equal(t, "2641c89aac959fb804ec6f2a4a22e129f4ac4900", sha)

	fullDesc = "Job is running... BaseSHA:zzzzzzzzzzzzzzzzz"
	sha = ParseBaseFromDescription(fullDesc)
	assert.Equal(t, "", sha)
}
