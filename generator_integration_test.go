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
					LineNo:       28,
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.DirectCalls",
					PropertyKeys: []string{"environment", "region"},
					MetricType:   "COUNT",
				},
				"direct_gauge": {
					EventName:    "direct_gauge",
					LineNo:       29,
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.DirectCalls",
					PropertyKeys: []string{"host"},
					MetricType:   "GAUGE",
				},
				"direct_info_log": {
					EventName:    "direct_info_log",
					LineNo:       32,
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.DirectCalls",
					PropertyKeys: []string{"action", "user"},
					MetricType:   "COUNT",
				},
				"direct_error_log": {
					EventName:    "direct_error_log",
					LineNo:       33,
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.DirectCalls",
					PropertyKeys: []string{"code"},
					MetricType:   "COUNT",
				},
				"direct_debugf_log": {
					EventName:    "direct_debugf_log",
					LineNo:       34,
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.DirectCalls",
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
				// Registered metrics - recorded at INVOCATION site, not definition
				"user_login_metric": {
					EventName:    "user_login_metric",
					LineNo:       46, // Line where callback is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.RegisteredMetrics",
					PropertyKeys: []string{"user_id", "success"},
					MetricType:   "COUNT",
				},
				"request_duration": {
					EventName:    "request_duration",
					LineNo:       47, // Line where callback is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.RegisteredMetrics",
					PropertyKeys: []string{"endpoint", "method"},
					MetricType:   "TIMER",
				},
				"active_users": {
					EventName:    "active_users",
					LineNo:       48, // Line where callback is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.RegisteredMetrics",
					PropertyKeys: nil,
					MetricType:   "GAUGE",
				},
				// Registered logs - recorded at INVOCATION site, not definition
				"audit_log": {
					EventName:    "audit_log",
					LineNo:       59, // Line where callback is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.RegisteredLogs",
					PropertyKeys: []string{"action", "resource", "user"},
					MetricType:   "COUNT",
				},
				"error_log": {
					EventName:    "error_log",
					LineNo:       60, // Line where callback is invoked
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.RegisteredLogs",
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
				// Inline decorated - recorded at *FnCallsite invocation
				"decorated_metric": {
					EventName:    "decorated_metric",
					LineNo:       73, // Line where MetricFnCallsite is called
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.DecoratedFunctions",
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
				"decorated_log": {
					EventName:    "decorated_log",
					LineNo:       74, // Line where LogFnCallsite is called
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.DecoratedFunctions",
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
				// Indirect decoration - recorded at *FnCallsite invocation
				"cache_hit_metric": {
					EventName:    "cache_hit_metric",
					LineNo:       86, // Line where MetricFnCallsite is called
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.IndirectlyDecoratedFunctions",
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
				"auth_failure_log": {
					EventName:    "auth_failure_log",
					LineNo:       89, // Line where LogFnCallsite is called
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.IndirectlyDecoratedFunctions",
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
				// Indirect decoration with props - recorded at *FnCallsite invocation
				"bloom_filter_reset": {
					EventName:    "bloom_filter_reset",
					LineNo:       102, // Line where MetricFnCallsite is called
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.IndirectlyDecoratedFunctionsWithProps",
					PropertyKeys: []string{"density", "service_count"},
					MetricType:   "COUNT",
				},
				"critical_confabulation": {
					EventName:    "critical_confabulation",
					LineNo:       105, // Line where LogFnCallsite is called
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.IndirectlyDecoratedFunctionsWithProps",
					PropertyKeys: []string{"confabulacity"},
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
			}

			for _, unexpected := range unexpectedEvents {
				if _, found := callsites[unexpected]; found {
					t.Errorf("UNEXPECTED: found non-emitter call %q in callsites (should have been filtered out)", unexpected)
				}
			}

			if len(callsites) != len(tt.expected) {
				t.Errorf("expected %d callsites, got %d", len(tt.expected), len(callsites))
				t.Logf("Got callsites:")
				for name, cs := range callsites {
					t.Logf("  %q: line %d, func %q", name, cs.LineNo, cs.FuncName)
				}
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
