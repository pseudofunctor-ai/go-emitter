package emitter

import (
	"context"
	"fmt"
	"maps"
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

type PassthroughHistogram struct {
	ctx     context.Context
	emitter *Emitter
	props   map[string]interface{}
	event   string
}

func (h *PassthroughHistogram) Observe(value float64) {
	h.emitter.EmitFloat(h.ctx, h.event, h.props, value, HISTOGRAM)
}

type MetricsTimer[T any] interface {
	Time(ctx context.Context, event string, props map[string]interface{}, fn func() T) T
}

type EmitterBackend interface {
	EmitDuration(ctx context.Context, event string, props map[string]interface{}, value time.Duration, t MetricType)
	EmitFloat(ctx context.Context, event string, props map[string]interface{}, value float64, t MetricType)
	EmitInt(ctx context.Context, event string, props map[string]interface{}, value int64, t MetricType)
}

type Emitter struct {
	backends []EmitterBackend
}

type TimingEmitter[T any] struct {
	emitter *Emitter
}

func NewTimingEmitter[T any](emitter *Emitter) TimingEmitter[T] {
	return TimingEmitter[T]{emitter: emitter}
}

func NewEmitter(backends ...EmitterBackend) *Emitter {
	return &Emitter{backends: backends}
}

func (e *Emitter) EmitFloat(ctx context.Context, event string, props map[string]interface{}, value float64, t MetricType) {
	for _, backend := range e.backends {
		backend.EmitFloat(ctx, event, props, value, t)
	}
}

func (e *Emitter) EmitInt(ctx context.Context, event string, props map[string]interface{}, value int64, t MetricType) {
	for _, backend := range e.backends {
		backend.EmitInt(ctx, event, props, value, t)
	}
}

func (e *Emitter) EmitDuration(ctx context.Context, event string, props map[string]interface{}, value time.Duration, t MetricType) {
	for _, backend := range e.backends {
		backend.EmitDuration(ctx, event, props, value, t)
	}
}

func (e *Emitter) Count(ctx context.Context, event string, props map[string]interface{}, value int64) {
	e.EmitInt(ctx, event, props, value, COUNT)
}

func (e *Emitter) Gauge(ctx context.Context, event string, props map[string]interface{}, value float64) {
	e.EmitFloat(ctx, event, props, value, GAUGE)
}

func (e *Emitter) Histogram(ctx context.Context, event string, props map[string]interface{}) Histogram {
	return &PassthroughHistogram{ctx: ctx, emitter: e, event: event, props: props}
}

func (e TimingEmitter[T]) Time(ctx context.Context, event string, props map[string]interface{}, fn func() T) T {
	start := time.Now()
	r := fn()
	elapsed := time.Since(start)
	e.emitter.EmitInt(ctx, event, props, elapsed.Milliseconds(), TIMER)
	return r
}

func (e *Emitter) Meter(ctx context.Context, event string, props map[string]interface{}, value int64) {
	e.EmitInt(ctx, event, props, value, METER)
}

func (e *Emitter) Set(ctx context.Context, event string, props map[string]interface{}, value int64) {
	e.EmitInt(ctx, event, props, value, SET)
}

func (e *Emitter) Event(ctx context.Context, event string, props map[string]interface{}) {
	e.EmitInt(ctx, event, props, 1, EVENT)
}

func (e *Emitter) Info(event string, props map[string]interface{}, msg string) {
	e.InfoContext(context.Background(), event, props, msg)
}

func (e *Emitter) Warn(event string, props map[string]interface{}, msg string) {
	e.WarnContext(context.Background(), event, props, msg)
}

func (e *Emitter) Error(event string, props map[string]interface{}, msg string) {
	e.ErrorContext(context.Background(), event, props, msg)
}

func (e *Emitter) Fatal(event string, props map[string]interface{}, msg string) {
	e.FatalContext(context.Background(), event, props, msg)
}

func (e *Emitter) Debug(event string, props map[string]interface{}, msg string) {
	e.DebugContext(context.Background(), event, props, msg)
}

func (e *Emitter) Trace(event string, props map[string]interface{}, msg string) {
	e.TraceContext(context.Background(), event, props, msg)
}

func (e *Emitter) Infof(event string, props map[string]interface{}, format string, args ...interface{}) {
	e.InfofContext(context.Background(), event, props, format, args...)
}

func (e *Emitter) Warnf(event string, props map[string]interface{}, format string, args ...interface{}) {
	e.WarnfContext(context.Background(), event, props, format, args...)
}

func (e *Emitter) Errorf(event string, props map[string]interface{}, format string, args ...interface{}) {
	e.ErrorfContext(context.Background(), event, props, format, args...)
}

func (e *Emitter) Fatalf(event string, props map[string]interface{}, format string, args ...interface{}) {
	e.FatalfContext(context.Background(), event, props, format, args...)
}

func (e *Emitter) Debugf(event string, props map[string]interface{}, format string, args ...interface{}) {
	e.DebugfContext(context.Background(), event, props, format, args...)
}

func (e *Emitter) Tracef(event string, props map[string]interface{}, format string, args ...interface{}) {
	e.TracefContext(context.Background(), event, props, format, args...)
}

func (e *Emitter) InfoContext(ctx context.Context, event string, props map[string]interface{}, msg string) {
	updatedProps := maps.Clone(props)
	updatedProps["_message"] = msg
	updatedProps["_logLevel"] = "INFO"
	e.EmitInt(ctx, event, updatedProps, 1, COUNT)
}

func (e *Emitter) WarnContext(ctx context.Context, event string, props map[string]interface{}, msg string) {
	updatedProps := maps.Clone(props)
	updatedProps["_message"] = msg
	updatedProps["_logLevel"] = "WARN"
	e.EmitInt(ctx, event, updatedProps, 1, COUNT)
}

func (e *Emitter) ErrorContext(ctx context.Context, event string, props map[string]interface{}, msg string) {
	updatedProps := maps.Clone(props)
	updatedProps["_message"] = msg
	updatedProps["_logLevel"] = "ERROR"
	e.EmitInt(ctx, event, updatedProps, 1, COUNT)
}

func (e *Emitter) FatalContext(ctx context.Context, event string, props map[string]interface{}, msg string) {
	updatedProps := maps.Clone(props)
	updatedProps["_message"] = msg
	updatedProps["_logLevel"] = "FATAL"
	e.EmitInt(ctx, event, updatedProps, 1, COUNT)
}

func (e *Emitter) DebugContext(ctx context.Context, event string, props map[string]interface{}, msg string) {
	updatedProps := maps.Clone(props)
	updatedProps["_message"] = msg
	updatedProps["_logLevel"] = "DEBUG"
	e.EmitInt(ctx, event, updatedProps, 1, COUNT)
}

func (e *Emitter) TraceContext(ctx context.Context, event string, props map[string]interface{}, msg string) {
	updatedProps := maps.Clone(props)
	updatedProps["_message"] = msg
	updatedProps["_logLevel"] = "TRACE"
	e.EmitInt(ctx, event, updatedProps, 1, COUNT)
}

func (e *Emitter) InfofContext(ctx context.Context, event string, props map[string]interface{}, format string, args ...interface{}) {
	e.InfoContext(ctx, event, props, fmt.Sprintf(format, args...))
}

func (e *Emitter) WarnfContext(ctx context.Context, event string, props map[string]interface{}, format string, args ...interface{}) {
	e.WarnContext(ctx, event, props, fmt.Sprintf(format, args...))
}

func (e *Emitter) ErrorfContext(ctx context.Context, event string, props map[string]interface{}, format string, args ...interface{}) {
	e.ErrorContext(ctx, event, props, fmt.Sprintf(format, args...))
}

func (e *Emitter) FatalfContext(ctx context.Context, event string, props map[string]interface{}, format string, args ...interface{}) {
	e.FatalContext(ctx, event, props, fmt.Sprintf(format, args...))
}

func (e *Emitter) DebugfContext(ctx context.Context, event string, props map[string]interface{}, format string, args ...interface{}) {
	e.DebugContext(ctx, event, props, fmt.Sprintf(format, args...))
}

func (e *Emitter) TracefContext(ctx context.Context, event string, props map[string]interface{}, format string, args ...interface{}) {
	e.TraceContext(ctx, event, props, fmt.Sprintf(format, args...))
}
