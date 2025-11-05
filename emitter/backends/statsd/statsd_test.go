package statsd

import (
	"context"
	"time"

	"github.com/cactus/go-statsd-client/v5/statsd"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	emit "github.com/pseudofunctor-ai/go-emitter/emitter"
	"github.com/pseudofunctor-ai/go-emitter/emitter/backends/statsd/mocks"
	t "github.com/pseudofunctor-ai/go-emitter/emitter/types"
)

// . "github.com/onsi/gomega"
var _ = Describe("Interface", func() {
	It("should work with statsd client", func() {
		statsdClient, ok := statsd.NewClientWithConfig(&statsd.ClientConfig{})
		Expect(ok).To(BeNil())
		statsdBackend := NewStatsdBackend(statsdClient)

		emit.NewEmitter(statsdBackend)
	})
})

var _ = Describe("Statsd", func() {
	It("should emit an int", func() {
		ctrl := gomock.NewController(GinkgoT())
		defer ctrl.Finish()
		mockStatsdClient := mocks.NewMockStatsdClient(ctrl)
		mockStatsdClient.EXPECT().Inc("foo", int64(1), float32(1.0)).Return(nil)
		statsdBackend := NewStatsdBackend(mockStatsdClient)
		statsdBackend.EmitInt(context.Background(), "foo", map[string]interface{}{}, 1, t.COUNT)
	})

	It("should emit a duration", func() {
		ctrl := gomock.NewController(GinkgoT())
		defer ctrl.Finish()
		mockStatsdClient := mocks.NewMockStatsdClient(ctrl)
		mockStatsdClient.EXPECT().TimingDuration("foo", time.Duration(1), float32(1.0)).Return(nil)

		statsdBackend := NewStatsdBackend(mockStatsdClient)
		statsdBackend.EmitDuration(context.Background(), "foo", map[string]interface{}{}, time.Duration(1), t.TIMER)
	})

	It("should emit a counter", func() {
		ctrl := gomock.NewController(GinkgoT())
		defer ctrl.Finish()
		mockStatsdClient := mocks.NewMockStatsdClient(ctrl)
		mockStatsdClient.EXPECT().Inc("foo", int64(5), float32(1.0)).Return(nil)
		statsdBackend := NewStatsdBackend(mockStatsdClient)
		statsdBackend.EmitInt(context.Background(), "foo", map[string]interface{}{}, 5, t.COUNT)
	})

	It("should emit adjust the rate of a counter", func() {
		ctrl := gomock.NewController(GinkgoT())
		defer ctrl.Finish()
		mockStatsdClient := mocks.NewMockStatsdClient(ctrl)
		mockStatsdClient.EXPECT().Inc("foo", int64(5), float32(2.0)).Return(nil)
		statsdBackend := NewStatsdBackend(mockStatsdClient)
		statsdBackend.EmitInt(context.Background(), "foo", map[string]interface{}{"_rate": 2.0}, 5, t.COUNT)
	})

	It("should emit a gauge", func() {
		ctrl := gomock.NewController(GinkgoT())
		defer ctrl.Finish()
		mockStatsdClient := mocks.NewMockStatsdClient(ctrl)
		mockStatsdClient.EXPECT().Gauge("foo", int64(5), float32(2.0)).Return(nil)
		statsdBackend := NewStatsdBackend(mockStatsdClient)
		statsdBackend.EmitInt(context.Background(), "foo", map[string]interface{}{"_rate": 2.0}, 5, t.GAUGE)
	})

	It("should emit a histogram", func() {
		ctrl := gomock.NewController(GinkgoT())
		defer ctrl.Finish()
		mockStatsdClient := mocks.NewMockStatsdClient(ctrl)
		mockStatsdClient.EXPECT().Gauge("foo", int64(5), float32(2.0)).Return(nil)
		statsdBackend := NewStatsdBackend(mockStatsdClient)
		statsdBackend.EmitInt(context.Background(), "foo", map[string]interface{}{"_rate": 2.0}, 5, t.HISTOGRAM)
	})

	It("Should correctly process properties to tags", func() {
		ctrl := gomock.NewController(GinkgoT())
		defer ctrl.Finish()
		mockStatsdClient := mocks.NewMockStatsdClient(ctrl)
		mockStatsdClient.EXPECT().Gauge("foo", int64(5), float32(2.0), []statsd.Tag{{"handled_unwanted_chars", "value_too"}, {"hello", "world"}}).Return(nil)
		statsdBackend := NewStatsdBackend(mockStatsdClient)
		statsdBackend.EmitInt(context.Background(), "foo", map[string]interface{}{"_rate": 2.0, "Hello": "World", "Handled unwanted ... chars": "value####too"}, 5, t.HISTOGRAM)
	})
})
