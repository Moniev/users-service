package responses

type SuccessResponse struct {
	Status    string
	Data      interface{}
	RequestID string
}
