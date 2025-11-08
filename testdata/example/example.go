package example

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	//"github.com/pseudofunctor-ai/go-emitter/emitter"
	"github.com/pseudofunctor-ai/go-emitter/emitter/types"
	"github.com/pseudofunctor-ai/go-emitter/testdata/example/di"
	"github.com/pseudofunctor-ai/go-emitter/testdata/example/subemitter_example"
)

var em types.CombinedEmitter //*emitter.Emitter

func init() {
  deps := di.NewDependencies()
  em = deps.Emitter
  subemitter_example.SubEmitterExample(deps)
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
// Define the callbacks first (these should NOT generate callsite entries)
var (
	decoratedMetric = em.Metric("decorated_metric", types.COUNT)
	decoratedLog    = em.Log("decorated_log", em.InfofContext)
)

func DecoratedFunctions() {
	ctx := context.Background()
	// These MetricFnCallsite/LogFnCallsite calls SHOULD generate the callsite entries
	em.MetricFnCallsite(decoratedMetric)(ctx, nil)
	em.LogFnCallsite(decoratedLog)(ctx, map[string]interface{}{"level": "info"}, "This is a %s", "test")
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

// Test indirect decoration with property keys
var bloomFilterReset = em.MetricWithProps("bloom_filter_reset", types.COUNT, []string{"density", "service_count"})
var criticalConfabulation = em.LogWithProps("critical_confabulation", em.ErrorfContext, []string{"confabulacity"})

func IndirectlyDecoratedFunctionsWithProps() {
	ctx := context.Background()

	// Apply decorators HERE to mark THIS location as the call site
	// This is the pattern for use in wrappers or callbacks
	decoratedBloomFilter := em.MetricFnCallsite(bloomFilterReset)
	decoratedBloomFilter(ctx, map[string]interface{}(nil))

	decoratedConfabulation := em.LogFnCallsite(criticalConfabulation)
	decoratedConfabulation(ctx, map[string]interface{}(nil), "Critically excessive confabulation from %s", "192.168.1.1")
}

// FakeLogger is NOT an emitter - it just happens to have methods with the same names
type FakeLogger struct{}

func (FakeLogger) Info(msg string)                                                              {}
func (FakeLogger) Count(ctx context.Context, name string, props map[string]interface{}, n int) {}
func (FakeLogger) Gauge(ctx context.Context, name string, props map[string]interface{}, v float64) {
}

// NonEmitterCalls demonstrates calls that should NOT be picked up by the generator
func NonEmitterCalls() {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// These fmt.Errorf calls should NOT be detected as emitter calls
	// even though they have method names that match (Error, Errorf, etc.)
	_ = fmt.Errorf("error with code: %s", "E404")
	_ = fmt.Errorf("database error: %v", "connection failed")

	// These slog calls should NOT be detected
	// even though they have similar method signatures (Info, Error, Debug, etc.)
	logger.Info("user logged in", "user", "alice", "action", "login")
	logger.Error("database error", "code", "E500")
	logger.Debug("debug message", "key", "value")
	logger.InfoContext(ctx, "context log", "user", "bob")
	logger.ErrorContext(ctx, "context error", "code", "E404")

	// FakeLogger calls that share method names with emitters
	fake := FakeLogger{}
	fake.Info("should not be picked up")
	fake.Count(ctx, "not_an_emitter_count", nil, 1)
	fake.Gauge(ctx, "not_an_emitter_gauge", nil, 42.0)
}

func EmittersDefinedInAnotherCompilationUnit() {
  tem := someEmitters(em)
  tem.event1(context.Background(), map[string]interface{}{"prop1": "value1", "prop2": "value2"}, "Event 1 occurred")
  tem.event2(context.Background(), map[string]interface{}{"metric1": "value1", "metric2": "value2"}, 100)
  tem.event3.Time(context.Background(), "timed_event", map[string]interface{}{"Hello": "World"}, func() int { time.Sleep(1 * time.Millisecond); return 0 })
}

func EmittersInArrays() {
  ctx := context.Background()
  slice := emittersInSlice(em)

  // Access via index
  slice[0](ctx, nil)
  slice[1](ctx, map[string]interface{}{"size": 100, "duration": 50})
  slice[2](ctx, map[string]interface{}{"value": 42})
}

func EmittersInMaps() {
  ctx := context.Background()
  m := emittersInMap(em)

  // Access via key
  m["info"](ctx, nil, "Info message: %s", "test")
  m["warn"](ctx, map[string]interface{}{"severity": "high", "component": "auth"}, "Warning: %s", "failure")
  m["error"](ctx, nil, "Error occurred: %v", "timeout")
}
