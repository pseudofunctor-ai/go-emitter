package emitter

import (
	"context"
	"fmt"
	"maps"
	"os"
	"runtime"
	"time"

	t "github.com/pseudofunctor-ai/go-emitter/types"
)

type PassthroughHistogram struct {
	ctx     context.Context
	emitter *Emitter
	props   map[string]interface{}
	event   string
}

func (h *PassthroughHistogram) Observe(value float64) {
	h.emitter.EmitFloat(h.ctx, h.event, h.props, value, t.HISTOGRAM)
}

type Emitter struct {
	callback          func(context.Context, string, map[string]interface{})
	hostname_provider func() (string, error)
	backends          []t.EmitterBackend
	magicHostname     bool
	magicFilename     bool
	magicLineNo       bool
	magicFuncName     bool
}

type TimingEmitter[T any] struct {
	emitter *Emitter
}

func NewTimingEmitter[T any](emitter *Emitter) TimingEmitter[T] {
	return TimingEmitter[T]{emitter: emitter}
}

func NewEmitter(backends ...t.EmitterBackend) *Emitter {
	return &Emitter{backends: backends, magicHostname: false, magicFilename: false, magicLineNo: false, magicFuncName: false, callback: nil, hostname_provider: os.Hostname}
}

func (e *Emitter) WithCallback(callback func(context.Context, string, map[string]interface{})) *Emitter {
	e.callback = callback
	return e
}

func (e *Emitter) WithHostnameProvider(hostname_provider func() (string, error)) *Emitter {
	e.hostname_provider = hostname_provider
	return e
}

func (e *Emitter) WithBackend(backend t.EmitterBackend) *Emitter {
	e.backends = append(e.backends, backend)
	return e
}

func (e *Emitter) WithMagicHostname() *Emitter {
	e.magicHostname = true
	return e
}

func (e *Emitter) WithMagicFilename() *Emitter {
	e.magicFilename = true
	return e
}

func (e *Emitter) WithMagicLineNo() *Emitter {
	e.magicLineNo = true
	return e
}

func (e *Emitter) WithMagicFuncName() *Emitter {
	e.magicFuncName = true
	return e
}

func (e *Emitter) WithAllMagicProps() *Emitter {
	return e.WithMagicHostname().WithMagicFilename().WithMagicLineNo().WithMagicFuncName()
}

func (e *Emitter) addMagicPropsToEvent(ctx context.Context, event string, props map[string]interface{}) map[string]interface{} {
	if props == nil {
		props = make(map[string]interface{}, 5)
	}
	p := props
	if e.callback != nil || e.magicFilename || e.magicLineNo || e.magicFuncName || e.magicHostname {
		p = maps.Clone(props)
	} else {
		return p
	}

	_, thisFile, _, _ := runtime.Caller(0)
	pc, file, line, ok := runtime.Caller(2)
	if file == thisFile || !ok {
		return p
	}

	if e.magicFilename {
		p["filename"] = file
	}
	if e.magicLineNo {
		p["lineNo"] = line
	}
	if e.magicFuncName {
		p["funcName"] = runtime.FuncForPC(pc).Name()
	}
	if e.magicHostname {
		hostname, err := e.hostname_provider()
		if err == nil {
			p["hostname"] = hostname
		}
	}

	if e.callback != nil {
		e.callback(ctx, event, p)
	}

	return p
}

// Implement EmitterBackend in case we want to stack emitters
func (e *Emitter) EmitFloat(ctx context.Context, event string, props map[string]interface{}, value float64, metricType t.MetricType) {
	p := e.addMagicPropsToEvent(ctx, event, props)
	for _, backend := range e.backends {
		backend.EmitFloat(ctx, event, p, value, metricType)
	}
}

func (e *Emitter) EmitInt(ctx context.Context, event string, props map[string]interface{}, value int64, metricType t.MetricType) {
	p := e.addMagicPropsToEvent(ctx, event, props)
	for _, backend := range e.backends {
		backend.EmitInt(ctx, event, p, value, metricType)
	}
}

func (e *Emitter) EmitDuration(ctx context.Context, event string, props map[string]interface{}, value time.Duration, metricType t.MetricType) {
	p := e.addMagicPropsToEvent(ctx, event, props)
	for _, backend := range e.backends {
		backend.EmitDuration(ctx, event, p, value, metricType)
	}
}

// Implement MetricsEmitter
func (e *Emitter) Count(ctx context.Context, event string, props map[string]interface{}, value int64) {
	e.EmitInt(ctx, event, props, value, t.COUNT)
}

func (e *Emitter) Gauge(ctx context.Context, event string, props map[string]interface{}, value float64) {
	e.EmitFloat(ctx, event, props, value, t.GAUGE)
}

func (e *Emitter) Histogram(ctx context.Context, event string, props map[string]interface{}) t.Histogram {
	return &PassthroughHistogram{ctx: ctx, emitter: e, event: event, props: props}
}

func (e TimingEmitter[T]) Time(ctx context.Context, event string, props map[string]interface{}, fn func() T) T {
	start := time.Now()
	r := fn()
	elapsed := time.Since(start)
	e.emitter.EmitInt(ctx, event, props, elapsed.Milliseconds(), t.TIMER)
	return r
}

func (e *Emitter) Meter(ctx context.Context, event string, props map[string]interface{}, value int64) {
	e.EmitInt(ctx, event, props, value, t.METER)
}

func (e *Emitter) Set(ctx context.Context, event string, props map[string]interface{}, value int64) {
	e.EmitInt(ctx, event, props, value, t.SET)
}

func (e *Emitter) Event(ctx context.Context, event string, props map[string]interface{}) {
	e.EmitInt(ctx, event, props, 1, t.EVENT)
}

// Implement SimpleLogger
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

// Implement FormatLogger
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

// Implement SimpleContextLogger
func (e *Emitter) InfoContext(ctx context.Context, event string, props map[string]interface{}, msg string) {
	updatedProps := e.addMagicPropsToEvent(ctx, event, props)
	updatedProps["_message"] = msg
	updatedProps["_logLevel"] = "INFO"
	e.EmitInt(ctx, event, updatedProps, 1, t.COUNT)
}

func (e *Emitter) WarnContext(ctx context.Context, event string, props map[string]interface{}, msg string) {
	updatedProps := e.addMagicPropsToEvent(ctx, event, props)
	updatedProps["_message"] = msg
	updatedProps["_logLevel"] = "WARN"
	e.EmitInt(ctx, event, updatedProps, 1, t.COUNT)
}

func (e *Emitter) ErrorContext(ctx context.Context, event string, props map[string]interface{}, msg string) {
	updatedProps := e.addMagicPropsToEvent(ctx, event, props)
	updatedProps["_message"] = msg
	updatedProps["_logLevel"] = "ERROR"
	e.EmitInt(ctx, event, updatedProps, 1, t.COUNT)
}

func (e *Emitter) FatalContext(ctx context.Context, event string, props map[string]interface{}, msg string) {
	updatedProps := e.addMagicPropsToEvent(ctx, event, props)
	updatedProps["_message"] = msg
	updatedProps["_logLevel"] = "FATAL"
	e.EmitInt(ctx, event, updatedProps, 1, t.COUNT)
}

func (e *Emitter) DebugContext(ctx context.Context, event string, props map[string]interface{}, msg string) {
	updatedProps := e.addMagicPropsToEvent(ctx, event, props)
	updatedProps["_message"] = msg
	updatedProps["_logLevel"] = "DEBUG"
	e.EmitInt(ctx, event, updatedProps, 1, t.COUNT)
}

func (e *Emitter) TraceContext(ctx context.Context, event string, props map[string]interface{}, msg string) {
	updatedProps := e.addMagicPropsToEvent(ctx, event, props)
	updatedProps["_message"] = msg
	updatedProps["_logLevel"] = "TRACE"
	e.EmitInt(ctx, event, updatedProps, 1, t.COUNT)
}

// Implement FormatContextLogger
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
