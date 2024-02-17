package types

type Message struct {
	StatusCode int      `json:"status"`
	Message    string   `json:"message"`
	Queries    []string `json:"queries",omitempty`
}
