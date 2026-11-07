package progress_test

import (
	"context"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/platform/progress"
)

func TestEmitToContextSink(t *testing.T) {
	var captured []progress.Event
	ctx := progress.WithSink(context.Background(), progress.SinkFunc(func(ev progress.Event) {
		captured = append(captured, ev)
	}))

	progress.Emit(ctx, "scaffold", progress.StatusRunning, "rendering files")
	progress.Emit(ctx, "scaffold", progress.StatusDone, "done")

	if len(captured) != 2 {
		t.Fatalf("expected 2 events, got %d", len(captured))
	}
	if captured[0].Phase != "scaffold" || captured[0].Status != progress.StatusRunning {
		t.Fatalf("unexpected first event: %+v", captured[0])
	}
}
