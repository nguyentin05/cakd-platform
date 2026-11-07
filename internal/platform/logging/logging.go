package logging

import (
	"io"
	"log/slog"
)

// NewLogger creates a structured slog.Logger writing to w.
func NewLogger(w io.Writer, jsonFormat bool) *slog.Logger {
	if w == nil {
		w = io.Discard
	}
	if jsonFormat {
		return slog.New(slog.NewJSONHandler(w, nil))
	}
	return slog.New(slog.NewTextHandler(w, nil))
}
