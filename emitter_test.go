package emitter

import (
	"context"
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	. "github.com/pseudofunctor-ai/go-emitter/types"
	"github.com/pseudofunctor-ai/go-emitter/types/mocks"
)

var _ = Describe("Emitter", func() {
	It("Should populate magic properties when using WithAllMagicProps", func() {
		ctrl := gomock.NewController(GinkgoT())
		defer ctrl.Finish()

		m := mocks.NewMockEmitterBackend(ctrl)
		emitter := NewEmitter(m).WithAllMagicProps().WithHostnameProvider(func() (string, error) { return "localhost", nil })

		Expect(emitter.magicHostname).To(BeTrue())
		Expect(emitter.magicFilename).To(BeTrue())
		Expect(emitter.magicLineNo).To(BeTrue())
		Expect(emitter.magicFuncName).To(BeTrue())

		// XXX: BEGIN -- Do not add/remove lines between these comments without updating lineNo offset
		pc, file, lineNo, _ := runtime.Caller(0)
		funcName := runtime.FuncForPC(pc).Name()
		m.EXPECT().EmitInt(context.Background(), "foo", map[string]interface{}{
			"Hello":    "World",
			"hostname": "localhost",
			"funcName": funcName,
			"lineNo":   lineNo + 9,
			"filename": file,
		}, int64(3), COUNT)
		emitter.EmitInt(context.Background(), "foo", map[string]interface{}{"Hello": "World"}, 3, COUNT)
		// XXX: END -- Do not add/remove lines between these comments without updating lineNo offset
	})

	It("Should populate magic properties when loggers", func() {
		ctrl := gomock.NewController(GinkgoT())
		defer ctrl.Finish()

		m := mocks.NewMockEmitterBackend(ctrl)
		emitter := NewEmitter(m).WithAllMagicProps().WithHostnameProvider(func() (string, error) { return "localhost", nil })

		Expect(emitter.magicHostname).To(BeTrue())
		Expect(emitter.magicFilename).To(BeTrue())
		Expect(emitter.magicLineNo).To(BeTrue())
		Expect(emitter.magicFuncName).To(BeTrue())

		// XXX: BEGIN -- Do not add/remove lines between these comments without updating lineNo offset
		pc, file, lineNo, _ := runtime.Caller(0)
		funcName := runtime.FuncForPC(pc).Name()
		m.EXPECT().EmitInt(context.Background(), "foo", map[string]interface{}{
			"Hello":     "World",
			"hostname":  "localhost",
			"funcName":  funcName,
			"_logLevel": "INFO",
			"_message":  "Hello world!",
			"lineNo":    lineNo + 11,
			"filename":  file,
		}, int64(1), COUNT)
		emitter.InfoContext(context.Background(), "foo", map[string]interface{}{"Hello": "World"}, "Hello world!")
		// XXX: END -- Do not add/remove lines between these comments without updating lineNo offset
	})
})
