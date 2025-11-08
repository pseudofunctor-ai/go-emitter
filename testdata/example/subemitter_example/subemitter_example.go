package subemitter_example

import (
	"context"

	"github.com/pseudofunctor-ai/go-emitter/emitter"
	"github.com/pseudofunctor-ai/go-emitter/emitter/backends/dummy"
	"github.com/pseudofunctor-ai/go-emitter/emitter/types"
	"github.com/pseudofunctor-ai/go-emitter/testdata/example/di"
)

// SubEmitterExample demonstrates using NewSubEmitter to create child emitters
// with different metadata for different packages while sharing configuration.
func SubEmitterExample(deps di.Dependencies) {
	// Create a parent emitter with shared configuration
	parentEmitter := deps.Emitter

	// Example 1: Sub-emitter with static callsite provider for package1
	// In a real scenario, this would use generated metadata from package1
	package1Metadata := map[string]types.CallSiteDetails{
		"package1.api_request": {
			Filename:     "/app/package1/api.go",
			LineNo:       25,
			FuncName:     "github.com/example/package1.HandleRequest",
			Package:      "github.com/example/package1",
			PropertyKeys: []string{"endpoint", "method"},
			MetricType:   "COUNT",
		},
		"package1.response_time": {
			Filename:     "/app/package1/api.go",
			LineNo:       42,
			FuncName:     "github.com/example/package1.HandleRequest",
			Package:      "github.com/example/package1",
			PropertyKeys: []string{"endpoint"},
			MetricType:   "TIMER",
		},
	}

	package1Emitter := parentEmitter.NewSubEmitter().
		WithCallsiteProvider(emitter.StaticCallsiteProvider(package1Metadata)).
		WithStaticMetadata(package1Metadata)

	// Example 2: Sub-emitter with different static metadata (simulating another package)
	package2Metadata := map[string]types.CallSiteDetails{
		"package2.db_query": {
			Filename:     "/app/package2/db.go",
			LineNo:       42,
			FuncName:     "github.com/example/package2.QueryDB",
			Package:      "github.com/example/package2",
			PropertyKeys: []string{"table", "operation"},
			MetricType:   "HISTOGRAM",
		},
		"package2.cache_hit": {
			Filename:     "/app/package2/cache.go",
			LineNo:       15,
			FuncName:     "github.com/example/package2.Get",
			Package:      "github.com/example/package2",
			PropertyKeys: []string{"key"},
			MetricType:   "COUNT",
		},
	}

	package2Emitter := parentEmitter.NewSubEmitter().
		WithCallsiteProvider(emitter.StaticCallsiteProvider(package2Metadata)).
		WithStaticMetadata(package2Metadata)

	// Example 3: Sub-emitter using default dynamic callsite provider
	// This still uses runtime.Caller but inherits other configuration
	dynamicEmitter := parentEmitter.NewSubEmitter()

	// All three sub-emitters share the same backend and configuration,
	// but have independent registered events and metadata

	ctx := context.Background()

	// Use package1 emitter (uses static metadata)
	package1Emitter.Count(ctx, "package1.api_request", map[string]interface{}{"endpoint": "/users", "method": "GET"}, 1)

	// Use package2 emitter (uses different static metadata)
	package2Emitter.Count(ctx, "package2.cache_hit", map[string]interface{}{"key": "user:123"}, 1)

	// Use dynamic emitter (uses runtime.Caller)
	dynamicEmitter.Count(ctx, "dynamic.event", nil, 1)

	// Verify each has its own manifest
	_ = package1Emitter.GetManifest() // Contains package1 events
	_ = package2Emitter.GetManifest() // Contains package2 events
	_ = dynamicEmitter.GetManifest()  // Currently empty (no registered events)
}

// MultiPackageSetup demonstrates a more realistic multi-package scenario
// where you might have a global emitter and per-package emitters
type MultiPackageSetup struct {
	global   *emitter.Emitter
	package1 *emitter.Emitter
	package2 *emitter.Emitter
}

func NewMultiPackageSetup() *MultiPackageSetup {
	// Create global emitter with shared backends
	backend := dummy.NewDummyEmitter()
	global := emitter.NewEmitter(backend).
		WithMagicHostname().
		WithMagicPackage()

	// Create package-specific emitters with their own metadata
	pkg1Metadata := map[string]types.CallSiteDetails{
		"api.request": {
			Package:      "github.com/example/api",
			PropertyKeys: []string{"endpoint", "method"},
			MetricType:   "COUNT",
		},
	}

	pkg2Metadata := map[string]types.CallSiteDetails{
		"db.query": {
			Package:      "github.com/example/db",
			PropertyKeys: []string{"table"},
			MetricType:   "TIMER",
		},
	}

	package1 := global.NewSubEmitter().
		WithCallsiteProvider(emitter.StaticCallsiteProvider(pkg1Metadata)).
		WithStaticMetadata(pkg1Metadata)

	package2 := global.NewSubEmitter().
		WithCallsiteProvider(emitter.StaticCallsiteProvider(pkg2Metadata)).
		WithStaticMetadata(pkg2Metadata)

	return &MultiPackageSetup{
		global:   global,
		package1: package1,
		package2: package2,
	}
}
