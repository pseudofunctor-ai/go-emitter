//go:generate mockgen -destination=mocks/mock_types.go -package "mocks" "github.com/pseudofunctor-ai/go-emitter/emitter/types" SimpleLogger,SimpleLoggerFmt,ContextLogger,ContextLoggerFmt,MetricsEmitter,EmitterBackend
package types

import (
	"context"
	"time"
)

type CallSiteDetails struct {
	Filename     string
	LineNo       int
	FuncName     string
	Package      string
	PropertyKeys []string
	MetricType   string
}

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

type ContextLoggerCompat interface {
	InfoContext(ctx context.Context, event string, msg string, args ...interface{})
	WarnContext(ctx context.Context, event string, msg string, args ...interface{})
	ErrorContext(ctx context.Context, event string, msg string, args ...interface{})
	FatalContext(ctx context.Context, event string, msg string, args ...interface{})
	DebugContext(ctx context.Context, event string, msg string, args ...interface{})
	TraceContext(ctx context.Context, event string, msg string, args ...interface{})
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

type (
	MetricEmitterFn func(ctx context.Context, props map[string]interface{}, value ...interface{})
	LogEmitterFn    func(ctx context.Context, props map[string]interface{}, format string, args ...interface{})
)

type IntEmitter interface {
  EmitInt(ctx context.Context, event string, props map[string]interface{}, value int64, metricType MetricType)
}

type FloatEmitter interface {
  EmitFloat(ctx context.Context, event string, props map[string]interface{}, value float64, metricType MetricType)
}

type DurationEmitter interface {
  EmitDuration(ctx context.Context, event string, props map[string]interface{}, value time.Duration, metricType MetricType)
}

type CombinedEmitter interface {
	SimpleLogger
	SimpleLoggerFmt
	ContextLogger
	ContextLoggerFmt
	MetricsEmitter
  DurationEmitter
  FloatEmitter
  IntEmitter
	Metric(event string, metricType MetricType) MetricEmitterFn
	MetricWithProps(event string, metricType MetricType, propKeys []string) MetricEmitterFn
	Log(event string, logfn func(ctx context.Context, event string, props map[string]interface{}, format string, args ...interface{})) LogEmitterFn
	LogWithProps(event string, logfn func(ctx context.Context, event string, props map[string]interface{}, format string, args ...interface{}), propKeys []string) LogEmitterFn
  NewSubEmitter() CombinedEmitter
  WithStaticMetadata(staticData map[string]CallSiteDetails) CombinedEmitter
  MetricFnCallsite(fn MetricEmitterFn) MetricEmitterFn
  LogFnCallsite(fn LogEmitterFn) LogEmitterFn
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

func (m MetricType) String() string {
	switch m {
	case COUNT:
		return "COUNT"
	case GAUGE:
		return "GAUGE"
	case HISTOGRAM:
		return "HISTOGRAM"
	case TIMER:
		return "TIMER"
	case METER:
		return "METER"
	case SET:
		return "SET"
	case EVENT:
		return "EVENT"
	default:
		return "UNKNOWN"
	}
}

type Histogram interface {
	Observe(value float64)
}

type MetricsEmitter interface {
	Count(ctx context.Context, event string, props map[string]interface{}, value int64)
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

// MetricManifestEntry represents a single metric in the manifest
type MetricManifestEntry struct {
	Name         string     `json:"name"`
	MetricType   MetricType `json:"metric_type"`
	TypeString   string     `json:"type_string"`
	PropertyKeys []string   `json:"property_keys,omitempty"`
}
