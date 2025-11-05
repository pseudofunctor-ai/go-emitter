package otel

import (
	"context"
	"fmt"
	"maps"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	t "github.com/pseudofunctor-ai/go-emitter/emitter/types"
)

// OtelBackend implements EmitterBackend for OpenTelemetry metrics
type OtelBackend struct {
	meter metric.Meter

	// Cache instruments to avoid recreating them
	int64Counters     sync.Map // map[string]metric.Int64Counter
	float64Counters   sync.Map // map[string]metric.Float64Counter
	int64Gauges       sync.Map // map[string]metric.Int64Gauge
	float64Gauges     sync.Map // map[string]metric.Float64Gauge
	int64Histograms   sync.Map // map[string]metric.Int64Histogram
	float64Histograms sync.Map // map[string]metric.Float64Histogram
}

// NewOtelBackend creates a new OpenTelemetry backend
func NewOtelBackend(meter metric.Meter) *OtelBackend {
	return &OtelBackend{
		meter: meter,
	}
}

// propsToAttributes converts a property map to OpenTelemetry attributes
// It filters out special properties like _rate, _message, _logLevel
func propsToAttributes(props map[string]interface{}) []attribute.KeyValue {
	p := maps.Clone(props)

	// Remove special properties that aren't meant to be attributes
	delete(p, "_rate")
	delete(p, "_message")
	delete(p, "_logLevel")

	attrs := make([]attribute.KeyValue, 0, len(p))
	for k, v := range p {
		switch val := v.(type) {
		case string:
			attrs = append(attrs, attribute.String(k, val))
		case int:
			attrs = append(attrs, attribute.Int(k, val))
		case int64:
			attrs = append(attrs, attribute.Int64(k, val))
		case float64:
			attrs = append(attrs, attribute.Float64(k, val))
		case bool:
			attrs = append(attrs, attribute.Bool(k, val))
		default:
			// Fallback to string representation
			attrs = append(attrs, attribute.String(k, fmt.Sprintf("%v", val)))
		}
	}
	return attrs
}

// EmitInt implements EmitterBackend.EmitInt
func (b *OtelBackend) EmitInt(ctx context.Context, event string, props map[string]interface{}, value int64, metricType t.MetricType) {
	attrs := propsToAttributes(props)
	opts := metric.WithAttributes(attrs...)

	switch metricType {
	case t.COUNT, t.METER:
		counter, err := b.getOrCreateInt64Counter(event)
		if err != nil {
			return
		}
		counter.Add(ctx, value, opts)

	case t.GAUGE:
		gauge, err := b.getOrCreateInt64Gauge(event)
		if err != nil {
			return
		}
		gauge.Record(ctx, value, opts)

	case t.HISTOGRAM:
		histogram, err := b.getOrCreateInt64Histogram(event)
		if err != nil {
			return
		}
		histogram.Record(ctx, value, opts)

	case t.TIMER:
		// For timer with int64, treat as milliseconds and convert to histogram
		histogram, err := b.getOrCreateFloat64Histogram(event)
		if err != nil {
			return
		}
		histogram.Record(ctx, float64(value)/1000.0, opts) // Convert ms to seconds
	}
}

// EmitFloat implements EmitterBackend.EmitFloat
func (b *OtelBackend) EmitFloat(ctx context.Context, event string, props map[string]interface{}, value float64, metricType t.MetricType) {
	attrs := propsToAttributes(props)
	opts := metric.WithAttributes(attrs...)

	switch metricType {
	case t.COUNT, t.METER:
		counter, err := b.getOrCreateFloat64Counter(event)
		if err != nil {
			return
		}
		counter.Add(ctx, value, opts)

	case t.GAUGE:
		gauge, err := b.getOrCreateFloat64Gauge(event)
		if err != nil {
			return
		}
		gauge.Record(ctx, value, opts)

	case t.HISTOGRAM, t.TIMER:
		histogram, err := b.getOrCreateFloat64Histogram(event)
		if err != nil {
			return
		}
		histogram.Record(ctx, value, opts)
	}
}

// EmitDuration implements EmitterBackend.EmitDuration
func (b *OtelBackend) EmitDuration(ctx context.Context, event string, props map[string]interface{}, value time.Duration, metricType t.MetricType) {
	attrs := propsToAttributes(props)
	opts := metric.WithAttributes(attrs...)

	// Record duration as seconds (float64) in a histogram
	histogram, err := b.getOrCreateFloat64Histogram(event)
	if err != nil {
		return
	}
	histogram.Record(ctx, value.Seconds(), opts)
}

// Instrument cache getters/creators

func (b *OtelBackend) getOrCreateInt64Counter(name string) (metric.Int64Counter, error) {
	if val, ok := b.int64Counters.Load(name); ok {
		return val.(metric.Int64Counter), nil
	}

	counter, err := b.meter.Int64Counter(name)
	if err != nil {
		return nil, err
	}

	b.int64Counters.Store(name, counter)
	return counter, nil
}

func (b *OtelBackend) getOrCreateFloat64Counter(name string) (metric.Float64Counter, error) {
	if val, ok := b.float64Counters.Load(name); ok {
		return val.(metric.Float64Counter), nil
	}

	counter, err := b.meter.Float64Counter(name)
	if err != nil {
		return nil, err
	}

	b.float64Counters.Store(name, counter)
	return counter, nil
}

func (b *OtelBackend) getOrCreateInt64Gauge(name string) (metric.Int64Gauge, error) {
	if val, ok := b.int64Gauges.Load(name); ok {
		return val.(metric.Int64Gauge), nil
	}

	gauge, err := b.meter.Int64Gauge(name)
	if err != nil {
		return nil, err
	}

	b.int64Gauges.Store(name, gauge)
	return gauge, nil
}

func (b *OtelBackend) getOrCreateFloat64Gauge(name string) (metric.Float64Gauge, error) {
	if val, ok := b.float64Gauges.Load(name); ok {
		return val.(metric.Float64Gauge), nil
	}

	gauge, err := b.meter.Float64Gauge(name)
	if err != nil {
		return nil, err
	}

	b.float64Gauges.Store(name, gauge)
	return gauge, nil
}

func (b *OtelBackend) getOrCreateInt64Histogram(name string) (metric.Int64Histogram, error) {
	if val, ok := b.int64Histograms.Load(name); ok {
		return val.(metric.Int64Histogram), nil
	}

	histogram, err := b.meter.Int64Histogram(name)
	if err != nil {
		return nil, err
	}

	b.int64Histograms.Store(name, histogram)
	return histogram, nil
}

func (b *OtelBackend) getOrCreateFloat64Histogram(name string) (metric.Float64Histogram, error) {
	if val, ok := b.float64Histograms.Load(name); ok {
		return val.(metric.Float64Histogram), nil
	}

	histogram, err := b.meter.Float64Histogram(name)
	if err != nil {
		return nil, err
	}

	b.float64Histograms.Store(name, histogram)
	return histogram, nil
}
