package responses

import "errors"

type SuccessResponse struct {
	Status    string      `json:"status"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

type SuccessInterface interface {
	Valid() error
}

func (r *SuccessResponse) Valid() error {
	if r.Status == "" {
		return errors.New("no status provided")
	}

	return nil
}
