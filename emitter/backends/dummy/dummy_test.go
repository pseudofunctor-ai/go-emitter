package dummy

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	t "github.com/pseudofunctor-ai/go-emitter/emitter/types"
)

var _ = Describe("Dummy Emitter", func() {
	It("Should be able to be created", func() {
		dummy := NewDummyEmitter()
		Expect(dummy).NotTo(BeNil())
	})

	It("Should be able to have float metrics emitted to it", func() {
		dummy := NewDummyEmitter()
		dummy.EmitFloat(context.Background(), "test", nil, 5.3, t.GAUGE)
		Expect(dummy.Memo).Should(HaveLen(1))
		Expect(dummy.Memo["test"]).Should(HaveLen(1))
		Expect(dummy.Memo["test"][0].Value).To(Equal(5.3))
	})

	It("Should be able to have int metrics emitted to it", func() {
		dummy := NewDummyEmitter()
		dummy.EmitInt(context.Background(), "test", nil, 5, t.GAUGE)
		Expect(dummy.Memo).Should(HaveLen(1))
		Expect(dummy.Memo["test"]).Should(HaveLen(1))
		Expect(dummy.Memo["test"][0].Value).To(Equal(int64(5)))
	})

	It("Should be able to have duration metrics emitted to it", func() {
		dummy := NewDummyEmitter()
		dummy.EmitDuration(context.Background(), "test", nil, time.Duration(5)*time.Millisecond, t.TIMER)
		Expect(dummy.Memo).Should(HaveLen(1))
		Expect(dummy.Memo["test"]).Should(HaveLen(1))
		Expect(dummy.Memo["test"][0].Value).To(Equal(time.Duration(5) * time.Millisecond))
	})

	It("Should be able to have multiple metrics emitted to it", func() {
		dummy := NewDummyEmitter()
		dummy.EmitFloat(context.Background(), "test", nil, 5.3, t.GAUGE)
		dummy.EmitFloat(context.Background(), "test", nil, 5.3, t.GAUGE)
		dummy.EmitFloat(context.Background(), "test", nil, 5.3, t.GAUGE)

		Expect(dummy.Memo).Should(HaveLen(1))
		Expect(dummy.Memo["test"]).Should(HaveLen(3))
	})

	It("Should be able to have properties added to it", func() {
		dummy := NewDummyEmitter()

		dummy.EmitFloat(context.Background(), "test", map[string]interface{}{"hello": "world"}, 5.3, t.GAUGE)
		Expect(dummy.Memo).Should(HaveLen(1))
		Expect(dummy.Memo["test"]).Should(HaveLen(1))
		Expect(dummy.Memo["test"][0].Props).Should(HaveLen(1))
		Expect(dummy.Memo["test"][0].Props).Should(Equal(map[string]interface{}{"hello": "world"}))
	})

	It("Should correctly record a metric's type", func() {
		dummy := NewDummyEmitter()
		dummy.EmitDuration(context.Background(), "test", nil, time.Duration(5)*time.Millisecond, t.TIMER)
		dummy.EmitDuration(context.Background(), "test", nil, time.Duration(5)*time.Millisecond, t.COUNT)
		dummy.EmitDuration(context.Background(), "test", nil, time.Duration(5)*time.Millisecond, t.GAUGE)
		dummy.EmitDuration(context.Background(), "test", nil, time.Duration(5)*time.Millisecond, t.SET)
		dummy.EmitDuration(context.Background(), "test", nil, time.Duration(5)*time.Millisecond, t.METER)
		dummy.EmitDuration(context.Background(), "test", nil, time.Duration(5)*time.Millisecond, t.HISTOGRAM)

		Expect(dummy.Memo["test"]).To(HaveLen(6))
		Expect(dummy.Memo["test"][0].Type).To(Equal(t.TIMER))
		Expect(dummy.Memo["test"][1].Type).To(Equal(t.COUNT))
		Expect(dummy.Memo["test"][2].Type).To(Equal(t.GAUGE))
		Expect(dummy.Memo["test"][3].Type).To(Equal(t.SET))
		Expect(dummy.Memo["test"][4].Type).To(Equal(t.METER))
		Expect(dummy.Memo["test"][5].Type).To(Equal(t.HISTOGRAM))
	})
})
