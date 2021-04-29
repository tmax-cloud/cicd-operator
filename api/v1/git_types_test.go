package v1

import (
	"github.com/bmizerany/assert"
	"testing"
)

func TestGitRef_GetBranch(t *testing.T) {
	tc := map[string]gitTypeTestCase{
		"non-ref": {Input: "master", ExpectedOutput: "master"},
		"branch":  {Input: "refs/heads/master", ExpectedOutput: "master"},
		"tag":     {Input: "refs/tags/v0.1.1", ExpectedOutput: ""},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, c.ExpectedOutput, c.Input.GetBranch())
		})
	}
}

func TestGitRef_GetTag(t *testing.T) {
	tc := map[string]gitTypeTestCase{
		"non-ref": {Input: "master", ExpectedOutput: "master"},
		"branch":  {Input: "refs/heads/master", ExpectedOutput: ""},
		"tag":     {Input: "refs/tags/v0.1.1", ExpectedOutput: "v0.1.1"},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, c.ExpectedOutput, c.Input.GetTag())
		})
	}
}

type gitTypeTestCase struct {
	Input          GitRef
	ExpectedOutput string
}
