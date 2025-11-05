# Static Call Site Generator - TODO

This document tracks future enhancements for the `emitter-gen` static call site generator.

## Current Status (Phase 1 - Complete)

✅ Single package mode
✅ AST walking to find emitter calls
✅ Extract event names from direct calls
✅ Extract event names from `Metric()` and `Log()` registration
✅ Handle `LogFnCallsite()` and `MetricFnCallsite()` decorators
✅ Enforce event name uniqueness per package
✅ Require event names to be string literals
✅ Generate `map[string]types.CallSiteDetails`
✅ Command-line flags: `-o`, `-var`, `-package`
✅ Single directory input

## Phase 2 - Multi-Package Support

### Recursive Package Discovery
- [ ] Support `./...` pattern to find all packages recursively
- [ ] Support multiple package patterns (e.g., `./pkg1 ./pkg2 pkg3/...`)
- [ ] Generate one output file per package
- [ ] Use package name to determine default output filename if `-o` not specified

### Implementation Notes
- When given `./...` or similar patterns:
  1. Use `packages.Load()` with the pattern
  2. For each package found that contains emitter calls:
     - Generate output file in that package's directory
     - Use `-var` value (or default) for variable name
     - Use package's name for package declaration
- Handle package name conflicts (multiple packages with same name in different paths)
- Consider adding `-recursive` flag vs auto-detecting from pattern

## Phase 3 - Enhanced Error Reporting

### Better Diagnostics
- [ ] Report all duplicate event names at once (don't stop at first error)
- [ ] Show suggestion for nested `*FnCallsite` errors
- [ ] Warn about event names that look like they should be constants
- [ ] Detect and warn about similar event names (typos)

### Validation
- [ ] Validate that event names follow a naming convention (configurable)
- [ ] Check for unused registered metrics/logs
- [ ] Detect metrics that are registered but never called

## Phase 4 - Advanced Features

### Constant Support
- [ ] Support event names defined as constants
- [ ] Trace constant definitions across package boundaries
- [ ] Handle const blocks and iota patterns

### Configuration File
- [ ] Support `.emitter-gen.yaml` or similar for:
  - Default flags
  - Naming conventions
  - Excluded files/directories
  - Custom validation rules

### IDE Integration
- [ ] Generate LSP-compatible error messages
- [ ] Add `//go:generate` template examples to README
- [ ] Support watch mode for development

## Phase 5 - Optimization

### Performance
- [ ] Parallel package processing
- [ ] Incremental regeneration (only changed packages)
- [ ] Cache AST parsing results

### Output
- [ ] Add option to generate const declarations for event names
- [ ] Generate helper functions for common patterns
- [ ] Add metadata comments to generated code (last generated time, tool version, etc.)

## Phase 6 - Testing

### Generator Tests
- [ ] Unit tests for AST extraction logic
- [ ] Integration tests with real Go code
- [ ] Test error cases (duplicate names, non-literal strings, etc.)
- [ ] Test all emitter method types
- [ ] Test nested package scenarios

### Test Fixtures
- [ ] Create example packages with various patterns
- [ ] Test with complex codebases
- [ ] Benchmark generator performance

## Phase 7 - Documentation

### User Guide
- [ ] Add comprehensive README to `cmd/emitter-gen/`
- [ ] Document all command-line flags
- [ ] Provide usage examples for common scenarios
- [ ] Document integration with build systems (Make, Bazel, etc.)

### Developer Guide
- [ ] Document AST walking strategy
- [ ] Explain event name extraction logic
- [ ] Provide guide for adding new emitter method types
- [ ] Document extension points

## Known Limitations

### Current Constraints
- Event names must be string literals (cannot use variables or constants)
- Does not handle cross-package event registration
- Single package per invocation
- No support for generic/templated event names

### Future Considerations
- Should we support environment-specific event names?
- How to handle event names generated at build time?
- Integration with metric aggregation services?

## Migration Path

### From Dynamic to Static
1. Run `emitter-gen` on existing codebase
2. Review generated call sites for accuracy
3. Create static provider function:
   ```go
   func staticCallsiteProvider(eventName string) types.CallSiteDetails {
       if details, ok := emitterCallSiteDetails[eventName]; ok {
           return details
       }
       // Fallback or error
       return types.CallSiteDetails{}
   }
   ```
4. Replace `WithCallsiteProvider(defaultCallsiteProvider)` with static provider
5. Test thoroughly
6. Remove `WithMagic*` calls (no longer needed)

### Backwards Compatibility
- Ensure dynamic mode continues to work
- Document when to use each mode
- Provide benchmarks showing performance difference
