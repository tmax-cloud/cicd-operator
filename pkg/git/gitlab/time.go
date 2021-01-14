package gitlab

import (
	"encoding/json"
	"time"
)

const (
	timeFormat = "2006-01-02 15:04:05 MST"
)

type gitlabTime struct {
	time.Time `protobuf:"-"`
}

// UnmarshalJSON un-marshals gitlabTime object
func (g *gitlabTime) UnmarshalJSON(b []byte) error {
	if len(b) == 4 && string(b) == "null" {
		g.Time = time.Time{}
		return nil
	}

	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}

	pt, err := time.Parse(timeFormat, str)
	if err != nil {
		return err
	}

	g.Time = pt.Local()
	return nil
}

// MarshalJSON marshals gitlabTime object
func (g *gitlabTime) MarshalJSON() ([]byte, error) {
	buf := make([]byte, 0, len(timeFormat)+2)
	buf = append(buf, '"')
	buf = g.UTC().AppendFormat(buf, timeFormat)
	buf = append(buf, '"')
	return buf, nil
}
