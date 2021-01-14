package gitlab

import (
	"encoding/json"
	"github.com/bmizerany/assert"
	"testing"
	"time"
)

type testTimeStruct struct {
	Time *gitlabTime `json:"time"`
}

func TestGitlabTime_UnmarshalJSON(t *testing.T) {
	buf := []byte("{\"time\": \"2021-02-03 07:13:23 UTC\"}")

	tt := &testTimeStruct{}
	if err := json.Unmarshal(buf, tt); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "2021-02-03 07:13:23 +0000 UTC", tt.Time.UTC().String())
}

func TestGitlabTime_MarshalJSON(t *testing.T) {
	targetTime, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", "2021-02-03 07:13:23 +0000 UTC")
	if err != nil {
		t.Fatal(err)
	}
	tt := &testTimeStruct{
		Time: &gitlabTime{
			Time: targetTime,
		},
	}

	result, err := json.Marshal(tt)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "{\"time\":\"2021-02-03 07:13:23 UTC\"}", string(result))
}
