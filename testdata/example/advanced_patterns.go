package example

import (
	"context"

	"github.com/pseudofunctor-ai/go-emitter/emitter/types"
	"github.com/pseudofunctor-ai/go-emitter/testdata/example/di"
)

// Pattern 1: Direct function result indexing
// Tests extractCallbackFromIndexedFunctionResult

func getCallbackSlice(e types.CombinedEmitter) []types.MetricEmitterFn {
	return []types.MetricEmitterFn{
		e.Metric("inline_slice_index_0", types.COUNT),
		e.Metric("inline_slice_index_1", types.GAUGE),
	}
}

func getCallbackMap(e types.CombinedEmitter) map[string]types.LogEmitterFn {
	return map[string]types.LogEmitterFn{
		"success": e.Log("inline_map_success", e.InfofContext),
		"failure": e.Log("inline_map_failure", e.ErrorfContext),
	}
}

func DirectFunctionResultIndexing() {
	ctx := context.Background()
	deps := di.NewDependencies()

	// Direct indexing of function result (no intermediate variable)
	getCallbackSlice(deps.Emitter)[0](ctx, map[string]interface{}{"direct": true})
	getCallbackSlice(deps.Emitter)[1](ctx, map[string]interface{}{"index": 1})

	// Direct map key access on function result
	getCallbackMap(deps.Emitter)["success"](ctx, nil, "Success: %s", "ok")
	getCallbackMap(deps.Emitter)["failure"](ctx, map[string]interface{}{"code": 500}, "Failure: %s", "error")
}

// Pattern 2: Struct field containing callback arrays
// Tests findCallbackInStructFieldArray

type ServiceWithCallbackArrays struct {
	deps           *di.Dependencies
	metricHandlers []types.MetricEmitterFn
	logHandlers    map[string]types.LogEmitterFn
}

func newServiceWithCallbackArrays(deps *di.Dependencies) *ServiceWithCallbackArrays {
	return &ServiceWithCallbackArrays{
		deps: deps,
		metricHandlers: []types.MetricEmitterFn{
			deps.Emitter.Metric("struct_array_metric_0", types.COUNT),
			deps.Emitter.Metric("struct_array_metric_1", types.HISTOGRAM),
		},
		logHandlers: map[string]types.LogEmitterFn{
			"info":  deps.Emitter.Log("struct_map_info", deps.Emitter.InfofContext),
			"error": deps.Emitter.Log("struct_map_error", deps.Emitter.ErrorfContext),
		},
	}
}

func (s *ServiceWithCallbackArrays) UseCallbackArray() {
	ctx := context.Background()

	// Access callbacks from struct field arrays
	s.metricHandlers[0](ctx, map[string]interface{}{"source": "array"})
	s.metricHandlers[1](ctx, nil)

	s.logHandlers["info"](ctx, map[string]interface{}{"level": "info"}, "Info: %s", "message")
	s.logHandlers["error"](ctx, nil, "Error: %s", "problem")
}

func TestStructFieldArrays() {
	deps := di.NewDependencies()
	svc := newServiceWithCallbackArrays(&deps)
	svc.UseCallbackArray()
}

// Pattern 3: Inline collection assignment (not in function)
// Tests extractCallbackFromCollectionAssignment

func InlineCollectionAssignment() {
	ctx := context.Background()
	deps := di.NewDependencies()

	// Inline slice literal assignment
	handlers := []types.MetricEmitterFn{
		deps.Emitter.Metric("inline_assigned_0", types.COUNT),
		deps.Emitter.Metric("inline_assigned_1", types.GAUGE),
	}
	handlers[0](ctx, map[string]interface{}{"inline": true})
	handlers[1](ctx, nil)

	// Inline map literal assignment
	logMap := map[string]types.LogEmitterFn{
		"warn": deps.Emitter.Log("inline_assigned_warn", deps.Emitter.WarnfContext),
		"debug": deps.Emitter.Log("inline_assigned_debug", deps.Emitter.DebugfContext),
	}
	logMap["warn"](ctx, map[string]interface{}{"severity": "high"}, "Warning: %s", "test")
	logMap["debug"](ctx, nil, "Debug: %s", "info")
}

// Pattern 4: Var declaration with collection
// Tests extractCallbackFromCollectionValueSpec

var (
	globalMetricSlice = []types.MetricEmitterFn{
		// Note: these use a package-level emitter initialized in example.go
		em.Metric("var_decl_metric_0", types.COUNT),
		em.Metric("var_decl_metric_1", types.TIMER),
	}

	globalLogMap = map[string]types.LogEmitterFn{
		"trace": em.Log("var_decl_trace", em.TracefContext),
		"fatal": em.Log("var_decl_fatal", em.FatalfContext),
	}
)

func VarDeclCollectionUsage() {
	ctx := context.Background()

	globalMetricSlice[0](ctx, map[string]interface{}{"global": true})
	globalMetricSlice[1](ctx, nil)

	globalLogMap["trace"](ctx, map[string]interface{}{"trace_level": 3}, "Trace: %s", "data")
	globalLogMap["fatal"](ctx, nil, "Fatal: %s", "critical")
}

// Pattern 5: Decorator override behavior
// Tests recordCallsite with decorator overriding existing entry

var baseCallback = em.Metric("decorator_override_event", types.COUNT)

func CallbackWithoutDecorator() {
	ctx := context.Background()
	// This invocation should be recorded initially
	baseCallback(ctx, map[string]interface{}{"decorated": false})
}

func CallbackWithDecorator() {
	ctx := context.Background()
	// This decorator call should OVERRIDE the above invocation's callsite
	em.MetricFnCallsite(baseCallback)(ctx, map[string]interface{}{"decorated": true})
}
