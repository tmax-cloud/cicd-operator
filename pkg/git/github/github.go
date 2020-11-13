package github

import "github.com/tmax-cloud/cicd-operator/pkg/git"

type Client struct {
}

func (c *Client) ParseWebhook(jsonString []byte) (git.Webhook, error) {
	// TODO
	return git.Webhook{}, nil
}

func (c *Client) RegisterWebhook() error {
	// TODO
	return nil
}
