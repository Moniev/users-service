package responses

import "errors"

type ErrorResponse struct {
	Status    string `json:"status"`
	Error     string `json:"error"`
	RequestID string `json:"request_id,omitempty"`
}

type ErrorResponseInterface interface {
	Valid() error
}

func (m ErrorResponse) Valid() error {
	if m.Status == "" || m.Error == "" {
		return errors.New("no data provided")
	}

	return nil
}
