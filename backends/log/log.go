//go:generate mockgen -destination=mocks/mock_log.go -package=mocks "github.com/pseudofunctor-ai/go-emitter/backends/log" LoggerInterface
package log

import (
	"context"
	"fmt"
	"maps"
	"sort"
	"time"

	emitter "github.com/pseudofunctor-ai/go-emitter"
)

type LoggerInterface interface {
	InfoContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
	DebugContext(ctx context.Context, msg string, args ...any)
}

type LogEmitter struct {
	logger LoggerInterface
}

// NewLogEmitter creates a new slog emitter
func NewLogEmitter(logger LoggerInterface) *LogEmitter {
	return &LogEmitter{
		logger: logger,
	}
}

// mapToLogParams converts a map of properties to a slice of alternating keys
// and values, which is required by the slog API
func mapToLogParams(props map[string]interface{}) []any {
	keys := make([]string, 0, len(props))

	for k := range props {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	args := make([]any, 0, len(props)*2)

	for _, k := range keys {
		args = append(args, k, fmt.Sprintf("%v", props[k]))
	}

	return args
}

func isLog(props map[string]interface{}) (msg string, lvl string, isLog bool) {
	message, okMsg := props["_message"]
	level, okLvl := props["_logLevel"]

	if !okMsg || !okLvl {
		return "", "", false
	}

	messageStr, ok := message.(string)
	if !ok {
		return "", "", false
	}

	levelStr, ok := level.(string)
	if !ok {
		return "", "", false
	}

	return messageStr, levelStr, true
}

// log is used internally to log a structured log event, and ignores all other events
func (se *LogEmitter) log(ctx context.Context, event string, props map[string]interface{}) error {
	message, level, ok := isLog(props)
	if !ok {
		return fmt.Errorf("not a log event")
	}

	propsCopy := maps.Clone(props)
	delete(propsCopy, "_message")
	delete(propsCopy, "_logLevel")

	switch level {
	case "INFO":
		se.logger.InfoContext(ctx, message, mapToLogParams(propsCopy)...)
	case "WARN":
		se.logger.WarnContext(ctx, message, mapToLogParams(propsCopy)...)
	case "ERROR":
		se.logger.ErrorContext(ctx, message, mapToLogParams(propsCopy)...)
	case "FATAL":
		se.logger.ErrorContext(ctx, message, mapToLogParams(propsCopy)...)
	case "DEBUG":
		se.logger.DebugContext(ctx, message, mapToLogParams(propsCopy)...)
	case "TRACE":
		se.logger.DebugContext(ctx, message, mapToLogParams(propsCopy)...)
	}

	return nil
}

// EmitFloat satisfies the EmitterBackend interface and for this backend logs the event as a structured log
func (se *LogEmitter) EmitFloat(ctx context.Context, event string, props map[string]interface{}, value float64, t emitter.MetricType) {
	se.log(ctx, event, props)
}

// EmitInt satisfies the EmitterBackend interface and for this backend logs the event as a structured log
func (se *LogEmitter) EmitInt(ctx context.Context, event string, props map[string]interface{}, value int64, t emitter.MetricType) {
	se.log(ctx, event, props)
}

func (se *LogEmitter) EmitDuration(ctx context.Context, event string, props map[string]interface{}, value time.Duration, t emitter.MetricType) {
	se.log(ctx, event, props)
}
