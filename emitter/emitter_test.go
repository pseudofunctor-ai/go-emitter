package emitter

import (
	"context"
	"runtime"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	. "github.com/pseudofunctor-ai/go-emitter/emitter/types"
	"github.com/pseudofunctor-ai/go-emitter/emitter/types/mocks"
)

var _ = Describe("Emitter", func() {
	var ctrl *gomock.Controller
	var mockBackend *mocks.MockEmitterBackend
	var emitter *Emitter

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockBackend = mocks.NewMockEmitterBackend(ctrl)
		emitter = NewEmitter(mockBackend)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Magic Properties", func() {
		It("Should populate magic properties when using WithAllMagicProps", func() {
			emitter := NewEmitter(mockBackend).WithAllMagicProps().WithHostnameProvider(func() (string, error) { return "localhost", nil })

			Expect(emitter.magicHostname).To(BeTrue())
			Expect(emitter.magicFilename).To(BeTrue())
			Expect(emitter.magicLineNo).To(BeTrue())
			Expect(emitter.magicFuncName).To(BeTrue())
			Expect(emitter.magicPackage).To(BeTrue())

			// XXX: BEGIN -- Do not add/remove lines between these comments without updating lineNo offset
			pc, file, lineNo, _ := runtime.Caller(0)
			funcName := runtime.FuncForPC(pc).Name()
			pkg := ""
			if lastSlash := strings.LastIndex(funcName, "/"); lastSlash >= 0 {
				if dot := strings.Index(funcName[lastSlash:], "."); dot >= 0 {
					pkg = funcName[:lastSlash+dot]
				}
			}
			mockBackend.EXPECT().EmitInt(context.Background(), "foo", map[string]interface{}{
				"Hello":    "World",
				"hostname": "localhost",
				"funcName": funcName,
				"lineNo":   lineNo + 16,
				"filename": file,
				"package":  pkg,
			}, int64(3), COUNT)
			emitter.EmitInt(context.Background(), "foo", map[string]interface{}{"Hello": "World"}, 3, COUNT)
			// XXX: END -- Do not add/remove lines between these comments without updating lineNo offset
		})

		It("Should populate magic properties when loggers", func() {
			emitter := NewEmitter(mockBackend).WithAllMagicProps().WithHostnameProvider(func() (string, error) { return "localhost", nil })

			Expect(emitter.magicHostname).To(BeTrue())
			Expect(emitter.magicFilename).To(BeTrue())
			Expect(emitter.magicLineNo).To(BeTrue())
			Expect(emitter.magicFuncName).To(BeTrue())
			Expect(emitter.magicPackage).To(BeTrue())

			// XXX: BEGIN -- Do not add/remove lines between these comments without updating lineNo offset
			pc, file, lineNo, _ := runtime.Caller(0)
			funcName := runtime.FuncForPC(pc).Name()
			pkg := ""
			if lastSlash := strings.LastIndex(funcName, "/"); lastSlash >= 0 {
				if dot := strings.Index(funcName[lastSlash:], "."); dot >= 0 {
					pkg = funcName[:lastSlash+dot]
				}
			}
			mockBackend.EXPECT().EmitInt(context.Background(), "foo", map[string]interface{}{
				"Hello":     "World",
				"hostname":  "localhost",
				"funcName":  funcName,
				"_logLevel": "INFO",
				"_message":  "Hello world!",
				"lineNo":    lineNo + 18,
				"filename":  file,
				"package":   pkg,
			}, int64(1), COUNT)
			emitter.InfoContext(context.Background(), "foo", map[string]interface{}{"Hello": "World"}, "Hello world!")
			// XXX: END -- Do not add/remove lines between these comments without updating lineNo offset
		})

		It("Should only add filename when WithMagicFilename is enabled", func() {
			_, file, _, _ := runtime.Caller(0)
			emitter := NewEmitter(mockBackend).WithMagicFilename()

			mockBackend.EXPECT().EmitInt(gomock.Any(), "test", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKey("filename"))
				Expect(props["filename"]).To(Equal(file))
				Expect(props).NotTo(HaveKey("lineNo"))
				Expect(props).NotTo(HaveKey("funcName"))
				Expect(props).NotTo(HaveKey("hostname"))
			})

			emitter.Count(context.Background(), "test", nil, 1)
		})

		It("Should only add line number when WithMagicLineNo is enabled", func() {
			emitter := NewEmitter(mockBackend).WithMagicLineNo()

			mockBackend.EXPECT().EmitInt(gomock.Any(), "test", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKey("lineNo"))
				Expect(props).NotTo(HaveKey("filename"))
				Expect(props).NotTo(HaveKey("funcName"))
				Expect(props).NotTo(HaveKey("hostname"))
			})

			emitter.Count(context.Background(), "test", nil, 1)
		})

		It("Should only add function name when WithMagicFuncName is enabled", func() {
			pc, _, _, _ := runtime.Caller(0)
			funcName := runtime.FuncForPC(pc).Name()
			emitter := NewEmitter(mockBackend).WithMagicFuncName()

			mockBackend.EXPECT().EmitInt(gomock.Any(), "test", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKey("funcName"))
				Expect(props["funcName"]).To(Equal(funcName))
				Expect(props).NotTo(HaveKey("filename"))
				Expect(props).NotTo(HaveKey("lineNo"))
				Expect(props).NotTo(HaveKey("hostname"))
			})

			emitter.Count(context.Background(), "test", nil, 1)
		})

		It("Should only add hostname when WithMagicHostname is enabled", func() {
			emitter := NewEmitter(mockBackend).WithMagicHostname().WithHostnameProvider(func() (string, error) { return "test-host", nil })

			mockBackend.EXPECT().EmitInt(gomock.Any(), "test", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKey("hostname"))
				Expect(props["hostname"]).To(Equal("test-host"))
				Expect(props).NotTo(HaveKey("filename"))
				Expect(props).NotTo(HaveKey("lineNo"))
				Expect(props).NotTo(HaveKey("funcName"))
			})

			emitter.Count(context.Background(), "test", nil, 1)
		})

		It("Should not add any magic properties when WithoutMagicProps is called", func() {
			emitter := NewEmitter(mockBackend).WithAllMagicProps().WithoutMagicProps()

			Expect(emitter.magicHostname).To(BeFalse())
			Expect(emitter.magicFilename).To(BeFalse())
			Expect(emitter.magicLineNo).To(BeFalse())
			Expect(emitter.magicFuncName).To(BeFalse())

			mockBackend.EXPECT().EmitInt(gomock.Any(), "test", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).NotTo(HaveKey("filename"))
				Expect(props).NotTo(HaveKey("lineNo"))
				Expect(props).NotTo(HaveKey("funcName"))
				Expect(props).NotTo(HaveKey("hostname"))
			})

			emitter.Count(context.Background(), "test", nil, 1)
		})

		It("Should not modify original props map", func() {
			emitter := NewEmitter(mockBackend).WithMagicFilename()

			originalProps := map[string]interface{}{"key": "value"}

			mockBackend.EXPECT().EmitInt(gomock.Any(), "test", gomock.Any(), int64(1), COUNT)

			emitter.Count(context.Background(), "test", originalProps, 1)

			Expect(originalProps).To(HaveLen(1))
			Expect(originalProps).To(HaveKeyWithValue("key", "value"))
			Expect(originalProps).NotTo(HaveKey("filename"))
		})
	})

	Describe("Callback", func() {
		It("Should invoke callback when set", func() {
			var capturedCtx context.Context
			var capturedEvent string
			var capturedProps map[string]interface{}

			emitter := NewEmitter(mockBackend).WithCallback(func(ctx context.Context, event string, props map[string]interface{}) {
				capturedCtx = ctx
				capturedEvent = event
				capturedProps = props
			})

			ctx := context.Background()
			mockBackend.EXPECT().EmitInt(ctx, "test_event", gomock.Any(), int64(1), COUNT)

			emitter.Count(ctx, "test_event", map[string]interface{}{"key": "value"}, 1)

			Expect(capturedCtx).To(Equal(ctx))
			Expect(capturedEvent).To(Equal("test_event"))
			Expect(capturedProps).To(HaveKeyWithValue("key", "value"))
		})
	})

	Describe("MetricsEmitter Interface", func() {
		It("Should emit Count metric", func() {
			mockBackend.EXPECT().EmitInt(context.Background(), "counter", map[string]interface{}{"foo": "bar"}, int64(42), COUNT)

			emitter.Count(context.Background(), "counter", map[string]interface{}{"foo": "bar"}, 42)
		})

		It("Should emit Gauge metric", func() {
			mockBackend.EXPECT().EmitFloat(context.Background(), "gauge", map[string]interface{}{"foo": "bar"}, float64(3.14), GAUGE)

			emitter.Gauge(context.Background(), "gauge", map[string]interface{}{"foo": "bar"}, 3.14)
		})

		It("Should emit Histogram metric", func() {
			mockBackend.EXPECT().EmitFloat(context.Background(), "histogram", map[string]interface{}{"foo": "bar"}, float64(1.5), HISTOGRAM)

			hist := emitter.Histogram(context.Background(), "histogram", map[string]interface{}{"foo": "bar"})
			hist.Observe(1.5)
		})

		It("Should emit Meter metric", func() {
			mockBackend.EXPECT().EmitInt(context.Background(), "meter", map[string]interface{}{"foo": "bar"}, int64(100), METER)

			emitter.Meter(context.Background(), "meter", map[string]interface{}{"foo": "bar"}, 100)
		})

		It("Should emit Set metric", func() {
			mockBackend.EXPECT().EmitInt(context.Background(), "set", map[string]interface{}{"foo": "bar"}, int64(5), SET)

			emitter.Set(context.Background(), "set", map[string]interface{}{"foo": "bar"}, 5)
		})

		It("Should emit Event metric", func() {
			mockBackend.EXPECT().EmitInt(context.Background(), "event", map[string]interface{}{"foo": "bar"}, int64(1), EVENT)

			emitter.Event(context.Background(), "event", map[string]interface{}{"foo": "bar"})
		})
	})

	Describe("SimpleLogger Interface", func() {
		It("Should emit Info log", func() {
			mockBackend.EXPECT().EmitInt(context.Background(), "test_event", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKeyWithValue("_logLevel", "INFO"))
				Expect(props).To(HaveKeyWithValue("_message", "info message"))
			})

			emitter.Info("test_event", map[string]interface{}{"key": "value"}, "info message")
		})

		It("Should emit Warn log", func() {
			mockBackend.EXPECT().EmitInt(context.Background(), "test_event", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKeyWithValue("_logLevel", "WARN"))
				Expect(props).To(HaveKeyWithValue("_message", "warn message"))
			})

			emitter.Warn("test_event", map[string]interface{}{"key": "value"}, "warn message")
		})

		It("Should emit Error log", func() {
			mockBackend.EXPECT().EmitInt(context.Background(), "test_event", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKeyWithValue("_logLevel", "ERROR"))
				Expect(props).To(HaveKeyWithValue("_message", "error message"))
			})

			emitter.Error("test_event", map[string]interface{}{"key": "value"}, "error message")
		})

		It("Should emit Fatal log", func() {
			mockBackend.EXPECT().EmitInt(context.Background(), "test_event", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKeyWithValue("_logLevel", "FATAL"))
				Expect(props).To(HaveKeyWithValue("_message", "fatal message"))
			})

			emitter.Fatal("test_event", map[string]interface{}{"key": "value"}, "fatal message")
		})

		It("Should emit Debug log", func() {
			mockBackend.EXPECT().EmitInt(context.Background(), "test_event", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKeyWithValue("_logLevel", "DEBUG"))
				Expect(props).To(HaveKeyWithValue("_message", "debug message"))
			})

			emitter.Debug("test_event", map[string]interface{}{"key": "value"}, "debug message")
		})

		It("Should emit Trace log", func() {
			mockBackend.EXPECT().EmitInt(context.Background(), "test_event", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKeyWithValue("_logLevel", "TRACE"))
				Expect(props).To(HaveKeyWithValue("_message", "trace message"))
			})

			emitter.Trace("test_event", map[string]interface{}{"key": "value"}, "trace message")
		})
	})

	Describe("SimpleLoggerFmt Interface", func() {
		It("Should emit Infof log", func() {
			mockBackend.EXPECT().EmitInt(context.Background(), "test_event", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKeyWithValue("_logLevel", "INFO"))
				Expect(props).To(HaveKeyWithValue("_message", "info 42"))
			})

			emitter.Infof("test_event", map[string]interface{}{"key": "value"}, "info %d", 42)
		})

		It("Should emit Warnf log", func() {
			mockBackend.EXPECT().EmitInt(context.Background(), "test_event", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKeyWithValue("_logLevel", "WARN"))
				Expect(props).To(HaveKeyWithValue("_message", "warn 42"))
			})

			emitter.Warnf("test_event", map[string]interface{}{"key": "value"}, "warn %d", 42)
		})

		It("Should emit Errorf log", func() {
			mockBackend.EXPECT().EmitInt(context.Background(), "test_event", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKeyWithValue("_logLevel", "ERROR"))
				Expect(props).To(HaveKeyWithValue("_message", "error 42"))
			})

			emitter.Errorf("test_event", map[string]interface{}{"key": "value"}, "error %d", 42)
		})

		It("Should emit Fatalf log", func() {
			mockBackend.EXPECT().EmitInt(context.Background(), "test_event", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKeyWithValue("_logLevel", "FATAL"))
				Expect(props).To(HaveKeyWithValue("_message", "fatal 42"))
			})

			emitter.Fatalf("test_event", map[string]interface{}{"key": "value"}, "fatal %d", 42)
		})

		It("Should emit Debugf log", func() {
			mockBackend.EXPECT().EmitInt(context.Background(), "test_event", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKeyWithValue("_logLevel", "DEBUG"))
				Expect(props).To(HaveKeyWithValue("_message", "debug 42"))
			})

			emitter.Debugf("test_event", map[string]interface{}{"key": "value"}, "debug %d", 42)
		})

		It("Should emit Tracef log", func() {
			mockBackend.EXPECT().EmitInt(context.Background(), "test_event", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKeyWithValue("_logLevel", "TRACE"))
				Expect(props).To(HaveKeyWithValue("_message", "trace 42"))
			})

			emitter.Tracef("test_event", map[string]interface{}{"key": "value"}, "trace %d", 42)
		})
	})

	Describe("ContextLogger Interface", func() {
		It("Should emit InfoContext log", func() {
			ctx := context.Background()
			mockBackend.EXPECT().EmitInt(ctx, "test_event", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKeyWithValue("_logLevel", "INFO"))
				Expect(props).To(HaveKeyWithValue("_message", "info message"))
			})

			emitter.InfoContext(ctx, "test_event", map[string]interface{}{"key": "value"}, "info message")
		})

		It("Should emit WarnContext log", func() {
			ctx := context.Background()
			mockBackend.EXPECT().EmitInt(ctx, "test_event", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKeyWithValue("_logLevel", "WARN"))
				Expect(props).To(HaveKeyWithValue("_message", "warn message"))
			})

			emitter.WarnContext(ctx, "test_event", map[string]interface{}{"key": "value"}, "warn message")
		})

		It("Should emit ErrorContext log", func() {
			ctx := context.Background()
			mockBackend.EXPECT().EmitInt(ctx, "test_event", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKeyWithValue("_logLevel", "ERROR"))
				Expect(props).To(HaveKeyWithValue("_message", "error message"))
			})

			emitter.ErrorContext(ctx, "test_event", map[string]interface{}{"key": "value"}, "error message")
		})

		It("Should emit FatalContext log", func() {
			ctx := context.Background()
			mockBackend.EXPECT().EmitInt(ctx, "test_event", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKeyWithValue("_logLevel", "FATAL"))
				Expect(props).To(HaveKeyWithValue("_message", "fatal message"))
			})

			emitter.FatalContext(ctx, "test_event", map[string]interface{}{"key": "value"}, "fatal message")
		})

		It("Should emit DebugContext log", func() {
			ctx := context.Background()
			mockBackend.EXPECT().EmitInt(ctx, "test_event", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKeyWithValue("_logLevel", "DEBUG"))
				Expect(props).To(HaveKeyWithValue("_message", "debug message"))
			})

			emitter.DebugContext(ctx, "test_event", map[string]interface{}{"key": "value"}, "debug message")
		})

		It("Should emit TraceContext log", func() {
			ctx := context.Background()
			mockBackend.EXPECT().EmitInt(ctx, "test_event", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKeyWithValue("_logLevel", "TRACE"))
				Expect(props).To(HaveKeyWithValue("_message", "trace message"))
			})

			emitter.TraceContext(ctx, "test_event", map[string]interface{}{"key": "value"}, "trace message")
		})
	})

	Describe("ContextLoggerFmt Interface", func() {
		It("Should emit InfofContext log", func() {
			ctx := context.Background()
			mockBackend.EXPECT().EmitInt(ctx, "test_event", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKeyWithValue("_logLevel", "INFO"))
				Expect(props).To(HaveKeyWithValue("_message", "info 42"))
			})

			emitter.InfofContext(ctx, "test_event", map[string]interface{}{"key": "value"}, "info %d", 42)
		})

		It("Should emit WarnfContext log", func() {
			ctx := context.Background()
			mockBackend.EXPECT().EmitInt(ctx, "test_event", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKeyWithValue("_logLevel", "WARN"))
				Expect(props).To(HaveKeyWithValue("_message", "warn 42"))
			})

			emitter.WarnfContext(ctx, "test_event", map[string]interface{}{"key": "value"}, "warn %d", 42)
		})

		It("Should emit ErrorfContext log", func() {
			ctx := context.Background()
			mockBackend.EXPECT().EmitInt(ctx, "test_event", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKeyWithValue("_logLevel", "ERROR"))
				Expect(props).To(HaveKeyWithValue("_message", "error 42"))
			})

			emitter.ErrorfContext(ctx, "test_event", map[string]interface{}{"key": "value"}, "error %d", 42)
		})

		It("Should emit FatalfContext log", func() {
			ctx := context.Background()
			mockBackend.EXPECT().EmitInt(ctx, "test_event", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKeyWithValue("_logLevel", "FATAL"))
				Expect(props).To(HaveKeyWithValue("_message", "fatal 42"))
			})

			emitter.FatalfContext(ctx, "test_event", map[string]interface{}{"key": "value"}, "fatal %d", 42)
		})

		It("Should emit DebugfContext log", func() {
			ctx := context.Background()
			mockBackend.EXPECT().EmitInt(ctx, "test_event", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKeyWithValue("_logLevel", "DEBUG"))
				Expect(props).To(HaveKeyWithValue("_message", "debug 42"))
			})

			emitter.DebugfContext(ctx, "test_event", map[string]interface{}{"key": "value"}, "debug %d", 42)
		})

		It("Should emit TracefContext log", func() {
			ctx := context.Background()
			mockBackend.EXPECT().EmitInt(ctx, "test_event", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKeyWithValue("_logLevel", "TRACE"))
				Expect(props).To(HaveKeyWithValue("_message", "trace 42"))
			})

			emitter.TracefContext(ctx, "test_event", map[string]interface{}{"key": "value"}, "trace %d", 42)
		})
	})

	Describe("EmitFloat, EmitInt, EmitDuration", func() {
		It("Should emit float values", func() {
			mockBackend.EXPECT().EmitFloat(context.Background(), "float_metric", map[string]interface{}{"key": "value"}, float64(3.14), GAUGE)

			emitter.EmitFloat(context.Background(), "float_metric", map[string]interface{}{"key": "value"}, 3.14, GAUGE)
		})

		It("Should emit int values", func() {
			mockBackend.EXPECT().EmitInt(context.Background(), "int_metric", map[string]interface{}{"key": "value"}, int64(42), COUNT)

			emitter.EmitInt(context.Background(), "int_metric", map[string]interface{}{"key": "value"}, 42, COUNT)
		})

		It("Should emit duration values", func() {
			duration := 5 * time.Second
			mockBackend.EXPECT().EmitDuration(context.Background(), "duration_metric", map[string]interface{}{"key": "value"}, duration, TIMER)

			emitter.EmitDuration(context.Background(), "duration_metric", map[string]interface{}{"key": "value"}, duration, TIMER)
		})
	})

	Describe("Multiple Backends", func() {
		It("Should forward to all backends", func() {
			backend1 := mocks.NewMockEmitterBackend(ctrl)
			backend2 := mocks.NewMockEmitterBackend(ctrl)

			emitter := NewEmitter(backend1, backend2)

			backend1.EXPECT().EmitInt(context.Background(), "test", map[string]interface{}{"key": "value"}, int64(1), COUNT)
			backend2.EXPECT().EmitInt(context.Background(), "test", map[string]interface{}{"key": "value"}, int64(1), COUNT)

			emitter.Count(context.Background(), "test", map[string]interface{}{"key": "value"}, 1)
		})

		It("Should add backends via WithBackend", func() {
			backend1 := mocks.NewMockEmitterBackend(ctrl)
			backend2 := mocks.NewMockEmitterBackend(ctrl)

			emitter := NewEmitter(backend1).WithBackend(backend2)

			backend1.EXPECT().EmitInt(context.Background(), "test", map[string]interface{}{"key": "value"}, int64(1), COUNT)
			backend2.EXPECT().EmitInt(context.Background(), "test", map[string]interface{}{"key": "value"}, int64(1), COUNT)

			emitter.Count(context.Background(), "test", map[string]interface{}{"key": "value"}, 1)
		})
	})

	Describe("Metric Registration", func() {
		It("Should register and use metric emitter function", func() {
			mockBackend.EXPECT().EmitInt(context.Background(), "registered_metric", map[string]interface{}{}, int64(0), COUNT)
			metricFn := emitter.Metric("registered_metric", COUNT)

			mockBackend.EXPECT().EmitInt(context.Background(), "registered_metric", map[string]interface{}{"foo": "bar"}, int64(1), COUNT)
			metricFn(context.Background(), map[string]interface{}{"foo": "bar"})
		})

		It("Should panic when registering duplicate event", func() {
			mockBackend.EXPECT().EmitInt(gomock.Any(), "duplicate", map[string]interface{}{}, int64(0), COUNT)
			emitter.Metric("duplicate", COUNT)

			Expect(func() {
				emitter.Metric("duplicate", COUNT)
			}).To(Panic())
		})
	})

	Describe("Log Registration", func() {
		It("Should register and use log emitter function", func() {
			mockBackend.EXPECT().EmitInt(context.Background(), "registered_log", map[string]interface{}{}, int64(0), COUNT)

			logFn := emitter.Log("registered_log", func(ctx context.Context, event string, props map[string]interface{}, format string, args ...interface{}) {
				emitter.DebugfContext(ctx, event, props, format, args...)
			})

			mockBackend.EXPECT().EmitInt(context.Background(), "registered_log", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKeyWithValue("_logLevel", "DEBUG"))
				Expect(props).To(HaveKeyWithValue("_message", "test 42"))
				Expect(props).To(HaveKeyWithValue("foo", "bar"))
			})

			logFn(context.Background(), map[string]interface{}{"foo": "bar"}, "test %d", 42)
		})

		It("Should panic when registering duplicate log event", func() {
			mockBackend.EXPECT().EmitInt(gomock.Any(), "duplicate", map[string]interface{}{}, int64(0), COUNT)
			emitter.Log("duplicate", func(ctx context.Context, event string, props map[string]interface{}, format string, args ...interface{}) {
			})

			Expect(func() {
				emitter.Log("duplicate", func(ctx context.Context, event string, props map[string]interface{}, format string, args ...interface{}) {
				})
			}).To(Panic())
		})
	})

	Describe("TimingEmitter", func() {
		It("Should measure execution time", func() {
			timingEmitter := NewTimingEmitter[int](emitter)

			mockBackend.EXPECT().EmitInt(context.Background(), "timing_test", map[string]interface{}{"foo": "bar"}, gomock.Any(), TIMER).Do(func(_ context.Context, _ string, _ map[string]interface{}, value int64, _ MetricType) {
				Expect(value).To(BeNumerically(">=", 10))
			})

			result := timingEmitter.Time(context.Background(), "timing_test", map[string]interface{}{"foo": "bar"}, func() int {
				time.Sleep(10 * time.Millisecond)
				return 42
			})

			Expect(result).To(Equal(42))
		})
	})

	Describe("CompatAdapter", func() {
		It("Should adapt variadic args to props map for InfoContext", func() {
			adapter := MakeCompatAdapter("compat_event", emitter)

			mockBackend.EXPECT().EmitInt(gomock.Any(), "compat_event", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKeyWithValue("_logLevel", "INFO"))
				Expect(props).To(HaveKeyWithValue("_message", "test message"))
				Expect(props).To(HaveKeyWithValue("key1", "value1"))
				Expect(props).To(HaveKeyWithValue("key2", "value2"))
			})

			adapter.InfoContext(context.Background(), "test message", "key1", "value1", "key2", "value2")
		})

		It("Should adapt variadic args to props map for WarnContext", func() {
			adapter := MakeCompatAdapter("compat_event", emitter)

			mockBackend.EXPECT().EmitInt(gomock.Any(), "compat_event", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKeyWithValue("_logLevel", "WARN"))
				Expect(props).To(HaveKeyWithValue("_message", "test message"))
			})

			adapter.WarnContext(context.Background(), "test message")
		})

		It("Should adapt variadic args to props map for ErrorContext", func() {
			adapter := MakeCompatAdapter("compat_event", emitter)

			mockBackend.EXPECT().EmitInt(gomock.Any(), "compat_event", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKeyWithValue("_logLevel", "ERROR"))
				Expect(props).To(HaveKeyWithValue("_message", "test message"))
			})

			adapter.ErrorContext(context.Background(), "test message")
		})

		It("Should adapt variadic args to props map for FatalContext", func() {
			adapter := MakeCompatAdapter("compat_event", emitter)

			mockBackend.EXPECT().EmitInt(gomock.Any(), "compat_event", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKeyWithValue("_logLevel", "FATAL"))
				Expect(props).To(HaveKeyWithValue("_message", "test message"))
			})

			adapter.FatalContext(context.Background(), "test message")
		})

		It("Should adapt variadic args to props map for DebugContext", func() {
			adapter := MakeCompatAdapter("compat_event", emitter)

			mockBackend.EXPECT().EmitInt(gomock.Any(), "compat_event", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKeyWithValue("_logLevel", "DEBUG"))
				Expect(props).To(HaveKeyWithValue("_message", "test message"))
			})

			adapter.DebugContext(context.Background(), "test message")
		})

		It("Should adapt variadic args to props map for TraceContext", func() {
			adapter := MakeCompatAdapter("compat_event", emitter)

			mockBackend.EXPECT().EmitInt(gomock.Any(), "compat_event", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKeyWithValue("_logLevel", "TRACE"))
				Expect(props).To(HaveKeyWithValue("_message", "test message"))
			})

			adapter.TraceContext(context.Background(), "test message")
		})

		It("Should panic on odd number of args", func() {
			adapter := MakeCompatAdapter("compat_event", emitter)

			Expect(func() {
				adapter.InfoContext(context.Background(), "test message", "key1")
			}).To(Panic())
		})
	})

	Describe("MetricWithProps", func() {
		It("Should register metric with property keys and emit seed value", func() {
			// Expect seed emission with placeholder values
			mockBackend.EXPECT().EmitInt(context.Background(), "api.request", map[string]interface{}{
				"endpoint": "*",
				"method":   "*",
			}, int64(0), COUNT)

			metricFn := emitter.MetricWithProps("api.request", COUNT, []string{"endpoint", "method"})

			// Verify it's registered
			_, exists := emitter.registeredEvents["api.request"]
			Expect(exists).To(BeTrue())

			// Emit with valid props
			mockBackend.EXPECT().EmitInt(gomock.Any(), "api.request", map[string]interface{}{
				"endpoint": "/users",
				"method":   "GET",
			}, int64(1), COUNT)

			metricFn(context.Background(), map[string]interface{}{
				"endpoint": "/users",
				"method":   "GET",
			})
		})

		It("Should emit an error when using unexpected property key", func() {
			mockBackend.EXPECT().EmitInt(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
			mockBackend.EXPECT().EmitInt(gomock.Any(), gomock.Any(), gomock.Eq(map[string]interface{}{"_logLevel": "ERROR", "_message": "Unexpected property key 'bad_key' for event 'api.request'. Expected keys: [endpoint]"}), int64(1), gomock.Any()).Times(1)

			metricFn := emitter.MetricWithProps("api.request", COUNT, []string{"endpoint"})

      metricFn(context.Background(), map[string]interface{}{
        "endpoint":   "/users",
        "bad_key":    "value",
      })
		})

		It("Should allow subset of property keys", func() {
			mockBackend.EXPECT().EmitInt(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())

			metricFn := emitter.MetricWithProps("api.request", COUNT, []string{"endpoint", "method", "status"})

			// Should not panic - using subset is fine
			mockBackend.EXPECT().EmitInt(gomock.Any(), "api.request", map[string]interface{}{
				"endpoint": "/users",
			}, int64(1), COUNT)

			metricFn(context.Background(), map[string]interface{}{
				"endpoint": "/users",
			})
		})
	})

	Describe("LogWithProps", func() {
		It("Should register log event with property keys and emit seed value", func() {
			// Expect seed emission with placeholder values
			mockBackend.EXPECT().EmitInt(context.Background(), "user.action", map[string]interface{}{
				"user_id": "*",
				"action":  "*",
			}, int64(0), COUNT)

			logFn := emitter.LogWithProps("user.action", emitter.InfofContext, []string{"user_id", "action"})

			// Verify it's registered
			_, exists := emitter.registeredEvents["user.action"]
			Expect(exists).To(BeTrue())

			// Emit with valid props
			mockBackend.EXPECT().EmitInt(gomock.Any(), "user.action", gomock.Any(), int64(1), COUNT).Do(func(_ context.Context, _ string, props map[string]interface{}, _ int64, _ MetricType) {
				Expect(props).To(HaveKeyWithValue("user_id", "123"))
				Expect(props).To(HaveKeyWithValue("action", "login"))
				Expect(props).To(HaveKeyWithValue("_message", "User logged in"))
				Expect(props).To(HaveKeyWithValue("_logLevel", "INFO"))
			})

			logFn(context.Background(), map[string]interface{}{
				"user_id": "123",
				"action":  "login",
			}, "User logged in")
		})

		It("Should panic when using unexpected property key", func() {
			mockBackend.EXPECT().EmitInt(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())

			logFn := emitter.LogWithProps("user.action", emitter.InfofContext, []string{"user_id"})

			Expect(func() {
				logFn(context.Background(), map[string]interface{}{
					"user_id":  "123",
					"bad_key": "value",
				}, "test")
			}).To(Panic())
		})
	})

	Describe("GetManifest", func() {
		It("Should return empty manifest for emitter with no registered events", func() {
			manifest := emitter.GetManifest()
			Expect(manifest).To(BeEmpty())
		})

		It("Should return manifest with registered metrics", func() {
			mockBackend.EXPECT().EmitInt(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			emitter.Metric("counter.event", COUNT)
			emitter.Metric("gauge.event", GAUGE)
			emitter.MetricWithProps("histogram.event", HISTOGRAM, []string{"bucket", "tag"})

			manifest := emitter.GetManifest()
			Expect(manifest).To(HaveLen(3))

			// Check counter
			Expect(manifest).To(ContainElement(MetricManifestEntry{
				Name:         "counter.event",
				MetricType:   COUNT,
				TypeString:   "COUNT",
				PropertyKeys: nil,
			}))

			// Check gauge
			Expect(manifest).To(ContainElement(MetricManifestEntry{
				Name:         "gauge.event",
				MetricType:   GAUGE,
				TypeString:   "GAUGE",
				PropertyKeys: nil,
			}))

			// Check histogram with props
			Expect(manifest).To(ContainElement(MetricManifestEntry{
				Name:         "histogram.event",
				MetricType:   HISTOGRAM,
				TypeString:   "HISTOGRAM",
				PropertyKeys: []string{"bucket", "tag"},
			}))
		})

		It("Should include log events in manifest", func() {
			mockBackend.EXPECT().EmitInt(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			emitter.Log("log.event", emitter.InfofContext)
			emitter.LogWithProps("log.event.props", emitter.WarnfContext, []string{"severity", "component"})

			manifest := emitter.GetManifest()
			Expect(manifest).To(HaveLen(2))

			// Check log without props
			Expect(manifest).To(ContainElement(MetricManifestEntry{
				Name:         "log.event",
				MetricType:   COUNT,
				TypeString:   "COUNT",
				PropertyKeys: nil,
			}))

			// Check log with props
			Expect(manifest).To(ContainElement(MetricManifestEntry{
				Name:         "log.event.props",
				MetricType:   COUNT,
				TypeString:   "COUNT",
				PropertyKeys: []string{"severity", "component"},
			}))
		})
	})

	Describe("NewSubEmitter", func() {
		It("Should create a sub-emitter with fresh registeredEvents and memoTable", func() {
			mockBackend.EXPECT().EmitInt(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			// Register an event in parent
			emitter.Metric("parent.event", COUNT)
			Expect(emitter.registeredEvents).To(HaveLen(1))

			// Create sub-emitter
			subEmitter := emitter.NewSubEmitter().(*Emitter)

			// Sub-emitter should have fresh maps
			Expect(subEmitter.registeredEvents).To(BeEmpty())
			Expect(subEmitter.memoTable).To(BeEmpty())

			// Parent should still have its registered event
			Expect(emitter.registeredEvents).To(HaveLen(1))
		})

		It("Should inherit all configuration from parent", func() {
			customCallback := func(ctx context.Context, event string, props map[string]interface{}) {}
			customHostnameProvider := func() (string, error) { return "test-host", nil }
			customCallsiteProvider := func(eventName string) CallSiteDetails {
				return CallSiteDetails{Filename: "test.go", LineNo: 42}
			}

			// Configure parent
			parent := NewEmitter(mockBackend).
				WithCallback(customCallback).
				WithHostnameProvider(customHostnameProvider).
				WithCallsiteProvider(customCallsiteProvider).
				WithMagicHostname().
				WithMagicFilename().
				WithMagicLineNo().
				WithMagicFuncName().
				WithMagicPackage()

			// Create sub-emitter
			sub := parent.NewSubEmitter().(*Emitter)

			// Check all flags are inherited
			Expect(sub.magicHostname).To(Equal(parent.magicHostname))
			Expect(sub.magicFilename).To(Equal(parent.magicFilename))
			Expect(sub.magicLineNo).To(Equal(parent.magicLineNo))
			Expect(sub.magicFuncName).To(Equal(parent.magicFuncName))
			Expect(sub.magicPackage).To(Equal(parent.magicPackage))

			// Verify providers are inherited by checking they return expected values
			hostname, _ := sub.hostname_provider()
			Expect(hostname).To(Equal("test-host"))

			callsite := sub.callsite_provider("test")
			Expect(callsite.Filename).To(Equal("test.go"))
			Expect(callsite.LineNo).To(Equal(42))
		})

		It("Should copy backends slice independently", func() {
			mockBackend2 := mocks.NewMockEmitterBackend(ctrl)
			parent := NewEmitter(mockBackend)

			// Create sub-emitter
			sub := parent.NewSubEmitter().(*Emitter)

			// Both should have same backend initially
			Expect(sub.backends).To(HaveLen(1))

			// Add backend to sub - should not affect parent
			sub.WithBackend(mockBackend2)
			Expect(sub.backends).To(HaveLen(2))
			Expect(parent.backends).To(HaveLen(1))

			// Add backend to parent - should not affect sub
			mockBackend3 := mocks.NewMockEmitterBackend(ctrl)
			parent.WithBackend(mockBackend3)
			Expect(parent.backends).To(HaveLen(2))
			Expect(sub.backends).To(HaveLen(2))
		})

		It("Should allow different metadata via WithStaticMetadata", func() {
			mockBackend.EXPECT().EmitInt(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			parent := NewEmitter(mockBackend)

			// Create metadata for parent (package1)
			parentMetadata := map[string]CallSiteDetails{
				"package1.event1": {
					MetricType:   "COUNT",
					PropertyKeys: []string{"prop1"},
				},
			}
			parent.WithStaticMetadata(parentMetadata)

			// Create sub-emitter with different metadata (package2)
			sub := parent.NewSubEmitter().(*Emitter)
			subMetadata := map[string]CallSiteDetails{
				"package2.event1": {
					MetricType:   "GAUGE",
					PropertyKeys: []string{"prop2"},
				},
			}
			sub.WithStaticMetadata(subMetadata)

			// Verify parent has its metadata
			parentManifest := parent.GetManifest()
			Expect(parentManifest).To(HaveLen(1))
			Expect(parentManifest[0].Name).To(Equal("package1.event1"))
			Expect(parentManifest[0].MetricType).To(Equal(COUNT))

			// Verify sub has its own metadata
			subManifest := sub.GetManifest()
			Expect(subManifest).To(HaveLen(1))
			Expect(subManifest[0].Name).To(Equal("package2.event1"))
			Expect(subManifest[0].MetricType).To(Equal(GAUGE))
		})

		It("Should emit to same backends as parent", func() {
			parent := NewEmitter(mockBackend).WithMagicHostname().WithHostnameProvider(func() (string, error) { return "test-host", nil })
			sub := parent.NewSubEmitter().(*Emitter)

			// Both should emit to the same backend
			mockBackend.EXPECT().EmitInt(gomock.Any(), "parent.event", gomock.Any(), int64(1), COUNT)
			parent.Count(context.Background(), "parent.event", nil, 1)

			mockBackend.EXPECT().EmitInt(gomock.Any(), "sub.event", gomock.Any(), int64(2), COUNT)
			sub.Count(context.Background(), "sub.event", nil, 2)
		})

		It("Should allow further configuration with builder methods", func() {
			parent := NewEmitter(mockBackend)
			sub := parent.NewSubEmitter().(*Emitter)

			// Initially should inherit parent settings (all false)
			Expect(sub.magicHostname).To(BeFalse())

			// Configure sub independently
			sub.WithMagicHostname()
			Expect(sub.magicHostname).To(BeTrue())
			Expect(parent.magicHostname).To(BeFalse())
		})

		It("Should have independent memoization tables", func() {
			parent := NewEmitter(mockBackend).WithMagicHostname().WithHostnameProvider(func() (string, error) { return "parent-host", nil })
			sub := parent.NewSubEmitter().(*Emitter).WithHostnameProvider(func() (string, error) { return "sub-host", nil })

			mockBackend.EXPECT().EmitInt(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			// Emit from parent - should memoize
			parent.Count(context.Background(), "shared.event", nil, 1)
			Expect(parent.memoTable).To(HaveKey("shared.event"))

			// Sub should have its own memoization
			sub.Count(context.Background(), "shared.event", nil, 1)
			Expect(sub.memoTable).To(HaveKey("shared.event"))

			// Verify they memoized different values (different hostnames)
			Expect(parent.memoTable["shared.event"].hostname).To(Equal("parent-host"))
			Expect(sub.memoTable["shared.event"].hostname).To(Equal("sub-host"))
		})
	})
})
