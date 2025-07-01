package responses

type ProgressReporter struct {
	channel chan<- Message
}

type Reporter interface {
	Processing(message string)
	Success(message string, data any)
	Error(message string, err error)
}

var _ Reporter = (*ProgressReporter)(nil)

func NewProgressReporter(progressChan chan<- Message) *ProgressReporter {
	return &ProgressReporter{channel: progressChan}
}

func (r *ProgressReporter) Processing(message string) {
	r.channel <- Message{Status: "processing", Message: message}
}

func (r *ProgressReporter) Success(message string, data any) {
	r.channel <- Message{Status: "success", Message: message, Data: data}
}

func (r *ProgressReporter) Error(message string, err error) {
	errorString := ""
	if err != nil {
		errorString = err.Error()
	}

	r.channel <- Message{Status: "error", Message: message, Error: errorString}
}
