package responses

type Message struct {
	Status    string      `json:"status"`
	Message   string      `json:"message"`
	Error     string      `json:"error,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}
