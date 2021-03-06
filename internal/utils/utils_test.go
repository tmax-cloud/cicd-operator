package utils

import (
	"github.com/bmizerany/assert"
	"testing"
)

func TestParseApproversList(t *testing.T) {
	// Success test
	str := `admin@tmax.co.kr=admin@tmax.co.kr,test@tmax.co.kr
test2@tmax.co.kr=test2@tmax.co.kr`
	list, err := ParseApproversList(str)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 3, len(list), "list is not parsed well")
	assert.Equal(t, "admin@tmax.co.kr=admin@tmax.co.kr", list[0], "list is not parsed well")
	assert.Equal(t, "test@tmax.co.kr", list[1], "list is not parsed well")
	assert.Equal(t, "test2@tmax.co.kr=test2@tmax.co.kr", list[2], "list is not parsed well")

	// Fail test
	str = "admin,,ttt"
	list, err = ParseApproversList(str)
	if err == nil {
		for i, l := range list {
			t.Logf("%d : %s", i, l)
		}
		t.Fatal("error not occur")
	}
}

func TestParseEmailFromUsers(t *testing.T) {
	// Include test
	users := []string{
		"aweilfjlwesfj",
		"aweilfjlwesfj=aweiojweio",
		"aweilfjlwesfj=admin@tmax.co.kr",
		"asdij@oisdjf.sdfioj=test@tmax.co.kr",
	}

	tos := ParseEmailFromUsers(users)

	assert.Equal(t, 2, len(tos), "list is not parsed well")
	assert.Equal(t, "admin@tmax.co.kr", tos[0], "list is not parsed well")
	assert.Equal(t, "test@tmax.co.kr", tos[1], "list is not parsed well")
}
