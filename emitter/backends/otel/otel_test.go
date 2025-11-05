package otel_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	emit "github.com/pseudofunctor-ai/go-emitter/emitter"
	"github.com/pseudofunctor-ai/go-emitter/emitter/backends/otel"
	t "github.com/pseudofunctor-ai/go-emitter/emitter/types"
)

var _ = Describe("OpenTelemetry Backend", func() {
	var (
		reader  *sdkmetric.ManualReader
		mp      *sdkmetric.MeterProvider
		backend *otel.OtelBackend
		ctx     context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		reader = sdkmetric.NewManualReader()
		mp = sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
		meter := mp.Meter("test-meter")
		backend = otel.NewOtelBackend(meter)
	})

	AfterEach(func() {
		mp.Shutdown(ctx)
	})

	Context("Integration with Emitter", func() {
		It("should work with the emitter", func() {
			emitter := emit.NewEmitter(backend)
			Expect(emitter).NotTo(BeNil())
		})
	})

	Context("EmitInt", func() {
		It("should emit a counter", func() {
			backend.EmitInt(ctx, "test.counter", map[string]interface{}{}, 5, t.COUNT)

			var rm metricdata.ResourceMetrics
			err := reader.Collect(ctx, &rm)
			Expect(err).To(BeNil())

			Expect(rm.ScopeMetrics).To(HaveLen(1))
			Expect(rm.ScopeMetrics[0].Metrics).To(HaveLen(1))

			metric := rm.ScopeMetrics[0].Metrics[0]
			Expect(metric.Name).To(Equal("test.counter"))

			sum, ok := metric.Data.(metricdata.Sum[int64])
			Expect(ok).To(BeTrue())
			Expect(sum.DataPoints).To(HaveLen(1))
			Expect(sum.DataPoints[0].Value).To(Equal(int64(5)))
		})

		It("should emit a gauge", func() {
			backend.EmitInt(ctx, "test.gauge", map[string]interface{}{}, 42, t.GAUGE)

			var rm metricdata.ResourceMetrics
			err := reader.Collect(ctx, &rm)
			Expect(err).To(BeNil())

			Expect(rm.ScopeMetrics).To(HaveLen(1))
			Expect(rm.ScopeMetrics[0].Metrics).To(HaveLen(1))

			metric := rm.ScopeMetrics[0].Metrics[0]
			Expect(metric.Name).To(Equal("test.gauge"))

			gauge, ok := metric.Data.(metricdata.Gauge[int64])
			Expect(ok).To(BeTrue())
			Expect(gauge.DataPoints).To(HaveLen(1))
			Expect(gauge.DataPoints[0].Value).To(Equal(int64(42)))
		})

		It("should emit a histogram", func() {
			backend.EmitInt(ctx, "test.histogram", map[string]interface{}{}, 100, t.HISTOGRAM)

			var rm metricdata.ResourceMetrics
			err := reader.Collect(ctx, &rm)
			Expect(err).To(BeNil())

			Expect(rm.ScopeMetrics).To(HaveLen(1))
			Expect(rm.ScopeMetrics[0].Metrics).To(HaveLen(1))

			metric := rm.ScopeMetrics[0].Metrics[0]
			Expect(metric.Name).To(Equal("test.histogram"))

			hist, ok := metric.Data.(metricdata.Histogram[int64])
			Expect(ok).To(BeTrue())
			Expect(hist.DataPoints).To(HaveLen(1))
			Expect(hist.DataPoints[0].Count).To(Equal(uint64(1)))
		})

		It("should include attributes from props", func() {
			props := map[string]interface{}{
				"user":   "alice",
				"status": 200,
				"active": true,
			}
			backend.EmitInt(ctx, "test.with.attrs", props, 1, t.COUNT)

			var rm metricdata.ResourceMetrics
			err := reader.Collect(ctx, &rm)
			Expect(err).To(BeNil())

			metric := rm.ScopeMetrics[0].Metrics[0]
			sum := metric.Data.(metricdata.Sum[int64])

			attrs := sum.DataPoints[0].Attributes
			val, ok := attrs.Value(attribute.Key("user"))
			Expect(ok).To(BeTrue())
			Expect(val.AsString()).To(Equal("alice"))

			val, ok = attrs.Value(attribute.Key("status"))
			Expect(ok).To(BeTrue())
			Expect(val.AsInt64()).To(Equal(int64(200)))

			val, ok = attrs.Value(attribute.Key("active"))
			Expect(ok).To(BeTrue())
			Expect(val.AsBool()).To(BeTrue())
		})

		It("should filter out special properties", func() {
			props := map[string]interface{}{
				"user":      "bob",
				"_rate":     2.0,
				"_message":  "test message",
				"_logLevel": "INFO",
			}
			backend.EmitInt(ctx, "test.filtered", props, 1, t.COUNT)

			var rm metricdata.ResourceMetrics
			err := reader.Collect(ctx, &rm)
			Expect(err).To(BeNil())

			metric := rm.ScopeMetrics[0].Metrics[0]
			sum := metric.Data.(metricdata.Sum[int64])

			attrs := sum.DataPoints[0].Attributes

			// Should have user
			_, ok := attrs.Value(attribute.Key("user"))
			Expect(ok).To(BeTrue())

			// Should NOT have special properties
			_, ok = attrs.Value(attribute.Key("_rate"))
			Expect(ok).To(BeFalse())
			_, ok = attrs.Value(attribute.Key("_message"))
			Expect(ok).To(BeFalse())
			_, ok = attrs.Value(attribute.Key("_logLevel"))
			Expect(ok).To(BeFalse())
		})
	})

	Context("EmitFloat", func() {
		It("should emit a float counter", func() {
			backend.EmitFloat(ctx, "test.float.counter", map[string]interface{}{}, 3.14, t.COUNT)

			var rm metricdata.ResourceMetrics
			err := reader.Collect(ctx, &rm)
			Expect(err).To(BeNil())

			metric := rm.ScopeMetrics[0].Metrics[0]
			Expect(metric.Name).To(Equal("test.float.counter"))

			sum, ok := metric.Data.(metricdata.Sum[float64])
			Expect(ok).To(BeTrue())
			Expect(sum.DataPoints[0].Value).To(Equal(3.14))
		})

		It("should emit a float gauge", func() {
			backend.EmitFloat(ctx, "test.float.gauge", map[string]interface{}{}, 98.6, t.GAUGE)

			var rm metricdata.ResourceMetrics
			err := reader.Collect(ctx, &rm)
			Expect(err).To(BeNil())

			metric := rm.ScopeMetrics[0].Metrics[0]
			gauge, ok := metric.Data.(metricdata.Gauge[float64])
			Expect(ok).To(BeTrue())
			Expect(gauge.DataPoints[0].Value).To(Equal(98.6))
		})

		It("should emit a float histogram", func() {
			backend.EmitFloat(ctx, "test.float.histogram", map[string]interface{}{}, 42.5, t.HISTOGRAM)

			var rm metricdata.ResourceMetrics
			err := reader.Collect(ctx, &rm)
			Expect(err).To(BeNil())

			metric := rm.ScopeMetrics[0].Metrics[0]
			hist, ok := metric.Data.(metricdata.Histogram[float64])
			Expect(ok).To(BeTrue())
			Expect(hist.DataPoints[0].Count).To(Equal(uint64(1)))
		})
	})

	Context("EmitDuration", func() {
		It("should emit duration as histogram in seconds", func() {
			duration := 1500 * time.Millisecond
			backend.EmitDuration(ctx, "test.duration", map[string]interface{}{}, duration, t.TIMER)

			var rm metricdata.ResourceMetrics
			err := reader.Collect(ctx, &rm)
			Expect(err).To(BeNil())

			metric := rm.ScopeMetrics[0].Metrics[0]
			Expect(metric.Name).To(Equal("test.duration"))

			hist, ok := metric.Data.(metricdata.Histogram[float64])
			Expect(ok).To(BeTrue())
			Expect(hist.DataPoints[0].Count).To(Equal(uint64(1)))
			// Duration should be recorded in seconds
			Expect(hist.DataPoints[0].Sum).To(Equal(1.5))
		})

		It("should handle multiple duration recordings", func() {
			backend.EmitDuration(ctx, "test.multi.duration", map[string]interface{}{}, 100*time.Millisecond, t.TIMER)
			backend.EmitDuration(ctx, "test.multi.duration", map[string]interface{}{}, 200*time.Millisecond, t.TIMER)

			var rm metricdata.ResourceMetrics
			err := reader.Collect(ctx, &rm)
			Expect(err).To(BeNil())

			metric := rm.ScopeMetrics[0].Metrics[0]
			hist, ok := metric.Data.(metricdata.Histogram[float64])
			Expect(ok).To(BeTrue())
			Expect(hist.DataPoints[0].Count).To(Equal(uint64(2)))
			Expect(hist.DataPoints[0].Sum).To(BeNumerically("~", 0.3, 0.001)) // 0.1 + 0.2 seconds
		})
	})

	Context("Instrument Caching", func() {
		It("should reuse instruments for the same metric name", func() {
			// Emit the same metric multiple times
			backend.EmitInt(ctx, "test.cached", map[string]interface{}{}, 1, t.COUNT)
			backend.EmitInt(ctx, "test.cached", map[string]interface{}{}, 2, t.COUNT)
			backend.EmitInt(ctx, "test.cached", map[string]interface{}{}, 3, t.COUNT)

			var rm metricdata.ResourceMetrics
			err := reader.Collect(ctx, &rm)
			Expect(err).To(BeNil())

			// Should have one metric with accumulated value
			Expect(rm.ScopeMetrics[0].Metrics).To(HaveLen(1))
			metric := rm.ScopeMetrics[0].Metrics[0]
			sum := metric.Data.(metricdata.Sum[int64])
			Expect(sum.DataPoints[0].Value).To(Equal(int64(6))) // 1+2+3
		})
	})

	Context("Metric Types", func() {
		It("should handle METER type as counter", func() {
			backend.EmitInt(ctx, "test.meter", map[string]interface{}{}, 10, t.METER)

			var rm metricdata.ResourceMetrics
			err := reader.Collect(ctx, &rm)
			Expect(err).To(BeNil())

			metric := rm.ScopeMetrics[0].Metrics[0]
			sum, ok := metric.Data.(metricdata.Sum[int64])
			Expect(ok).To(BeTrue())
			Expect(sum.DataPoints[0].Value).To(Equal(int64(10)))
		})

		It("should handle TIMER type with int value", func() {
			// Int value treated as milliseconds, converted to seconds in histogram
			backend.EmitInt(ctx, "test.timer.int", map[string]interface{}{}, 2000, t.TIMER)

			var rm metricdata.ResourceMetrics
			err := reader.Collect(ctx, &rm)
			Expect(err).To(BeNil())

			metric := rm.ScopeMetrics[0].Metrics[0]
			hist, ok := metric.Data.(metricdata.Histogram[float64])
			Expect(ok).To(BeTrue())
			Expect(hist.DataPoints[0].Sum).To(Equal(2.0)) // 2000ms = 2s
		})
	})
})
