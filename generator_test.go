package main

import (
	"go/ast"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestExtractCallSites(t *testing.T) {
	// Note: Tests that require emitter imports are covered in TestGeneratorIntegration
	// These tests focus on non-emitter code that should be filtered out
	tests := []struct {
		name      string
		source    string
		expected  map[string]CallSite
		shouldErr bool
	}{
		{
			name: "fmt.Errorf should be ignored",
			source: `package test
import "fmt"
func main() {
	_ = fmt.Errorf("error: %s", "message")
}`,
			expected: map[string]CallSite{}, // Should find nothing
		},
		{
			name: "slog calls should be ignored",
			source: `package test
import (
	"log/slog"
	"os"
)
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("message", "key", "value")
	logger.Error("error message", "code", "E500")
	logger.Debug("debug message")
}`,
			expected: map[string]CallSite{}, // Should find nothing
		},
		{
			name: "non-emitter type with matching method names should be ignored",
			source: `package test
import "context"
type FakeLogger struct{}
func (FakeLogger) Count(ctx context.Context, name string, props map[string]interface{}, n int) {}
func (FakeLogger) Info(msg string) {}
func main() {
	fake := FakeLogger{}
	fake.Count(context.Background(), "not_an_emitter", nil, 1)
	fake.Info("not an emitter")
}`,
			expected: map[string]CallSite{}, // Should find nothing
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for the test file
			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test.go")

			if err := os.WriteFile(testFile, []byte(tt.source), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			// Create a minimal go.mod file with absolute path to emitter
			wd, _ := os.Getwd()
			goMod := `module test

go 1.23

replace github.com/pseudofunctor-ai/go-emitter => ` + wd + `

require github.com/pseudofunctor-ai/go-emitter v0.0.0
`
			if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
				t.Fatalf("failed to write go.mod: %v", err)
			}

			// Load the package with full type information
			cfg := &packages.Config{
				Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
				Dir:  tmpDir,
			}

			pkgs, err := packages.Load(cfg, ".")
			if err != nil {
				t.Fatalf("failed to load package: %v", err)
			}

			if len(pkgs) == 0 {
				t.Fatalf("no packages found (loaded %d packages)", len(pkgs))
			}

			// Log package info for debugging
			if t.Failed() {
				t.Logf("Package: %s, Errors: %v", pkgs[0].ID, pkgs[0].Errors)
			}

			pkg := pkgs[0]

			// Check for package errors (but allow type errors for incomplete test code)
			if len(pkg.Errors) > 0 && !tt.shouldErr {
				t.Logf("package has errors: %v", pkg.Errors)
			}

			extractor := &callSiteExtractor{
				pkg:         pkg,
				currentFile: testFile,
				callsites:   make(map[string]CallSite),
			}

			for _, file := range pkg.Syntax {
				ast.Walk(extractor, file)
			}

			if tt.shouldErr {
				if extractor.err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if extractor.err != nil {
				t.Errorf("unexpected error: %v", extractor.err)
				return
			}

			if len(extractor.callsites) != len(tt.expected) {
				t.Errorf("expected %d callsites, got %d", len(tt.expected), len(extractor.callsites))
			}

			for eventName, expectedSite := range tt.expected {
				actualSite, found := extractor.callsites[eventName]
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
			}
		})
	}
}

func TestIsEmitterMethod(t *testing.T) {
	tests := []struct {
		method   string
		expected bool
	}{
		{"Count", true},
		{"Gauge", true},
		{"InfoContext", true},
		{"Metric", true},
		{"Log", true},
		{"MetricFnCallsite", false},
		{"LogFnCallsite", false},
		{"NotAMethod", false},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			result := isEmitterMethod(tt.method)
			if result != tt.expected {
				t.Errorf("isEmitterMethod(%q) = %v, expected %v", tt.method, result, tt.expected)
			}
		})
	}
}

func TestGetEventNameArgIndex(t *testing.T) {
	tests := []struct {
		method   string
		expected int
	}{
		{"Count", 1},       // context methods have event at index 1
		{"Gauge", 1},       // context methods have event at index 1
		{"InfoContext", 1}, // context methods have event at index 1
		{"EmitInt", 1},     // context methods have event at index 1
		{"Info", 0},        // non-context methods have event at index 0
		{"Metric", 0},      // registration methods have event at index 0
		{"Log", 0},         // registration methods have event at index 0
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			result := getEventNameArgIndex(tt.method)
			if result != tt.expected {
				t.Errorf("getEventNameArgIndex(%q) = %d, expected %d", tt.method, result, tt.expected)
			}
		})
	}
}
