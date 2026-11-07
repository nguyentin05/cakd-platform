package progress

import "context"

// Status describes the lifecycle state of a progress event.
type Status string

const (
	StatusRunning Status = "running"
	StatusDone    Status = "done"
	StatusWarn    Status = "warn"
	StatusError   Status = "error"
)

// Event represents a structured progress notification from a use case or adapter.
type Event struct {
	Phase   string `json:"phase"`
	Status  Status `json:"status"`
	Message string `json:"message"`
}

// Sink consumes progress events emitted during long-running operations.
type Sink interface {
	Emit(Event)
}

// SinkFunc adapts a function into a Sink.
type SinkFunc func(Event)

// Emit calls fn(ev).
func (fn SinkFunc) Emit(ev Event) {
	fn(ev)
}

type sinkKey struct{}

// WithSink attaches a progress Sink to ctx.
func WithSink(ctx context.Context, sink Sink) context.Context {
	return context.WithValue(ctx, sinkKey{}, sink)
}

// Emit sends a progress event to the Sink stored in ctx, if present.
func Emit(ctx context.Context, phase string, status Status, message string) {
	if ctx == nil {
		return
	}
	if sink, ok := ctx.Value(sinkKey{}).(Sink); ok && sink != nil {
		sink.Emit(Event{
			Phase:   phase,
			Status:  status,
			Message: message,
		})
	}
}
