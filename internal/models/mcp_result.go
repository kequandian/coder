package models

type Content struct {
	Type string `json:"type"` // Must be "text"
	// The text content of the message.
	Text string `json:"text"`
}

type MCPResult struct {
	Content []*Content `json:"content"`
}
