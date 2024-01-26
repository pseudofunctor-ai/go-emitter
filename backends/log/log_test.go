package log

import (
	"context"
	"log/slog"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	emit "github.com/pseudofunctor-ai/go-emitter"
	"github.com/pseudofunctor-ai/go-emitter/backends/log/mocks"
	t "github.com/pseudofunctor-ai/go-emitter/types"
)

var _ = Describe("Interface", func() {
	It("should work with slog, otherwise what are we even doing here", func() {
		slog := slog.Default()
		emit := NewLogEmitter(slog)

		Expect(emit).ToNot(BeNil())
	})
})

var _ = Describe("LogEmitter", func() {
	It("should log at info", func() {
		ctrl := gomock.NewController(GinkgoT())
		defer ctrl.Finish()
		log := mocks.NewMockLoggerInterface(ctrl)
		logEmitter := NewLogEmitter(log)

		ctx := context.Background()
		log.EXPECT().InfoContext(ctx, "Hello World!")
		logEmitter.EmitInt(ctx, "test", map[string]interface{}{"_message": "Hello World!", "_logLevel": "INFO"}, 1, t.COUNT)
	})

	It("should log at error", func() {
		ctrl := gomock.NewController(GinkgoT())
		defer ctrl.Finish()
		log := mocks.NewMockLoggerInterface(ctrl)
		logEmitter := NewLogEmitter(log)

		ctx := context.Background()
		log.EXPECT().ErrorContext(ctx, "Hello World!")
		logEmitter.EmitInt(ctx, "test", map[string]interface{}{"_message": "Hello World!", "_logLevel": "ERROR"}, 1, t.COUNT)
	})

	It("should log at warn", func() {
		ctrl := gomock.NewController(GinkgoT())
		defer ctrl.Finish()
		log := mocks.NewMockLoggerInterface(ctrl)
		logEmitter := NewLogEmitter(log)

		ctx := context.Background()
		log.EXPECT().WarnContext(ctx, "Hello World!")
		logEmitter.EmitInt(ctx, "test", map[string]interface{}{"_message": "Hello World!", "_logLevel": "WARN"}, 1, t.COUNT)
	})

	It("should log at debug", func() {
		ctrl := gomock.NewController(GinkgoT())
		defer ctrl.Finish()
		log := mocks.NewMockLoggerInterface(ctrl)
		logEmitter := NewLogEmitter(log)

		ctx := context.Background()
		log.EXPECT().DebugContext(ctx, "Hello World!")
		logEmitter.EmitInt(ctx, "test", map[string]interface{}{"_message": "Hello World!", "_logLevel": "DEBUG"}, 1, t.COUNT)
	})

	It("should log at trace", func() {
		ctrl := gomock.NewController(GinkgoT())
		defer ctrl.Finish()
		log := mocks.NewMockLoggerInterface(ctrl)
		logEmitter := NewLogEmitter(log)

		ctx := context.Background()
		log.EXPECT().DebugContext(ctx, "Hello World!")
		logEmitter.EmitInt(ctx, "test", map[string]interface{}{"_message": "Hello World!", "_logLevel": "TRACE"}, 1, t.COUNT)
	})

	It("should log at fatal", func() {
		ctrl := gomock.NewController(GinkgoT())
		defer ctrl.Finish()
		log := mocks.NewMockLoggerInterface(ctrl)
		logEmitter := NewLogEmitter(log)

		ctx := context.Background()
		log.EXPECT().ErrorContext(ctx, "Hello World!")
		logEmitter.EmitInt(ctx, "test", map[string]interface{}{"_message": "Hello World!", "_logLevel": "FATAL"}, 1, t.COUNT)
	})

	It("Should correctly map properties", func() {
		ctrl := gomock.NewController(GinkgoT())
		defer ctrl.Finish()
		log := mocks.NewMockLoggerInterface(ctrl)
		logEmitter := NewLogEmitter(log)

		ctx := context.Background()
		log.EXPECT().InfoContext(ctx, "Hello World!", "Hello", "1", "foo", "bar")
		logEmitter.EmitInt(ctx, "test", map[string]interface{}{"_message": "Hello World!", "_logLevel": "INFO", "foo": "bar", "Hello": 1}, 1, t.COUNT)
	})

	It("should work as a backend emitter", func() {
		ctrl := gomock.NewController(GinkgoT())
		defer ctrl.Finish()
		log := mocks.NewMockLoggerInterface(ctrl)
		logEmitter := NewLogEmitter(log)
		emitter := emit.NewEmitter(logEmitter)

		log.EXPECT().InfoContext(context.Background(), "Hello World!")
		emitter.Info("test", nil, "Hello World!")
	})
})
