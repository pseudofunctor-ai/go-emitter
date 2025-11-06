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
					LineNo:       21,
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.DirectCalls",
					PropertyKeys: []string{"environment", "region"},
					MetricType:   "COUNT",
				},
				"direct_gauge": {
					EventName:    "direct_gauge",
					LineNo:       22,
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.DirectCalls",
					PropertyKeys: []string{"host"},
					MetricType:   "GAUGE",
				},
				"direct_info_log": {
					EventName:    "direct_info_log",
					LineNo:       25,
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.DirectCalls",
					PropertyKeys: []string{"action", "user"},
					MetricType:   "COUNT",
				},
				"direct_error_log": {
					EventName:    "direct_error_log",
					LineNo:       26,
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.DirectCalls",
					PropertyKeys: []string{"code"},
					MetricType:   "COUNT",
				},
				"direct_debugf_log": {
					EventName:    "direct_debugf_log",
					LineNo:       27,
					FuncName:     "github.com/pseudofunctor-ai/go-emitter/testdata/example.DirectCalls",
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
				// Registered metrics with property keys from MetricWithProps
				"user_login_metric": {
					EventName:    "user_login_metric",
					LineNo:       32,
					PropertyKeys: []string{"user_id", "success"},
					MetricType:   "COUNT",
				},
				"request_duration": {
					EventName:    "request_duration",
					LineNo:       33,
					PropertyKeys: []string{"endpoint", "method"},
					MetricType:   "TIMER",
				},
				"active_users": {
					EventName:    "active_users",
					LineNo:       34,
					PropertyKeys: nil,
					MetricType:   "GAUGE",
				},
				// Registered logs with property keys from LogWithProps
				"audit_log": {
					EventName:    "audit_log",
					LineNo:       46,
					PropertyKeys: []string{"action", "resource", "user"},
					MetricType:   "COUNT",
				},
				"error_log": {
					EventName:    "error_log",
					LineNo:       47,
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
				// Inline decorated
				"decorated_metric": {
					EventName:    "decorated_metric",
					LineNo:       58,
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
				"decorated_log": {
					EventName:    "decorated_log",
					LineNo:       59,
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
				// Indirect decoration - should use decorator location, not definition
				"cache_hit_metric": {
					EventName:    "cache_hit_metric",
					LineNo:       77, // This is where MetricFnCallsite is called
					PropertyKeys: nil,
					MetricType:   "COUNT",
				},
				"auth_failure_log": {
					EventName:    "auth_failure_log",
					LineNo:       80, // This is where LogFnCallsite is called
					PropertyKeys: nil,
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
