package emitter

import t "github.com/pseudofunctor-ai/go-emitter/emitter/types"

// NewStaticCallsiteProvider creates a callsite provider that looks up call sites from a static map.
// This eliminates the runtime overhead of runtime.Caller() by using pre-generated call site details.
//
// Usage:
//
//  1. Generate call site details using emitter-gen:
//     emitter-gen -o emitter_callsites.go -var emitterCallSiteDetails ./path/to/package
//
//  2. Create an emitter with the static provider:
//     emitter := NewEmitter(backends...).
//     WithCallsiteProvider(NewStaticCallsiteProvider(emitterCallSiteDetails))
//
// If an event name is not found in the map, an empty CallSiteDetails is returned.
// Consider logging this as an error or falling back to defaultCallsiteProvider.
func NewStaticCallsiteProvider(callsites map[string]t.CallSiteDetails) func(eventName string) t.CallSiteDetails {
	return func(eventName string) t.CallSiteDetails {
		if details, ok := callsites[eventName]; ok {
			return details
		}
		// Return empty details if event not found
		// In production, you might want to log an error here
		return t.CallSiteDetails{}
	}
}

// NewStaticEmitter creates an emitter that uses pre-generated static call site details.
// This is a decorator pattern that wraps an existing emitter and replaces its memoization
// table with the static call sites.
//
// Usage:
//
//  1. Generate call site details using emitter-gen:
//     emitter-gen -o emitter_callsites.go -var emitterCallSiteDetails ./path/to/package
//
//  2. Create a base emitter with your backends and configuration:
//     baseEmitter := NewEmitter(backends...)
//
//  3. Wrap it with static call sites:
//     emitter := NewStaticEmitter(baseEmitter, emitterCallSiteDetails)
//
// The static emitter pre-populates the memoization table, so the first call to each
// event uses the pre-generated details instead of calling runtime.Caller().
func NewStaticEmitter(base *Emitter, callsites map[string]t.CallSiteDetails) *Emitter {
	// Convert CallSiteDetails to eventCallSiteProps and populate memoTable
	for eventName, details := range callsites {
		// Get hostname from the base emitter's provider
		hostname, _ := base.hostname_provider()

		base.memoTable[eventName] = eventCallSiteProps{
			hostname: hostname,
			filename: details.Filename,
			lineNo:   details.LineNo,
			funcName: details.FuncName,
			package_: details.Package,
		}
	}

	return base
}
