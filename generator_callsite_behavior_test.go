package main

import (
	"testing"
)

// TestCallSiteBehavior documents and tests the correct behavior for different
// patterns of metric/log registration and invocation.
func TestCallSiteBehavior(t *testing.T) {
	tests := []struct {
		name        string
		description string
		dir         string
		expected    map[string]CallSite
		notExpected []string // Event names that should NOT appear
	}{
		{
			name:        "registered metrics called directly without decorator",
			description: "Metrics registered with Metric() or MetricWithProps() should record callsite where INVOKED, not where DEFINED",
			dir:         "testdata/example",
			expected: map[string]CallSite{
				// These should be recorded at their invocation site (lines 42-44)
				// NOT at their definition site (lines 35-37)
				"user_login_metric": {
					EventName:    "user_login_metric",
					LineNo:       42,
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.RegisteredMetrics",
					PropertyKeys: []string{"user_id", "success"},
					MetricType:   "COUNT",
				},
				"request_duration": {
					EventName:    "request_duration",
					LineNo:       43,
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.RegisteredMetrics",
					PropertyKeys: []string{"endpoint", "method"},
					MetricType:   "TIMER",
				},
				"active_users": {
					EventName:    "active_users",
					LineNo:       44,
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.RegisteredMetrics",
					PropertyKeys: nil,
					MetricType:   "GAUGE",
				},
			},
		},
		{
			name:        "registered logs called directly without decorator",
			description: "Logs registered with Log() or LogWithProps() should record callsite where INVOKED, not where DEFINED",
			dir:         "testdata/example",
			expected: map[string]CallSite{
				// These should be recorded at their invocation site (lines 55-56)
				// NOT at their definition site (lines 49-50)
				"audit_log": {
					EventName:    "audit_log",
					LineNo:       55,
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.RegisteredLogs",
					PropertyKeys: []string{"action", "resource", "user"},
					MetricType:   "COUNT",
				},
				"error_log": {
					EventName:    "error_log",
					LineNo:       56,
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.RegisteredLogs",
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
			},
		},
		{
			name:        "inline decorated metrics and logs",
			description: "Callbacks should not have Callsite information recorded when they are defined, but SHOULD have Callsite information when they are decorated with an `*FnCallsite`",
			dir:         "testdata/example",
			expected: map[string]CallSite{
				// These should be at the MetricFnCallsite/LogFnCallsite invocation (lines 69-70)
				// NOT at the callback definition (lines 62-63)
				"decorated_metric": {
					EventName:    "decorated_metric",
					LineNo:       69,
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.DecoratedFunctions",
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
				"decorated_log": {
					EventName:    "decorated_log",
					LineNo:       70,
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.DecoratedFunctions",
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
			},
		},
		{
			name:        "indirect decoration - separate definition and decoration",
			description: "Callbacks should not have Callsite information recorded when they are defined, but SHOULD have Callsite information when they are decorated with an `*FnCallsite` or invoked",
			dir:         "testdata/example",
			expected: map[string]CallSite{
				// These should be at the MetricFnCallsite/LogFnCallsite call (lines 82, 85)
				// NOT at the definition (lines 74-75)
				"cache_hit_metric": {
					EventName:    "cache_hit_metric",
					LineNo:       82,
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.IndirectlyDecoratedFunctions",
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
				"auth_failure_log": {
					EventName:    "auth_failure_log",
					LineNo:       85,
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.IndirectlyDecoratedFunctions",
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
			},
		},
		{
			name:        "indirect decoration with property keys",
			description: "MetricWithProps and LogWithProps should preserve property keys when used with decorators",
			dir:         "testdata/example",
			expected: map[string]CallSite{
				"bloom_filter_reset": {
					EventName:    "bloom_filter_reset",
					LineNo:       98, // Line where MetricFnCallsite is called
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.IndirectlyDecoratedFunctionsWithProps",
					PropertyKeys: []string{"density", "service_count"},
					MetricType:   "COUNT",
				},
				"critical_confabulation": {
					EventName:    "critical_confabulation",
					LineNo:       101, // Line where LogFnCallsite is called
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.IndirectlyDecoratedFunctionsWithProps",
					PropertyKeys: []string{"confabulacity"},
					MetricType:   "COUNT",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			pkg, err := loadPackage(tt.dir)
			if err != nil {
				t.Fatalf("failed to load package: %v", err)
			}

			callsites, err := extractCallSites(pkg)
			if err != nil {
				t.Fatalf("failed to extract callsites: %v", err)
			}

			// Check expected callsites
			for eventName, expectedSite := range tt.expected {
				actualSite, found := callsites[eventName]
				if !found {
					t.Errorf("expected callsite for event %q not found", eventName)
					continue
				}

				if expectedSite.LineNo > 0 && actualSite.LineNo != expectedSite.LineNo {
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
						t.Errorf("event %q: PropertyKeys[%d] expected %q, got %q", eventName, i, expectedSite.PropertyKeys[i], actualSite.PropertyKeys[i])
					}
				}
			}

			// Check that unexpected events don't appear
			for _, unexpected := range tt.notExpected {
				if _, found := callsites[unexpected]; found {
					t.Errorf("UNEXPECTED: event %q should not be present", unexpected)
				}
			}
		})
	}
}
