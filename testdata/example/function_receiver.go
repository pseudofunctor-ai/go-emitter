package example

import (
	"context"

	"github.com/pseudofunctor-ai/go-emitter/emitter/types"
	"github.com/pseudofunctor-ai/go-emitter/testdata/example/di"
)

// Helper function that returns an emitter
func getEmitter(deps *di.Dependencies) types.CombinedEmitter {
	return deps.Emitter
}

// Test LogFnCallsite with function call receiver
var functionCallTestLog = em.Log("function_call_test_log", em.InfofContext)

func FunctionCallReceiverTest() {
	ctx := context.Background()
	deps := di.NewDependencies()

	// This pattern should be detected: getEmitter(&deps).LogFnCallsite(...)
	getEmitter(&deps).LogFnCallsite(functionCallTestLog)(ctx, nil, "Test message")
}
