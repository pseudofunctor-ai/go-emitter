package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestExtractCallSites(t *testing.T) {
	// Note: Indirect decoration tests require full type information and are
	// covered in the integration tests (TestGeneratorIntegration).
	tests := []struct {
		name      string
		source    string
		expected  map[string]CallSite
		shouldErr bool
	}{
		{
			name: "direct count call",
			source: `package test
import "github.com/pseudofunctor-ai/go-emitter/emitter"
var em = emitter.NewEmitter()
func main() {
	em.Count(ctx, "direct_count", nil, 1)
}`,
			expected: map[string]CallSite{
				"direct_count": {
					EventName: "direct_count",
					LineNo:    5,
				},
			},
		},
		{
			name: "metric registration",
			source: `package test
import "github.com/pseudofunctor-ai/go-emitter/emitter"
var em = emitter.NewEmitter()
var userLogin = em.Metric("user_login", types.COUNT)`,
			expected: map[string]CallSite{
				"user_login": {
					EventName: "user_login",
					LineNo:    4,
				},
			},
		},
		{
			name: "log registration",
			source: `package test
import "github.com/pseudofunctor-ai/go-emitter/emitter"
var em = emitter.NewEmitter()
var auditLog = em.Log("audit_log", em.InfofContext)`,
			expected: map[string]CallSite{
				"audit_log": {
					EventName: "audit_log",
					LineNo:    4,
				},
			},
		},
		{
			name: "inline decorated metric",
			source: `package test
import "github.com/pseudofunctor-ai/go-emitter/emitter"
var em = emitter.NewEmitter()
var decorated = em.MetricFnCallsite(em.Metric("decorated_metric", types.COUNT))`,
			expected: map[string]CallSite{
				"decorated_metric": {
					EventName: "decorated_metric",
					LineNo:    4,
				},
			},
		},
		{
			name: "duplicate event names - should error",
			source: `package test
import "github.com/pseudofunctor-ai/go-emitter/emitter"
var em = emitter.NewEmitter()
var first = em.Metric("duplicate", types.COUNT)
var second = em.Metric("duplicate", types.GAUGE)`,
			shouldErr: true,
		},
		{
			name: "non-literal event name - should error",
			source: `package test
import "github.com/pseudofunctor-ai/go-emitter/emitter"
var em = emitter.NewEmitter()
const eventName = "my_event"
var metric = em.Metric(eventName, types.COUNT)`,
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.source, parser.ParseComments)
			if err != nil {
				t.Fatalf("failed to parse source: %v", err)
			}

			// Create a minimal package structure for testing
			pkg := &packages.Package{
				Syntax: []*ast.File{file},
				Fset:   fset,
				TypesInfo: &types.Info{
					Uses: make(map[*ast.Ident]types.Object),
				},
			}

			extractor := &callSiteExtractor{
				pkg:         pkg,
				currentFile: "test.go",
				callsites:   make(map[string]CallSite),
			}

			ast.Walk(extractor, file)

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
		{"Count", 1},         // context methods have event at index 1
		{"Gauge", 1},         // context methods have event at index 1
		{"InfoContext", 1},   // context methods have event at index 1
		{"EmitInt", 1},       // context methods have event at index 1
		{"Info", 0},          // non-context methods have event at index 0
		{"Metric", 0},        // registration methods have event at index 0
		{"Log", 0},           // registration methods have event at index 0
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
