package git

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// RequestHTTP requests api call
func RequestHTTP(method string, uri string, header map[string]string, data interface{}) ([]byte, http.Header, error) {
	var jsonBytes []byte
	var err error

	if data != nil {
		jsonBytes, err = json.Marshal(data)
		if err != nil {
			return nil, nil, err
		}
	}

	req, err := http.NewRequest(method, uri, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, nil, err
	}

	for k, v := range header {
		req.Header.Add(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	// Check additional response header
	var newErr error
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		newErr = fmt.Errorf("error requesting api, code %d, msg %s", resp.StatusCode, string(body))
	}

	return body, resp.Header, newErr
}
