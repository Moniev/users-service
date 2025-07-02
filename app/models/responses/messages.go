package responses

import "errors"

type Message struct {
	Status    string      `json:"status"`
	Message   string      `json:"message"`
	Error     string      `json:"error,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

type MessageInterface interface {
	Valid() error
}

func (m Message) Valid() error {
	if m.Status == "" || m.Message == "" {
		return errors.New("no data provided")
	}

	return nil
}
