package infrastructure

type EventListener struct {
}

type EventListenerInterface interface {
}

var _ EventListenerInterface = (*EventListener)(nil)
