package slack

// Message is a slack message
type Message struct {
	Text   string         `json:"text"`
	Blocks []MessageBlock `json:"blocks"`
}

// MessageBlock is a slack message block
type MessageBlock struct {
	Type string    `json:"type"`
	Text BlockText `json:"text"`
}

// BlockText is an actual text
type BlockText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
