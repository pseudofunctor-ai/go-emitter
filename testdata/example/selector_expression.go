package example

import (
	"context"

	"github.com/pseudofunctor-ai/go-emitter/emitter/types"
	"github.com/pseudofunctor-ai/go-emitter/testdata/example/di"
)

// ServiceEmitters contains all emitter functions for a service
// This pattern tests that LogFnCallsite/MetricFnCallsite work with selector expressions
// like s.emitters.createSuccess (not just simple identifiers like decoratedLog)
type ServiceEmitters struct {
	createSuccess types.LogEmitterFn
	createFailure types.LogEmitterFn
	updateSuccess types.MetricEmitterFn
	updateFailure types.MetricEmitterFn
}

// initServiceEmitters initializes the emitters
func initServiceEmitters(deps *di.Dependencies) *ServiceEmitters {
	return &ServiceEmitters{
		createSuccess: deps.Emitter.Log("selector_create_success", deps.Emitter.DebugfContext),
		createFailure: deps.Emitter.Log("selector_create_failure", deps.Emitter.ErrorfContext),
		updateSuccess: deps.Emitter.Metric("selector_update_success", types.COUNT),
		updateFailure: deps.Emitter.Metric("selector_update_failure", types.COUNT),
	}
}

// Package-level callback created for DeleteItem test
var deleteSuccessCallback = em.Log("selector_delete_success", em.DebugfContext)

type service struct {
	deps     *di.Dependencies
	emitters *ServiceEmitters
}

func pkgEmitterHelper(deps *di.Dependencies) types.CombinedEmitter {
	return deps.Emitter
}

func NewServiceExample(deps *di.Dependencies) *service {
	return &service{
		deps:     deps,
		emitters: initServiceEmitters(deps),
	}
}

// CreateItem demonstrates the selector expression pattern in LogFnCallsite
// This pattern is common in production code where emitters are stored in a struct
func (s *service) CreateItem(ctx context.Context) error {
	// This pattern uses a selector expression (s.emitters.createSuccess) in LogFnCallsite
	// The key fix: extractCallsiteDecorator now handles ast.SelectorExpr, not just ast.Ident
	pkgEmitterHelper(s.deps).LogFnCallsite(s.emitters.createSuccess)(ctx, nil, "Item created")
	pkgEmitterHelper(s.deps).LogFnCallsite(s.emitters.createFailure)(ctx, nil, "Failed to create item")
	return nil
}

// UpdateItem demonstrates the selector expression pattern in MetricFnCallsite
func (s *service) UpdateItem(ctx context.Context) error {
	// Same pattern with metrics
	pkgEmitterHelper(s.deps).MetricFnCallsite(s.emitters.updateSuccess)(ctx, nil)
	pkgEmitterHelper(s.deps).MetricFnCallsite(s.emitters.updateFailure)(ctx, nil)
	return nil
}

// DeleteItem demonstrates copying a callback to a local variable before using it
// This tests that the generator can trace through variable assignments
func (s *service) DeleteItem(ctx context.Context) error {
	em := pkgEmitterHelper(s.deps)
	// Copy the package-level callback to a local variable
	cb := deleteSuccessCallback
	// Copy it again to test chained assignments
	cb2 := cb
	// Now use it via LogFnCallsite - the generator should trace 'cb2' -> 'cb' -> 'deleteSuccessCallback'
	em.LogFnCallsite(cb2)(ctx, nil, "Item deleted")
	return nil
}
