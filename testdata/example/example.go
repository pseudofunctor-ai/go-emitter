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

	// Metric calls with props
	em.Count(ctx, "direct_count", map[string]interface{}{"environment": "prod", "region": "us-west"}, 1)
	em.Gauge(ctx, "direct_gauge", map[string]interface{}{"host": "server1"}, 42.5)

	// Logger calls with props
	em.InfoContext(ctx, "direct_info_log", map[string]interface{}{"user": "alice", "action": "login"}, "info message")
	em.ErrorContext(ctx, "direct_error_log", map[string]interface{}{"code": "E500"}, "error message")
	em.DebugfContext(ctx, "direct_debugf_log", nil, "debug %s", "message")
}

// Test registered metrics
var (
	userLoginMetric  = em.MetricWithProps("user_login_metric", types.COUNT, []string{"user_id", "success"})
	requestDuration  = em.MetricWithProps("request_duration", types.TIMER, []string{"endpoint", "method"})
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
	auditLog = em.LogWithProps("audit_log", em.InfofContext, []string{"action", "resource", "user"})
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
