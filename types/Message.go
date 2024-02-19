package types

type Message struct {
	StatusCode int    `json:"status"`
	Message    string `json:"message,omitempty"`
	Video      Video  `json:"video,omitempty"`
}
