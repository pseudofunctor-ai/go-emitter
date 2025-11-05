# go-emitter

A unified metrics and logging emitter library for Go that provides a single interface for emitting events to multiple backends simultaneously. Supports both dynamic (runtime) and static (build-time) call site capture for rich observability context.

## Why use go-emitter?

**Unified Interface**: Write your instrumentation code once, emit to multiple backends (StatsD, structured logs, custom backends) without vendor lock-in.

**Rich Context**: Automatically capture and attach call site details (filename, line number, function name, package) to every event for better debugging and analysis.

**Performance**: Choose between dynamic runtime call site capture or zero-overhead static generation using the `emitter-gen` tool.

**Type-Safe Registration**: Pre-register metrics and logs with strongly-typed emitter functions to avoid typos and ensure consistency.

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

Pre-register your events for type safety and consistency:

```go
package main

import (
    "context"

    "github.com/pseudofunctor-ai/go-emitter/emitter"
    "github.com/pseudofunctor-ai/go-emitter/emitter/types"
)

var em *emitter.Emitter

// Pre-registered metrics
var (
    userLoginMetric  = em.Metric("user_login", types.COUNT)
    requestDuration  = em.Metric("request_duration", types.TIMER)
    activeUsers      = em.Metric("active_users", types.GAUGE)
)

// Pre-registered logs
var (
    auditLog = em.Log("audit_log", em.InfofContext)
    errorLog = em.Log("error_log", em.ErrorfContext)
)

func handleLogin(ctx context.Context, userID string) {
    // Use the registered metric
    userLoginMetric(ctx, map[string]interface{}{"user_id": userID})

    // Use the registered log
    auditLog(ctx, map[string]interface{}{"action": "login"}, "User %s logged in", userID)
}
```

### Call Site Decorators

Mark specific locations as the call site for better troubleshooting (useful for wrappers, callbacks, middleware):

```go
package main

import (
    "context"

    "github.com/pseudofunctor-ai/go-emitter/emitter"
    "github.com/pseudofunctor-ai/go-emitter/emitter/types"
)

var em *emitter.Emitter

// Define the metric
var cacheHitMetric = em.Metric("cache_hit", types.COUNT)

// In your cache wrapper function
func cacheGet(ctx context.Context, key string) (interface{}, bool) {
    // Mark THIS location as the call site for troubleshooting
    recordCacheHit := em.MetricFnCallsite(cacheHitMetric)

    if value, ok := cache.Get(key); ok {
        recordCacheHit(ctx, map[string]interface{}{"key": key, "hit": true})
        return value, true
    }

    return nil, false
}
```

## Static Call Site Generation

For production deployments, eliminate runtime overhead by generating call site details at build time:

### 1. Generate call site map

```bash
emitter-gen -o emitter_callsites.go -var EmitterCallSites -package myapp ./path/to/package
```

### 2. Use the static provider

```go
package main

import (
    "github.com/pseudofunctor-ai/go-emitter/emitter"
)

func init() {
    // Use the generated call site map
    provider := emitter.NewStaticCallsiteProvider(EmitterCallSites)
    em = emitter.NewEmitter().WithCallsiteProvider(provider)
}
```

### 3. Integrate with your build

Add to your `Makefile` or build script:

```makefile
.PHONY: generate
generate:
	go generate ./...
	emitter-gen -o internal/metrics/emitter_callsites.go -var CallSiteDetails -package metrics ./internal/metrics
```

## Dependency Injection Example

Here's a complete example showing best practices with dependency injection, multiple backends, and configuration:

```go
package main

import (
    "context"
    "log/slog"
    "os"
    "time"

    "github.com/cactus/go-statsd-client/v5/statsd"
    "github.com/pseudofunctor-ai/go-emitter/emitter"
    "github.com/pseudofunctor-ai/go-emitter/emitter/backends/log"
    logbackend "github.com/pseudofunctor-ai/go-emitter/emitter/backends/log"
    statsdbackend "github.com/pseudofunctor-ai/go-emitter/emitter/backends/statsd"
    "github.com/pseudofunctor-ai/go-emitter/emitter/types"
)

// Config holds application configuration
type Config struct {
    StatsDAddr string
    LogLevel   slog.Level
    Hostname   string
}

// ObservabilityClient provides metrics and logging
type ObservabilityClient struct {
    emitter *emitter.Emitter

    // Pre-registered metrics
    HTTPRequestCount    emitter.MetricEmitterFn
    HTTPRequestDuration emitter.MetricEmitterFn
    DBQueryDuration     emitter.MetricEmitterFn
    ActiveConnections   emitter.MetricEmitterFn

    // Pre-registered logs
    AuditLog      emitter.LogEmitterFn
    ErrorLog      emitter.LogEmitterFn
    SecurityLog   emitter.LogEmitterFn
    PerformanceLog emitter.LogEmitterFn
}

// NewObservabilityClient creates a new observability client with configured backends
func NewObservabilityClient(cfg Config) (*ObservabilityClient, error) {
    // Create backends
    backends := []types.EmitterBackend{}

    // 1. StatsD backend for metrics
    if cfg.StatsDAddr != "" {
        statsdClient, err := statsd.NewClientWithConfig(&statsd.ClientConfig{
            Address: cfg.StatsDAddr,
            Prefix:  "myapp",
        })
        if err != nil {
            return nil, err
        }
        backends = append(backends, statsdbackend.NewStatsdBackend(statsdClient))
    }

    // 2. Structured logging backend
    logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: cfg.LogLevel,
    })
    logger := slog.New(logHandler)
    backends = append(backends, logbackend.NewLogEmitter(logger))

    // Create the emitter with all backends
    em := emitter.NewEmitter(backends...)

    // Configure with hostname provider
    em = em.WithHostnameProvider(func() string {
        return cfg.Hostname
    })

    // In production, use static call site provider for better performance
    // em = em.WithCallsiteProvider(emitter.NewStaticCallsiteProvider(EmitterCallSites))

    // Pre-register all metrics and logs
    oc := &ObservabilityClient{
        emitter: em,

        // Metrics
        HTTPRequestCount:    em.Metric("http_request_count", types.COUNT),
        HTTPRequestDuration: em.Metric("http_request_duration", types.TIMER),
        DBQueryDuration:     em.Metric("db_query_duration", types.TIMER),
        ActiveConnections:   em.Metric("active_connections", types.GAUGE),

        // Logs
        AuditLog:       em.Log("audit_log", em.InfofContext),
        ErrorLog:       em.Log("error_log", em.ErrorfContext),
        SecurityLog:    em.Log("security_log", em.WarnfContext),
        PerformanceLog: em.Log("performance_log", em.DebugfContext),
    }

    return oc, nil
}

// HTTPHandler demonstrates using the observability client in a service
type HTTPHandler struct {
    oc *ObservabilityClient
}

func NewHTTPHandler(oc *ObservabilityClient) *HTTPHandler {
    return &HTTPHandler{oc: oc}
}

func (h *HTTPHandler) HandleRequest(ctx context.Context, endpoint, userID string) error {
    start := time.Now()

    // Track request count
    h.oc.HTTPRequestCount(ctx, map[string]interface{}{
        "endpoint": endpoint,
        "user_id":  userID,
    })

    // Audit log
    h.oc.AuditLog(ctx, map[string]interface{}{
        "endpoint": endpoint,
        "user_id":  userID,
    }, "User %s accessed %s", userID, endpoint)

    // Simulate some work
    time.Sleep(100 * time.Millisecond)

    // Track duration
    duration := time.Since(start)
    h.oc.HTTPRequestDuration(ctx, map[string]interface{}{
        "endpoint": endpoint,
    })

    // Performance log
    h.oc.PerformanceLog(ctx, map[string]interface{}{
        "endpoint": endpoint,
        "duration_ms": duration.Milliseconds(),
    }, "Request to %s took %dms", endpoint, duration.Milliseconds())

    return nil
}

// DBService demonstrates database instrumentation
type DBService struct {
    oc *ObservabilityClient
}

func NewDBService(oc *ObservabilityClient) *DBService {
    return &DBService{oc: oc}
}

func (s *DBService) Query(ctx context.Context, query string) error {
    // Use call site decorator to mark THIS location for troubleshooting
    recordQuery := s.oc.emitter.MetricFnCallsite(s.oc.DBQueryDuration)

    start := time.Now()

    // Simulate query execution
    time.Sleep(50 * time.Millisecond)

    duration := time.Since(start)
    recordQuery(ctx, map[string]interface{}{
        "query_type": "SELECT",
        "duration_ms": duration.Milliseconds(),
    })

    return nil
}

// ConnectionPool demonstrates gauge metrics
type ConnectionPool struct {
    oc *ObservabilityClient
}

func NewConnectionPool(oc *ObservabilityClient) *ConnectionPool {
    return &ConnectionPool{oc: oc}
}

func (p *ConnectionPool) RecordActiveConnections(ctx context.Context, count int) {
    p.oc.ActiveConnections(ctx, map[string]interface{}{
        "pool": "main",
        "count": count,
    })
}

func main() {
    // Load configuration
    cfg := Config{
        StatsDAddr: "localhost:8125",
        LogLevel:   slog.LevelInfo,
        Hostname:   "app-server-01",
    }

    // Initialize observability
    oc, err := NewObservabilityClient(cfg)
    if err != nil {
        panic(err)
    }

    // Initialize services with dependency injection
    httpHandler := NewHTTPHandler(oc)
    dbService := NewDBService(oc)
    connPool := NewConnectionPool(oc)

    ctx := context.Background()

    // Use the services
    httpHandler.HandleRequest(ctx, "/api/users", "alice")
    dbService.Query(ctx, "SELECT * FROM users WHERE id = ?")
    connPool.RecordActiveConnections(ctx, 42)
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

1. **Pre-register events**: Use `Metric()` and `Log()` to create strongly-typed emitter functions
2. **Use static generation in production**: Eliminate runtime overhead with `emitter-gen`
3. **Mark call sites for wrappers**: Use `*FnCallsite()` decorators in middleware and callbacks
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
