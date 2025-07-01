package responses

type ErrorResponse struct {
	Status    string
	Error     string
	RequestID string
}
