package example

import (
	"context"

	"github.com/pseudofunctor-ai/go-emitter/emitter/types"
	"github.com/pseudofunctor-ai/go-emitter/testdata/example/di"
)

// This file tests field name collision handling - when multiple structs have fields with the same name
// The generator must use type information to find the correct initialization

// ServiceAEmitters has a getSuccess field with event "service_a_get_success"
type ServiceAEmitters struct {
	getSuccess       types.LogEmitterFn
	getFailure       types.LogEmitterFn
	directGetSuccess types.LogEmitterFn // For testing direct invocation
	directGetFailure types.LogEmitterFn
}

// ServiceBEmitters has a getSuccess field with event "service_b_get_success" (collision!)
type ServiceBEmitters struct {
	getSuccess       types.LogEmitterFn
	getFailure       types.LogEmitterFn
	directGetSuccess types.LogEmitterFn // For testing direct invocation
	directGetFailure types.LogEmitterFn
}

func initServiceAEmitters(deps *di.Dependencies) *ServiceAEmitters {
	return &ServiceAEmitters{
		getSuccess:       deps.Emitter.Log("service_a_get_success", deps.Emitter.DebugfContext),
		getFailure:       deps.Emitter.Log("service_a_get_failure", deps.Emitter.ErrorfContext),
		directGetSuccess: deps.Emitter.Log("service_a_direct_get_success", deps.Emitter.DebugfContext),
		directGetFailure: deps.Emitter.Log("service_a_direct_get_failure", deps.Emitter.ErrorfContext),
	}
}

func initServiceBEmitters(deps *di.Dependencies) *ServiceBEmitters {
	return &ServiceBEmitters{
		getSuccess:       deps.Emitter.Log("service_b_get_success", deps.Emitter.DebugfContext),
		getFailure:       deps.Emitter.Log("service_b_get_failure", deps.Emitter.ErrorfContext),
		directGetSuccess: deps.Emitter.Log("service_b_direct_get_success", deps.Emitter.DebugfContext),
		directGetFailure: deps.Emitter.Log("service_b_direct_get_failure", deps.Emitter.ErrorfContext),
	}
}

type serviceA struct {
	deps     *di.Dependencies
	emitters *ServiceAEmitters
}

type serviceB struct {
	deps     *di.Dependencies
	emitters *ServiceBEmitters
}

func NewServiceA(deps *di.Dependencies) *serviceA {
	return &serviceA{
		deps:     deps,
		emitters: initServiceAEmitters(deps),
	}
}

func NewServiceB(deps *di.Dependencies) *serviceB {
	return &serviceB{
		deps:     deps,
		emitters: initServiceBEmitters(deps),
	}
}

// TestServiceAWithCollision uses LogFnCallsite with s.emitters.getSuccess
// The generator must find "service_a_get_success", NOT "service_b_get_success"
func (s *serviceA) TestServiceAWithCollision(ctx context.Context) {
	pkgEmitterHelper(s.deps).LogFnCallsite(s.emitters.getSuccess)(ctx, nil, "Service A operation successful")
	pkgEmitterHelper(s.deps).LogFnCallsite(s.emitters.getFailure)(ctx, nil, "Service A operation failed")
}

// TestServiceBWithCollision uses LogFnCallsite with s.emitters.getSuccess
// The generator must find "service_b_get_success", NOT "service_a_get_success"
func (s *serviceB) TestServiceBWithCollision(ctx context.Context) {
	pkgEmitterHelper(s.deps).LogFnCallsite(s.emitters.getSuccess)(ctx, nil, "Service B operation successful")
	pkgEmitterHelper(s.deps).LogFnCallsite(s.emitters.getFailure)(ctx, nil, "Service B operation failed")
}

// TestDirectInvocationCollision tests direct callback invocation with field name collision
// Both ServiceA and ServiceB have a "directGetSuccess" field (name collision!)
// The generator must correctly identify which struct each callback belongs to
func TestDirectInvocationCollision() {
	ctx := context.Background()
	deps := di.NewDependencies()

	svcA := NewServiceA(&deps)
	svcB := NewServiceB(&deps)

	// Direct invocation of callbacks (not through LogFnCallsite)
	// Generator must find "service_a_direct_get_success", not "service_b_direct_get_success"
	svcA.emitters.directGetSuccess(ctx, nil, "Direct call from service A")
	// Generator must find "service_b_direct_get_success", not "service_a_direct_get_success"
	svcB.emitters.directGetSuccess(ctx, nil, "Direct call from service B")
}

