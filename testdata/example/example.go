package example

import (
	"context"

	"github.com/pseudofunctor-ai/go-emitter/emitter"
	"github.com/pseudofunctor-ai/go-emitter/emitter/types"
)

var em *emitter.Emitter

func init() {
	em = emitter.NewEmitter()
}

// Test direct method calls
func DirectCalls() {
	ctx := context.Background()

	// Metric calls
	em.Count(ctx, "direct_count", nil, 1)
	em.Gauge(ctx, "direct_gauge", nil, 42.5)

	// Logger calls
	em.InfoContext(ctx, "direct_info_log", nil, "info message")
	em.ErrorContext(ctx, "direct_error_log", nil, "error message")
	em.DebugfContext(ctx, "direct_debugf_log", nil, "debug %s", "message")
}

// Test registered metrics
var (
	userLoginMetric  = em.Metric("user_login_metric", types.COUNT)
	requestDuration  = em.Metric("request_duration", types.TIMER)
	activeUsers      = em.Metric("active_users", types.GAUGE)
)

func RegisteredMetrics() {
	ctx := context.Background()
	userLoginMetric(ctx, map[string]interface{}{"user_id": "123"})
	requestDuration(ctx, nil)
	activeUsers(ctx, map[string]interface{}{"region": "us-east"})
}

// Test registered logs
var (
	auditLog = em.Log("audit_log", em.InfofContext)
	errorLog = em.Log("error_log", em.ErrorfContext)
)

func RegisteredLogs() {
	ctx := context.Background()
	auditLog(ctx, map[string]interface{}{"action": "delete"}, "User %s deleted resource %s", "alice", "file.txt")
	errorLog(ctx, nil, "Database connection failed: %v", "timeout")
}

// Test callsite decorators - inline decoration
var (
	decoratedMetric = em.MetricFnCallsite(em.Metric("decorated_metric", types.COUNT))
	decoratedLog    = em.LogFnCallsite(em.Log("decorated_log", em.InfofContext))
)

func DecoratedFunctions() {
	ctx := context.Background()
	decoratedMetric(ctx, nil)
	decoratedLog(ctx, map[string]interface{}{"level": "info"}, "This is a %s", "test")
}

// Test indirect decoration - define symbols separately, decorate at point of use
var cacheHitMetric = em.Metric("cache_hit_metric", types.COUNT)
var authFailureLog = em.Log("auth_failure_log", em.ErrorfContext)

func IndirectlyDecoratedFunctions() {
	ctx := context.Background()

	// Apply decorators HERE to mark THIS location as the call site
	// This is the pattern for use in wrappers or callbacks
	decoratedCacheHit := em.MetricFnCallsite(cacheHitMetric)
	decoratedCacheHit(ctx, map[string]interface{}{"key": "user:123"})

	decoratedAuthFailure := em.LogFnCallsite(authFailureLog)
	decoratedAuthFailure(ctx, map[string]interface{}{"user": "alice"}, "Failed login attempt from %s", "192.168.1.1")
}
