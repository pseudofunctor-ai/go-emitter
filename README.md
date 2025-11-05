# go-emitter

A unified metrics and logging emitter library for Go that provides a single interface for emitting events to multiple backends simultaneously. Automatically captures call site details for rich observability context.

## Why use go-emitter?

**Unified Interface**: Write your instrumentation code once, emit to multiple backends (StatsD, structured logs, custom backends) without vendor lock-in.

**Rich Context**: Automatically capture and attach call site details (filename, line number, function name, package) to every event for better debugging and analysis.

**Metric Discovery**: Pre-register all metrics at startup to generate a complete manifest of every metric your application can emit - crucial for setting up alerts before rare error conditions occur.

**Flexible Call Site Capture**: Choose between dynamic runtime capture (simple, works everywhere) or static build-time generation (more precise for callbacks/wrappers, no runtime overhead).

**Flexible Architecture**: Plugin-based backend system makes it easy to add custom adapters for your observability stack.

## Installation

```bash
go get github.com/pseudofunctor-ai/go-emitter/emitter
```

To install the code generator:

```bash
go install github.com/pseudofunctor-ai/go-emitter@latest
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "log/slog"

    "github.com/pseudofunctor-ai/go-emitter/emitter"
    "github.com/pseudofunctor-ai/go-emitter/emitter/backends/log"
)

func main() {
    // Create a logger backend
    logger := slog.Default()
    logBackend := log.NewLogEmitter(logger)

    // Create an emitter with the backend
    em := emitter.NewEmitter(logBackend)

    ctx := context.Background()

    // Emit a log message
    em.Info("user_login", map[string]interface{}{"user_id": "alice"}, "User logged in successfully")

    // Emit a metric
    em.Count(ctx, "api_requests", map[string]interface{}{"endpoint": "/users"}, 1)
}
```

### Registered Metrics and Logs

Pre-register your events at startup to create a complete manifest of all metrics your application can emit. This ensures you can set up alerts and dashboards before rare events occur (like that error that happens once a month).

```go
package main

import (
    "context"
    "encoding/json"
    "net/http"

    "github.com/pseudofunctor-ai/go-emitter/emitter"
    "github.com/pseudofunctor-ai/go-emitter/emitter/types"
)

var em *emitter.Emitter

// Pre-registered metrics - all declared at startup
var (
    userLoginMetric     = em.Metric("user_login", types.COUNT)
    userLoginFailure    = em.Metric("user_login_failure", types.COUNT)  // Rare event!
    requestDuration     = em.Metric("request_duration", types.TIMER)
    activeUsers         = em.Metric("active_users", types.GAUGE)
    dbConnectionError   = em.Metric("db_connection_error", types.COUNT) // Hopefully rare!
)

// Pre-registered logs
var (
    auditLog = em.Log("audit_log", em.InfofContext)
    errorLog = em.Log("error_log", em.ErrorfContext)
)

func handleLogin(ctx context.Context, userID string) {
    userLoginMetric(ctx, map[string]interface{}{"user_id": userID})
    auditLog(ctx, map[string]interface{}{"action": "login"}, "User %s logged in", userID)
}

// Export metric manifest for monitoring setup
func metricManifestHandler(w http.ResponseWriter, r *http.Request) {
    manifest := []string{
        "user_login", "user_login_failure", "request_duration",
        "active_users", "db_connection_error",
    }
    json.NewEncoder(w).Encode(manifest)
}
```

### Call Site Decorators

Mark specific locations as the call site when using callbacks or wrappers - this is where static generation really shines:

```go
var cacheHitMetric = em.Metric("cache_hit", types.COUNT)

// In your cache wrapper - mark THIS location as the call site
func cacheGet(ctx context.Context, key string) (interface{}, bool) {
    recordCacheHit := em.MetricFnCallsite(cacheHitMetric)

    if value, ok := cache.Get(key); ok {
        recordCacheHit(ctx, map[string]interface{}{"key": key, "hit": true})
        return value, true
    }
    return nil, false
}
```

## Static Call Site Generation

Generate call site details at build time for better accuracy with callbacks/wrappers and zero runtime overhead:

```bash
# Generate call site map
emitter-gen -o emitter_callsites.go -var EmitterCallSites -package myapp ./path/to/package
```

```go
// Use the static provider
func init() {
    provider := emitter.NewStaticCallsiteProvider(EmitterCallSites)
    em = emitter.NewEmitter().WithCallsiteProvider(provider)
}
```

Integrate with your build:

```makefile
generate:
	emitter-gen -o internal/metrics/emitter_callsites.go -var CallSites -package metrics ./internal/metrics
```

## Dependency Injection Example

```go
package main

import (
    "context"
    "log/slog"
    "os"

    "github.com/cactus/go-statsd-client/v5/statsd"
    "github.com/pseudofunctor-ai/go-emitter/emitter"
    logbackend "github.com/pseudofunctor-ai/go-emitter/emitter/backends/log"
    statsdbackend "github.com/pseudofunctor-ai/go-emitter/emitter/backends/statsd"
    "github.com/pseudofunctor-ai/go-emitter/emitter/types"
)

// DI container for your application
type DI struct {
    Config    *Config
    Emitter   *emitter.Emitter
    DBService *DBService
    // ... other services
}

type Config struct {
    StatsDAddr string
    LogLevel   slog.Level
    Hostname   string
}

// Metrics - pre-registered at startup for discoverability
var (
    HTTPRequestCount    emitter.MetricEmitterFn
    HTTPRequestDuration emitter.MetricEmitterFn
    DBQueryDuration     emitter.MetricEmitterFn
    DBConnectionError   emitter.MetricEmitterFn  // Rare but critical!
    CacheHit            emitter.MetricEmitterFn
)

// Logs
var (
    AuditLog emitter.LogEmitterFn
    ErrorLog emitter.LogEmitterFn
)

func NewDI(cfg *Config) (*DI, error) {
    // Create backends
    statsdClient, _ := statsd.NewClientWithConfig(&statsd.ClientConfig{
        Address: cfg.StatsDAddr,
        Prefix:  "myapp",
    })

    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: cfg.LogLevel,
    }))

    // Create emitter with multiple backends
    em := emitter.NewEmitter(
        statsdbackend.NewStatsdBackend(statsdClient),
        logbackend.NewLogEmitter(logger),
    ).WithHostnameProvider(func() string { return cfg.Hostname })

    // Pre-register all metrics and logs
    HTTPRequestCount = em.Metric("http_request_count", types.COUNT)
    HTTPRequestDuration = em.Metric("http_request_duration", types.TIMER)
    DBQueryDuration = em.Metric("db_query_duration", types.TIMER)
    DBConnectionError = em.Metric("db_connection_error", types.COUNT)
    CacheHit = em.Metric("cache_hit", types.COUNT)

    AuditLog = em.Log("audit_log", em.InfofContext)
    ErrorLog = em.Log("error_log", em.ErrorfContext)

    return &DI{
        Config:    cfg,
        Emitter:   em,
        DBService: NewDBService(em),
    }, nil
}

// DBService shows timing with the built-in timer
type DBService struct {
    em    *emitter.Emitter
    timer *emitter.TimingEmitter[emitter.MetricEmitterFn]
}

func NewDBService(em *emitter.Emitter) *DBService {
    return &DBService{
        em:    em,
        timer: emitter.NewTimingEmitter(DBQueryDuration),
    }
}

func (s *DBService) Query(ctx context.Context, query string) error {
    // Automatic timing using the built-in timer
    return s.timer.Time(ctx, map[string]interface{}{"query_type": "SELECT"}, func() error {
        // Your query logic here
        return db.Execute(query)
    })
}

// HTTPHandler shows manual instrumentation
type HTTPHandler struct {
    di *DI
}

func (h *HTTPHandler) HandleRequest(ctx context.Context, endpoint, userID string) {
    props := map[string]interface{}{
        "endpoint": endpoint,
        "user_id":  userID,
    }

    HTTPRequestCount(ctx, props)
    AuditLog(ctx, props, "User %s accessed %s", userID, endpoint)
}

func main() {
    di, _ := NewDI(&Config{
        StatsDAddr: "localhost:8125",
        LogLevel:   slog.LevelInfo,
        Hostname:   "app-01",
    })

    handler := &HTTPHandler{di: di}
    handler.HandleRequest(context.Background(), "/api/users", "alice")
}
```

## Backends

### Built-in Backends

- **`backends/statsd`**: Emit metrics to StatsD-compatible servers (DataDog, etc.)
- **`backends/log`**: Emit to structured loggers (slog-compatible)
- **`backends/dummy`**: In-memory backend for testing

### Custom Backends

Implement the `EmitterBackend` interface:

```go
type EmitterBackend interface {
    EmitInt(ctx context.Context, event string, props map[string]interface{}, value int, metricType MetricType) error
    EmitFloat(ctx context.Context, event string, props map[string]interface{}, value float64, metricType MetricType) error
    EmitDuration(ctx context.Context, event string, props map[string]interface{}, value time.Duration, metricType MetricType) error
}
```

## Configuration

### Builder Pattern

```go
em := emitter.NewEmitter(backend1, backend2).
    WithCallsiteProvider(staticProvider).
    WithHostnameProvider(func() string { return "server-01" })
```

### Properties

Properties are key-value pairs attached to events. Special properties (prefixed with `_`) control backend behavior:

- `_message`: Log message (required for log backends)
- `_logLevel`: Log level (INFO, ERROR, WARN, DEBUG, TRACE, FATAL)
- `_rate`: Sample rate for metrics (StatsD)

Call site details are automatically added:
- `callsite_filename`: Source file path
- `callsite_lineno`: Line number
- `callsite_funcname`: Function name
- `callsite_package`: Package path
- `hostname`: Machine hostname

## Testing

Run all tests:

```bash
go test ./...
```

Run generator tests:

```bash
go test -v
```

Use the dummy backend for testing:

```go
import "github.com/pseudofunctor-ai/go-emitter/emitter/backends/dummy"

func TestMyCode(t *testing.T) {
    dummyBackend := dummy.NewDummyBackend()
    em := emitter.NewEmitter(dummyBackend)

    // Your code here

    // Assert events were emitted
    events := dummyBackend.GetEvents()
    if len(events["my_event"]) != 1 {
        t.Error("Expected event not emitted")
    }
}
```

## Best Practices

1. **Pre-register all events**: Declare metrics at startup to create a complete manifest - critical for setting up alerts for rare events
2. **Mark call sites for wrappers**: Use `*FnCallsite()` decorators in middleware and callbacks for accurate call site capture
3. **Consider static generation**: Use `emitter-gen` for better call site accuracy in callbacks and zero runtime overhead
4. **Consistent naming**: Use snake_case for event names
5. **Use context**: Pass `context.Context` to enable distributed tracing integration
6. **Structured properties**: Use maps for rich, queryable event data
7. **Avoid property conflicts**: Don't use properties starting with `_` or `callsite_`

## License

MIT

## Contributing

Contributions welcome! Please ensure all tests pass before submitting PRs:

```bash
go test ./...
go build ./...
```
