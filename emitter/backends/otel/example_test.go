package otel_test

import (
	"context"
	"fmt"
	"time"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	emit "github.com/pseudofunctor-ai/go-emitter/emitter"
	"github.com/pseudofunctor-ai/go-emitter/emitter/backends/otel"
)

// Example demonstrates basic usage of the OpenTelemetry backend
func Example() {
	ctx := context.Background()

	// Create a MeterProvider with a manual reader (for production, use a real exporter)
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	defer mp.Shutdown(ctx)

	// Create a meter
	meter := mp.Meter("example-app")

	// Create the OpenTelemetry backend
	otelBackend := otel.NewOtelBackend(meter)

	// Create an emitter with the OpenTelemetry backend
	emitter := emit.NewEmitter(otelBackend)

	// Use the emitter to record metrics
	emitter.Count(ctx, "requests.total", map[string]interface{}{
		"method": "GET",
		"status": 200,
	}, 1)

	emitter.Gauge(ctx, "temperature.celsius", map[string]interface{}{
		"location": "datacenter",
	}, 24.5)

	// Record a histogram value
	hist := emitter.Histogram(ctx, "request.size.bytes", map[string]interface{}{
		"endpoint": "/api/users",
	})
	hist.Observe(1024)

	fmt.Println("Metrics recorded successfully")
	// Output: Metrics recorded successfully
}

// ExampleMultipleBackends shows how to use OpenTelemetry alongside other backends
func Example_multipleBackends() {
	ctx := context.Background()

	// Setup OpenTelemetry (for production, use a real exporter)
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	defer mp.Shutdown(ctx)

	meter := mp.Meter("multi-backend-app")
	otelBackend := otel.NewOtelBackend(meter)

	// You can combine multiple backends
	// For example, send to both OpenTelemetry and a custom backend
	emitter := emit.NewEmitter(otelBackend /* , anotherBackend, ... */)

	// All metrics will be sent to all backends
	emitter.Count(ctx, "api.calls", map[string]interface{}{
		"service": "user-service",
	}, 1)

	fmt.Println("Metrics sent to all backends")
	// Output: Metrics sent to all backends
}

// ExampleTimingMetrics demonstrates recording timing information
func Example_timingMetrics() {
	ctx := context.Background()

	// Setup (for production, use a real exporter)
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	defer mp.Shutdown(ctx)

	otelBackend := otel.NewOtelBackend(mp.Meter("timing-app"))
	emitter := emit.NewEmitter(otelBackend)

	// Create a timing emitter
	timer := emit.NewTimingEmitter[int](emitter)

	// Time a function execution
	result := timer.Time(ctx, "db.query.duration", map[string]interface{}{
		"query": "SELECT * FROM users",
	}, func() int {
		time.Sleep(10 * time.Millisecond)
		return 42
	})

	fmt.Printf("Query returned: %d\n", result)
	// Output: Query returned: 42
}
