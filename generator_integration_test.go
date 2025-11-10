package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGeneratorIntegration(t *testing.T) {
	tests := []struct {
		name      string
		dir       string
		expected  map[string]CallSite
		shouldErr bool
	}{
		{
			name: "testdata example package",
			dir:  "testdata/example",
			expected: map[string]CallSite{
				// Direct calls with property keys extracted from map literals
				"direct_count": {
					EventName:    "direct_count",
					LineNo:       29,
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.DirectCalls",
					PropertyKeys: []string{"environment", "region"},
					MetricType:   "COUNT",
				},
				"direct_gauge": {
					EventName:    "direct_gauge",
					LineNo:       30,
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.DirectCalls",
					PropertyKeys: []string{"host"},
					MetricType:   "GAUGE",
				},
				"direct_info_log": {
					EventName:    "direct_info_log",
					LineNo:       33,
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.DirectCalls",
					PropertyKeys: []string{"action", "user"},
					MetricType:   "COUNT",
				},
				"direct_error_log": {
					EventName:    "direct_error_log",
					LineNo:       34,
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.DirectCalls",
					PropertyKeys: []string{"code"},
					MetricType:   "COUNT",
				},
				"direct_debugf_log": {
					EventName:    "direct_debugf_log",
					LineNo:       35,
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.DirectCalls",
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
				// Registered metrics - recorded at INVOCATION site, not definition
				"user_login_metric": {
					EventName:    "user_login_metric",
					LineNo:       47, // Line where callback is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.RegisteredMetrics",
					PropertyKeys: []string{"success", "user_id"},
					MetricType:   "COUNT",
				},
				"request_duration": {
					EventName:    "request_duration",
					LineNo:       48, // Line where callback is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.RegisteredMetrics",
					PropertyKeys: []string{"endpoint", "method"},
					MetricType:   "TIMER",
				},
				"active_users": {
					EventName:    "active_users",
					LineNo:       49, // Line where callback is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.RegisteredMetrics",
					PropertyKeys: nil,
					MetricType:   "GAUGE",
				},
				// Registered logs - recorded at INVOCATION site, not definition
				"audit_log": {
					EventName:    "audit_log",
					LineNo:       60, // Line where callback is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.RegisteredLogs",
					PropertyKeys: []string{"action", "resource", "user"},
					MetricType:   "COUNT",
				},
				"error_log": {
					EventName:    "error_log",
					LineNo:       61, // Line where callback is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.RegisteredLogs",
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
				// Inline decorated - recorded at *FnCallsite invocation
				"decorated_metric": {
					EventName:    "decorated_metric",
					LineNo:       74, // Line where MetricFnCallsite is called
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.DecoratedFunctions",
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
				"decorated_log": {
					EventName:    "decorated_log",
					LineNo:       75, // Line where LogFnCallsite is called
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.DecoratedFunctions",
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
				// Indirect decoration - recorded at *FnCallsite invocation
				"cache_hit_metric": {
					EventName:    "cache_hit_metric",
					LineNo:       87, // Line where MetricFnCallsite is called
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.IndirectlyDecoratedFunctions",
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
				"auth_failure_log": {
					EventName:    "auth_failure_log",
					LineNo:       90, // Line where LogFnCallsite is called
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.IndirectlyDecoratedFunctions",
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
				// Indirect decoration with props - recorded at *FnCallsite invocation
				"bloom_filter_reset": {
					EventName:    "bloom_filter_reset",
					LineNo:       103, // Line where MetricFnCallsite is called
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.IndirectlyDecoratedFunctionsWithProps",
					PropertyKeys: []string{"density", "service_count"},
					MetricType:   "COUNT",
				},
				"critical_confabulation": {
					EventName:    "critical_confabulation",
					LineNo:       106, // Line where LogFnCallsite is called
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.IndirectlyDecoratedFunctionsWithProps",
					PropertyKeys: []string{"confabulacity"},
					MetricType:   "COUNT",
				},
				// Function call receiver - emitter returned from function call
				"function_call_test_log": {
					EventName:    "function_call_test_log",
					LineNo:       23, // Line where getEmitter(&deps).LogFnCallsite is called
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.FunctionCallReceiverTest",
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
				// Selector expressions in decorators (e.g., s.emitters.createSuccess)
				"selector_create_success": {
					EventName:    "selector_create_success",
					LineNo:       54, // Line where LogFnCallsite(s.emitters.createSuccess) is called
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.CreateItem",
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
				"selector_create_failure": {
					EventName:    "selector_create_failure",
					LineNo:       55, // Line where LogFnCallsite(s.emitters.createFailure) is called
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.CreateItem",
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
				"selector_update_success": {
					EventName:    "selector_update_success",
					LineNo:       62, // Line where MetricFnCallsite(s.emitters.updateSuccess) is called
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.UpdateItem",
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
				"selector_update_failure": {
					EventName:    "selector_update_failure",
					LineNo:       63, // Line where MetricFnCallsite(s.emitters.updateFailure) is called
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.UpdateItem",
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
				// Tracing through chained variable assignments (cb2 := cb := deleteSuccessCallback)
				"selector_delete_success": {
					EventName:    "selector_delete_success",
					LineNo:       76, // Line where LogFnCallsite(cb2) is called
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.DeleteItem",
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
				// Field name collision handling - ServiceAEmitters.getSuccess vs ServiceBEmitters.getSuccess
				"service_a_get_success": {
					EventName:    "service_a_get_success",
					LineNo:       74, // Line where LogFnCallsite(s.emitters.getSuccess) is called in serviceA
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.TestServiceAWithCollision",
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
				"service_a_get_failure": {
					EventName:    "service_a_get_failure",
					LineNo:       75, // Line where LogFnCallsite(s.emitters.getFailure) is called in serviceA
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.TestServiceAWithCollision",
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
				"service_b_get_success": {
					EventName:    "service_b_get_success",
					LineNo:       81, // Line where LogFnCallsite(s.emitters.getSuccess) is called in serviceB
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.TestServiceBWithCollision",
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
				"service_b_get_failure": {
					EventName:    "service_b_get_failure",
					LineNo:       82, // Line where LogFnCallsite(s.emitters.getFailure) is called in serviceB
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.TestServiceBWithCollision",
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
				// Direct callback invocation with field name collision
				"service_a_direct_get_success": {
					EventName:    "service_a_direct_get_success",
					LineNo:       97, // Line where svcA.emitters.directGetSuccess is directly invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.TestDirectInvocationCollision",
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
				"service_b_direct_get_success": {
					EventName:    "service_b_direct_get_success",
					LineNo:       99, // Line where svcB.emitters.directGetSuccess is directly invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.TestDirectInvocationCollision",
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
				// Callbacks defined in separate compilation unit (struct fields)
				"event1": {
					EventName:    "event1",
					LineNo:       145, // Line where callback is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.EmittersDefinedInAnotherCompilationUnit",
					PropertyKeys: []string{"prop1", "prop2"},
					MetricType:   "COUNT",
				},
				"event2": {
					EventName:    "event2",
					LineNo:       146, // Line where callback is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.EmittersDefinedInAnotherCompilationUnit",
					PropertyKeys: []string{"metric1", "metric2"},
					MetricType:   "GAUGE",
				},
				"timed_event": {
					EventName:    "timed_event",
					LineNo:       147, // Line where timer callback is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.EmittersDefinedInAnotherCompilationUnit",
					PropertyKeys: []string{"Hello"},
					MetricType:   "TIMER",
				},
				// Callbacks in arrays/slices
				"array_event1": {
					EventName:    "array_event1",
					LineNo:       155, // Line where slice[0] is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.EmittersInArrays",
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
				"array_event2": {
					EventName:    "array_event2",
					LineNo:       156, // Line where slice[1] is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.EmittersInArrays",
					PropertyKeys: []string{"duration", "size"},
					MetricType:   "HISTOGRAM",
				},
				"array_event3": {
					EventName:    "array_event3",
					LineNo:       157, // Line where slice[2] is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.EmittersInArrays",
					PropertyKeys: nil,
					MetricType:   "GAUGE",
				},
				// Callbacks in maps
				"map_event1": {
					EventName:    "map_event1",
					LineNo:       165, // Line where m["info"] is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.EmittersInMaps",
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
				"map_event2": {
					EventName:    "map_event2",
					LineNo:       166, // Line where m["warn"] is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.EmittersInMaps",
					PropertyKeys: []string{"component", "severity"},
					MetricType:   "COUNT",
				},
				"map_event3": {
					EventName:    "map_event3",
					LineNo:       167, // Line where m["error"] is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.EmittersInMaps",
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
				// Timer calls
				"service.db.query": {
					EventName:    "service.db.query",
					LineNo:       41,
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.QueryDatabase",
					PropertyKeys: []string{"table"},
					MetricType:   "TIMER",
				},
				"service.api.call": {
					EventName:    "service.api.call",
					LineNo:       51,
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.CallExternalAPI",
					PropertyKeys: []string{"endpoint", "method"},
					MetricType:   "TIMER",
				},
				"service.process.data": {
					EventName:    "service.process.data",
					LineNo:       62,
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.ProcessData",
					PropertyKeys: nil,
					MetricType:   "TIMER",
				},
				"service.complex.operation": {
					EventName:    "service.complex.operation",
					LineNo:       71,
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.ComplexOperation",
					PropertyKeys: []string{"complexity"},
					MetricType:   "TIMER",
				},
				"service.chained.timer": {
					EventName:    "service.chained.timer",
					LineNo:       83,
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.ChainedTimerUsage",
					PropertyKeys: []string{"depth"},
					MetricType:   "TIMER",
				},
				"function.level.timer": {
					EventName:    "function.level.timer",
					LineNo:       95,
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.FunctionLevelTimer",
					PropertyKeys: []string{"scope"},
					MetricType:   "TIMER",
				},
				"inline.timer.event": {
					EventName:    "inline.timer.event",
					LineNo:       106,
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.InlineTimerCreation",
					PropertyKeys: []string{"inline"},
					MetricType:   "TIMER",
				},
				"slice.timer.event1": {
					EventName:    "slice.timer.event1",
					LineNo:       136,
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.TimerFromSlice",
					PropertyKeys: []string{"index"},
					MetricType:   "TIMER",
				},
				"map.timer.fast": {
					EventName:    "map.timer.fast",
					LineNo:       163,
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.TimerFromMap",
					PropertyKeys: []string{"category"},
					MetricType:   "TIMER",
				},
				"function.timer.event": {
					EventName:    "function.timer.event",
					LineNo:       180,
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.TimerFromFunction",
					PropertyKeys: []string{"source"},
					MetricType:   "TIMER",
				},
				// Advanced patterns: Direct function result indexing
				"inline_slice_index_0": {
					EventName:    "inline_slice_index_0",
					LineNo:       32, // Line where getCallbackSlice(em)[0](...) is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.DirectFunctionResultIndexing",
					PropertyKeys: nil, // Props at invocation site, not definition
					MetricType:   "COUNT",
				},
				"inline_slice_index_1": {
					EventName:    "inline_slice_index_1",
					LineNo:       33, // Line where getCallbackSlice(em)[1](...) is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.DirectFunctionResultIndexing",
					PropertyKeys: nil, // Props at invocation site, not definition
					MetricType:   "GAUGE",
				},
				"inline_map_success": {
					EventName:    "inline_map_success",
					LineNo:       36, // Line where getCallbackMap(em)["success"](...) is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.DirectFunctionResultIndexing",
					PropertyKeys: nil, // Props at invocation site, not definition
					MetricType:   "COUNT",
				},
				"inline_map_failure": {
					EventName:    "inline_map_failure",
					LineNo:       37, // Line where getCallbackMap(em)["failure"](...) is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.DirectFunctionResultIndexing",
					PropertyKeys: nil, // Props at invocation site, not definition
					MetricType:   "COUNT",
				},
				// Struct field arrays
				"struct_array_metric_0": {
					EventName:    "struct_array_metric_0",
					LineNo:       67, // Line where s.metricHandlers[0](...) is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.UseCallbackArray",
					PropertyKeys: nil, // Props at invocation site, not definition
					MetricType:   "COUNT",
				},
				"struct_array_metric_1": {
					EventName:    "struct_array_metric_1",
					LineNo:       68, // Line where s.metricHandlers[1](...) is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.UseCallbackArray",
					PropertyKeys: nil, // Props at invocation site, not definition
					MetricType:   "HISTOGRAM",
				},
				"struct_map_info": {
					EventName:    "struct_map_info",
					LineNo:       70, // Line where s.logHandlers["info"](...) is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.UseCallbackArray",
					PropertyKeys: nil, // Props at invocation site, not definition
					MetricType:   "COUNT",
				},
				"struct_map_error": {
					EventName:    "struct_map_error",
					LineNo:       71, // Line where s.logHandlers["error"](...) is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.UseCallbackArray",
					PropertyKeys: nil, // Props at invocation site, not definition
					MetricType:   "COUNT",
				},
				// Inline collection assignment
				"inline_assigned_0": {
					EventName:    "inline_assigned_0",
					LineNo:       92, // Line where handlers[0](...) is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.InlineCollectionAssignment",
					PropertyKeys: nil, // Props at invocation site, not definition
					MetricType:   "COUNT",
				},
				"inline_assigned_1": {
					EventName:    "inline_assigned_1",
					LineNo:       93, // Line where handlers[1](...) is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.InlineCollectionAssignment",
					PropertyKeys: nil, // Props at invocation site, not definition
					MetricType:   "GAUGE",
				},
				"inline_assigned_warn": {
					EventName:    "inline_assigned_warn",
					LineNo:       100, // Line where logMap["warn"](...) is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.InlineCollectionAssignment",
					PropertyKeys: nil, // Props at invocation site, not definition
					MetricType:   "COUNT",
				},
				"inline_assigned_debug": {
					EventName:    "inline_assigned_debug",
					LineNo:       101, // Line where logMap["debug"](...) is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.InlineCollectionAssignment",
					PropertyKeys: nil, // Props at invocation site, not definition
					MetricType:   "COUNT",
				},
				// Var declaration collections
				"var_decl_metric_0": {
					EventName:    "var_decl_metric_0",
					LineNo:       123, // Line where globalMetricSlice[0](...) is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.VarDeclCollectionUsage",
					PropertyKeys: nil, // Props at invocation site, not definition
					MetricType:   "COUNT",
				},
				"var_decl_metric_1": {
					EventName:    "var_decl_metric_1",
					LineNo:       124, // Line where globalMetricSlice[1](...) is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.VarDeclCollectionUsage",
					PropertyKeys: nil, // Props at invocation site, not definition
					MetricType:   "TIMER",
				},
				"var_decl_trace": {
					EventName:    "var_decl_trace",
					LineNo:       126, // Line where globalLogMap["trace"](...) is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.VarDeclCollectionUsage",
					PropertyKeys: nil, // Props at invocation site, not definition
					MetricType:   "COUNT",
				},
				"var_decl_fatal": {
					EventName:    "var_decl_fatal",
					LineNo:       127, // Line where globalLogMap["fatal"](...) is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.VarDeclCollectionUsage",
					PropertyKeys: nil, // Props at invocation site, not definition
					MetricType:   "COUNT",
				},
				// Decorator override (decorator should win, marking line 144 not 138)
				"decorator_override_event": {
					EventName:    "decorator_override_event",
					LineNo:       144, // Line where em.MetricFnCallsite(baseCallback)(...) is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.CallbackWithDecorator",
					PropertyKeys: nil, // Props at invocation site, not definition
					MetricType:   "COUNT",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg, err := loadPackage(tt.dir)
			if tt.shouldErr {
				if err == nil {
					t.Errorf("expected error loading package but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("failed to load package: %v", err)
			}

			callsites, err := extractCallSites(pkg)
			if tt.shouldErr {
				if err == nil {
					t.Errorf("expected error extracting callsites but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("failed to extract callsites: %v", err)
			}

			// First, check that we don't have any unexpected callsites (non-emitter calls)
			unexpectedEvents := []string{
				// These should NOT be picked up - they're not emitter calls
				"error with code: %s",       // fmt.Errorf
				"database error: %v",        // fmt.Errorf
				"user logged in",            // slog.Info
				"database error",            // slog.Error
				"debug message",             // slog.Debug
				"context log",               // slog.InfoContext
				"context error",             // slog.ErrorContext
				"should not be picked up",  // FakeLogger.Info
				"not_an_emitter_count",      // FakeLogger.Count
				"not_an_emitter_gauge",      // FakeLogger.Gauge
				// FakeEmitter calls with wrong signatures (should be filtered by signature validation)
				"fake_count",                // FakeEmitter.Count (wrong signature)
				"fake_gauge",                // FakeEmitter.Gauge (wrong signature)
				"fake_info",                 // FakeEmitter.InfoContext (wrong signature)
			}

			for _, unexpected := range unexpectedEvents {
				if _, found := callsites[unexpected]; found {
					t.Errorf("UNEXPECTED: found non-emitter call %q in callsites (should have been filtered out)", unexpected)
				}
			}

			if len(callsites) != len(tt.expected) {
				t.Errorf("expected %d callsites, got %d", len(tt.expected), len(callsites))
			}

			for eventName, expectedSite := range tt.expected {
				actualSite, found := callsites[eventName]
				if !found {
					t.Errorf("expected callsite for event %q not found", eventName)
					continue
				}

				if actualSite.EventName != expectedSite.EventName {
					t.Errorf("event %q: expected EventName %q, got %q", eventName, expectedSite.EventName, actualSite.EventName)
				}

				if actualSite.LineNo != expectedSite.LineNo {
					t.Errorf("event %q: expected LineNo %d, got %d", eventName, expectedSite.LineNo, actualSite.LineNo)
				}

				if expectedSite.FuncName != "" && actualSite.FuncName != expectedSite.FuncName {
					t.Errorf("event %q: expected FuncName %q, got %q", eventName, expectedSite.FuncName, actualSite.FuncName)
				}

				if expectedSite.MetricType != "" && actualSite.MetricType != expectedSite.MetricType {
					t.Errorf("event %q: expected MetricType %q, got %q", eventName, expectedSite.MetricType, actualSite.MetricType)
				}

				// Compare property keys
				if len(actualSite.PropertyKeys) != len(expectedSite.PropertyKeys) {
					t.Errorf("event %q: expected PropertyKeys %v, got %v", eventName, expectedSite.PropertyKeys, actualSite.PropertyKeys)
					continue
				}
				for i := range expectedSite.PropertyKeys {
					if actualSite.PropertyKeys[i] != expectedSite.PropertyKeys[i] {
						t.Errorf("event %q: expected PropertyKeys %v, got %v", eventName, expectedSite.PropertyKeys, actualSite.PropertyKeys)
						break
					}
				}
			}
		})
	}
}

func TestGenerateOutput(t *testing.T) {
	// Test that the generator creates valid Go code
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "test_callsites.go")

	config := GeneratorConfig{
		Directory:  "testdata/example",
		OutputFile: outputFile,
		VarName:    "TestCallsites",
		PkgName:    "example",
	}

	err := Generate(config)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check that the file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatalf("output file was not created")
	}

	// Read and verify basic structure
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	contentStr := string(content)

	// Check for required elements
	required := []string{
		"package example",
		"var TestCallsites = map[string]types.CallSiteDetails{",
		`"direct_count"`,
		`"cache_hit_metric"`,
	}

	for _, req := range required {
		if !contains(contentStr, req) {
			t.Errorf("output file missing required element: %q", req)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
