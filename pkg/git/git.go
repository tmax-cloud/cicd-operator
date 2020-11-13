package git

type Client interface {
	RegisterWebhook() error
	ParseWebhook([]byte) (Webhook, error)
}
