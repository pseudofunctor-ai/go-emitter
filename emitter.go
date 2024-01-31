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

type magicProps struct {
	hostname string
	filename string
	lineNo   int
	funcName string
}

type Emitter struct {
	registeredEvents  map[string]struct{}
	memoTable         map[string]magicProps
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
	return &Emitter{
		registeredEvents:  make(map[string]struct{}),
		memoTable:         make(map[string]magicProps),
		backends:          backends,
		magicHostname:     false,
		magicFilename:     false,
		magicLineNo:       false,
		magicFuncName:     false,
		callback:          nil,
		hostname_provider: os.Hostname,
	}
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

func (e *Emitter) WithoutMagicProps() *Emitter {
	e.magicHostname = false
	e.magicFilename = false
	e.magicLineNo = false
	e.magicFuncName = false
	return e
}

func (e *Emitter) WithAllMagicProps() *Emitter {
	return e.WithMagicHostname().WithMagicFilename().WithMagicLineNo().WithMagicFuncName()
}

func (e *Emitter) Metric(event string, metricType t.MetricType) t.MetricEmitterFn {
	if _, ok := e.registeredEvents[event]; ok {
		panic(fmt.Sprintf("Event %s already registered", event))
	}

	eCopy := *e
	silentE := (&eCopy).WithoutMagicProps()
	silentE.EmitInt(context.Background(), event, nil, 0, metricType)

	e.registeredEvents[event] = struct{}{}
	return func(ctx context.Context, props map[string]interface{}) {
		e.EmitInt(ctx, event, props, 1, metricType)
	}
}

func (e *Emitter) Log(event string, logfn func(ctx context.Context, event string, props map[string]interface{}, format string, args ...interface{})) t.LogEmitterFn {
	if _, ok := e.registeredEvents[event]; ok {
		panic(fmt.Sprintf("Event %s already registered", event))
	}
	eCopy := *e
	silentE := (&eCopy).WithoutMagicProps()
	silentE.EmitInt(context.Background(), event, nil, 0, t.COUNT)

	e.registeredEvents[event] = struct{}{}
	return func(ctx context.Context, props map[string]interface{}, format string, args ...interface{}) {
		logfn(ctx, event, props, format, args...)
	}
}

func (e *Emitter) addMagicPropsToEvent(ctx context.Context, event string, props map[string]interface{}) map[string]interface{} {
	if props == nil {
		props = make(map[string]interface{}, 5)
	}
	if props["__includes_magic_props"] != nil {
		return props
	}
	p := props
	if e.callback != nil || e.magicFilename || e.magicLineNo || e.magicFuncName || e.magicHostname {
		p = maps.Clone(props)
	} else {
		return p
	}

	// Memoize the magic properties so we don't have to do this work every time
	if details, ok := e.memoTable[event]; ok {
		p["filename"] = details.filename
		p["lineNo"] = details.lineNo
		p["funcName"] = details.funcName
		p["hostname"] = details.hostname
		p["__includes_magic_props"] = true
		return p
	}

	skip := 2

	_, thisFile, _, _ := runtime.Caller(0)
	pc, file, line, ok := runtime.Caller(skip)
	// XXX: For some reason, when testing we skip 2, but elsewhere we seem to need to skip 3
	for ok && file == thisFile {
		skip += 1
		pc, file, line, ok = runtime.Caller(skip)
	}

	if e.magicFilename {
		p["filename"] = file
		p["__includes_magic_props"] = true
	}
	if e.magicLineNo {
		p["lineNo"] = line
		p["__includes_magic_props"] = true
	}
	if e.magicFuncName {
		p["funcName"] = runtime.FuncForPC(pc).Name()
		p["__includes_magic_props"] = true
	}
	if e.magicHostname {
		hostname, err := e.hostname_provider()
		if err == nil {
			p["hostname"] = hostname
			p["__includes_magic_props"] = true
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
	delete(p, "__includes_magic_props")
	for _, backend := range e.backends {
		backend.EmitFloat(ctx, event, p, value, metricType)
	}
}

func (e *Emitter) EmitInt(ctx context.Context, event string, props map[string]interface{}, value int64, metricType t.MetricType) {
	p := e.addMagicPropsToEvent(ctx, event, props)
	delete(p, "__includes_magic_props")
	for _, backend := range e.backends {
		backend.EmitInt(ctx, event, p, value, metricType)
	}
}

func (e *Emitter) EmitDuration(ctx context.Context, event string, props map[string]interface{}, value time.Duration, metricType t.MetricType) {
	p := e.addMagicPropsToEvent(ctx, event, props)
	delete(p, "__includes_magic_props")
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

type CompatAdapter struct {
	emitter   t.ContextLogger
	eventName string
}

func MakeCompatAdapter(eventName string, emitter t.ContextLogger) *CompatAdapter {
	return &CompatAdapter{eventName: eventName, emitter: emitter}
}

func (c *CompatAdapter) propsFromArgs(args []interface{}) map[string]interface{} {
	props := make(map[string]interface{}, len(args)/2)
	if len(args)%2 != 0 {
		panic("Props must be an even number of arguments")
	}
	for i := 0; i < len(args); i += 2 {
		props[args[i].(string)] = args[i+1]
	}
	return props
}

// Implement ContextLoggerCompat so we can use our emitter as a legacy logger
func (c *CompatAdapter) InfoContext(ctx context.Context, msg string, args ...interface{}) {
	c.emitter.InfoContext(ctx, c.eventName, c.propsFromArgs(args), msg)
}

func (c *CompatAdapter) WarnContext(ctx context.Context, msg string, args ...interface{}) {
	c.emitter.WarnContext(ctx, c.eventName, c.propsFromArgs(args), msg)
}

func (c *CompatAdapter) ErrorContext(ctx context.Context, msg string, args ...interface{}) {
	c.emitter.ErrorContext(ctx, c.eventName, c.propsFromArgs(args), msg)
}

func (c *CompatAdapter) FatalContext(ctx context.Context, msg string, args ...interface{}) {
	c.emitter.FatalContext(ctx, c.eventName, c.propsFromArgs(args), msg)
}

func (c *CompatAdapter) DebugContext(ctx context.Context, msg string, args ...interface{}) {
	c.emitter.DebugContext(ctx, c.eventName, c.propsFromArgs(args), msg)
}

func (c *CompatAdapter) TraceContext(ctx context.Context, msg string, args ...interface{}) {
	c.emitter.TraceContext(ctx, c.eventName, c.propsFromArgs(args), msg)
}
