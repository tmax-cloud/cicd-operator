package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	messageTitle = "IntegrationJobNotification"
)

// SendMessage sends webhook payload
func SendMessage(url, message string) error {
	// Generate message payload
	data := &Message{
		Text: messageTitle,
		Blocks: []MessageBlock{{
			Type: "section",
			Text: BlockText{
				Type: "mrkdwn",
				Text: message,
			},
		}},
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("status: %d, error: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
