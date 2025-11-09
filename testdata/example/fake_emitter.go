package example

import (
	"context"
)

// FakeEmitter has methods with the same names as emitter methods
// but different signatures - these should NOT be detected
type FakeEmitter struct{}

// Count with wrong signature (missing props parameter)
func (f *FakeEmitter) Count(ctx context.Context, name string, value int64) {
	// This should NOT be detected as an emitter call
}

// Gauge with wrong signature (wrong parameter types)
func (f *FakeEmitter) Gauge(name string, value string) {
	// This should NOT be detected as an emitter call
}

// InfoContext with wrong signature (extra parameter)
func (f *FakeEmitter) InfoContext(ctx context.Context, event string, props map[string]interface{}, msg string, extra int) {
	// This should NOT be detected as an emitter call
}

// Time with wrong signature (not a timer)
func (f *FakeEmitter) Time() int64 {
	// This should NOT be detected as a timer call
	return 0
}

func FakeEmitterCalls() {
	ctx := context.Background()
	fake := &FakeEmitter{}

	// These calls should NOT be detected by the generator
	fake.Count(ctx, "fake_count", 42)
	fake.Gauge("fake_gauge", "100")
	fake.InfoContext(ctx, "fake_info", nil, "message", 123)
	fake.Time()
}
