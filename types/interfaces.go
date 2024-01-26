//go:generate mockgen -destination=mocks/mock_types.go -package "mocks" "github.com/pseudofunctor-ai/go-emitter/types" SimpleLogger,SimpleLoggerFmt,ContextLogger,ContextLoggerFmt,MetricsEmitter,EmitterBackend
package types

import (
	"context"
	"time"
)

type SimpleLogger interface {
	Info(event string, props map[string]interface{}, msg string)
	Warn(event string, props map[string]interface{}, msg string)
	Error(event string, props map[string]interface{}, msg string)
	Fatal(event string, props map[string]interface{}, msg string)
	Debug(event string, props map[string]interface{}, msg string)
	Trace(event string, props map[string]interface{}, msg string)
}

type SimpleLoggerFmt interface {
	Infof(event string, props map[string]interface{}, format string, args ...interface{})
	Warnf(event string, props map[string]interface{}, format string, args ...interface{})
	Errorf(event string, props map[string]interface{}, format string, args ...interface{})
	Fatalf(event string, props map[string]interface{}, format string, args ...interface{})
	Debugf(event string, props map[string]interface{}, format string, args ...interface{})
	Tracef(event string, props map[string]interface{}, format string, args ...interface{})
}

type ContextLogger interface {
	InfoContext(ctx context.Context, event string, props map[string]interface{}, msg string)
	WarnContext(ctx context.Context, event string, props map[string]interface{}, msg string)
	ErrorContext(ctx context.Context, event string, props map[string]interface{}, msg string)
	FatalContext(ctx context.Context, event string, props map[string]interface{}, msg string)
	DebugContext(ctx context.Context, event string, props map[string]interface{}, msg string)
	TraceContext(ctx context.Context, event string, props map[string]interface{}, msg string)
}

type ContextLoggerFmt interface {
	InfofContext(ctx context.Context, event string, props map[string]interface{}, format string, args ...interface{})
	WarnfContext(ctx context.Context, event string, props map[string]interface{}, format string, args ...interface{})
	ErrorfContext(ctx context.Context, event string, props map[string]interface{}, format string, args ...interface{})
	FatalfContext(ctx context.Context, event string, props map[string]interface{}, format string, args ...interface{})
	DebugfContext(ctx context.Context, event string, props map[string]interface{}, format string, args ...interface{})
	TracefContext(ctx context.Context, event string, props map[string]interface{}, format string, args ...interface{})
}

type MetricType int

const (
	COUNT MetricType = iota
	GAUGE
	HISTOGRAM
	TIMER
	METER
	SET
	EVENT
)

type Histogram interface {
	Observe(value float64)
}

type MetricsEmitter interface {
	Count(ctx context.Context, event string, props map[string]interface{}, value float64)
	Gauge(ctx context.Context, event string, props map[string]interface{}, value float64)
	Histogram(ctx context.Context, event string, props map[string]interface{}) Histogram
}

type MetricsTimer[T any] interface {
	Time(ctx context.Context, event string, props map[string]interface{}, fn func() T) T
}

type EmitterBackend interface {
	EmitDuration(ctx context.Context, event string, props map[string]interface{}, value time.Duration, metricType MetricType)
	EmitFloat(ctx context.Context, event string, props map[string]interface{}, value float64, metricType MetricType)
	EmitInt(ctx context.Context, event string, props map[string]interface{}, value int64, metricType MetricType)
}
