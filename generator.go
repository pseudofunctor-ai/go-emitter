package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

type GeneratorConfig struct {
	Directory  string
	OutputFile string
	VarName    string
	PkgName    string
}

type CallSite struct {
	EventName    string
	Filename     string
	LineNo       int
	FuncName     string
	Package      string
	PropertyKeys []string
	MetricType   string
}

// Generate is the main entry point for the generator
func Generate(config GeneratorConfig) error {
	pkg, err := loadPackage(config.Directory)
	if err != nil {
		return fmt.Errorf("loading package: %w", err)
	}

	packageName := config.PkgName
	if packageName == "" {
		packageName = pkg.Name
	}

	callsites, err := extractCallSites(pkg)
	if err != nil {
		return fmt.Errorf("extracting call sites: %w", err)
	}

	// Use OutputFile as-is if it contains path separators or is absolute
	// Otherwise, join with Directory to create output in the package directory
	outputPath := config.OutputFile
	if !filepath.IsAbs(outputPath) && filepath.Base(outputPath) == outputPath {
		// OutputFile is just a filename with no directory component
		outputPath = filepath.Join(config.Directory, config.OutputFile)
	}
	return writeOutputFile(outputPath, packageName, config.VarName, callsites)
}

// loadPackage loads a single Go package from the given directory
func loadPackage(dir string) (*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
		Dir:  dir,
	}

	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return nil, err
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages found in %s", dir)
	}

	if len(pkgs) > 1 {
		return nil, fmt.Errorf("multiple packages found in %s", dir)
	}

	pkg := pkgs[0]
	if len(pkg.Errors) > 0 {
		return nil, fmt.Errorf("package has errors: %v", pkg.Errors)
	}

	return pkg, nil
}

// extractCallSites walks the AST and extracts all emitter call sites
func extractCallSites(pkg *packages.Package) (map[string]CallSite, error) {
	extractor := &callSiteExtractor{
		pkg:       pkg,
		callsites: make(map[string]CallSite),
	}

	for _, file := range pkg.Syntax {
		// Use the file position to get the filename
		filename := pkg.Fset.Position(file.Pos()).Filename
		extractor.currentFile = filename
		ast.Walk(extractor, file)
	}

	return extractor.callsites, extractor.err
}

type callSiteExtractor struct {
	pkg         *packages.Package
	currentFile string
	callsites   map[string]CallSite
	err         error
}

// Visit implements ast.Visitor
func (e *callSiteExtractor) Visit(node ast.Node) ast.Visitor {
	if e.err != nil {
		return nil
	}

	callExpr, ok := node.(*ast.CallExpr)
	if !ok {
		return e
	}

	// First check if this is a callback invocation (calling a variable that holds a MetricEmitterFn/LogEmitterFn)
	if eventName, callsite := e.extractCallbackInvocation(callExpr); eventName != "" {
		// This is a direct callback invocation - record it
		e.recordCallsite(eventName, callsite, false)
		return e
	}

	// Check if this is an emitter method call (em.Count, em.Metric, etc.)
	eventName, callsite, isDecorator := e.extractFromCall(callExpr)
	if eventName == "" {
		return e
	}

	// Only record callsites for:
	// 1. Decorators (*FnCallsite)
	// 2. Direct emitter method calls (em.Count, em.InfoContext, etc.) - NOT em.Metric/em.Log
	if isDecorator {
		e.recordCallsite(eventName, callsite, true)
	} else {
		// Check if this is a registration method (Metric, MetricWithProps, Log, LogWithProps)
		// These should NOT generate callsites - only their invocations should
		methodName := e.getMethodName(callExpr)
		if !isRegistrationMethod(methodName) {
			// This is a direct emitter call like em.Count() - record it
			e.recordCallsite(eventName, callsite, false)
		}
	}

	return e
}

// recordCallsite records a callsite, handling duplicates appropriately
func (e *callSiteExtractor) recordCallsite(eventName string, callsite CallSite, isDecorator bool) {
	if existing, found := e.callsites[eventName]; found {
		// Allow decorators to override existing call sites
		if !isDecorator {
			e.err = fmt.Errorf("duplicate event name %q: already defined at %s:%d, found again at %s:%d",
				eventName, existing.Filename, existing.LineNo, callsite.Filename, callsite.LineNo)
			return
		}
		// Decorator overrides the existing call site
	}

	e.callsites[eventName] = callsite
}

// getMethodName extracts the method name from a call expression
func (e *callSiteExtractor) getMethodName(call *ast.CallExpr) string {
	selExpr, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	return selExpr.Sel.Name
}

// isRegistrationMethod returns true if the method creates a callback (Metric, Log, etc.)
func isRegistrationMethod(methodName string) bool {
	return methodName == "Metric" || methodName == "MetricWithProps" ||
		methodName == "Log" || methodName == "LogWithProps"
}

// extractCallbackInvocation detects when a registered callback (MetricEmitterFn/LogEmitterFn) is being invoked
// Returns: eventName, callsite (empty strings if not a callback invocation)
func (e *callSiteExtractor) extractCallbackInvocation(call *ast.CallExpr) (string, CallSite) {
	// Try direct identifier invocation first (e.g., userLoginMetric(...))
	if ident, ok := call.Fun.(*ast.Ident); ok {
		// Look up the identifier in the type info to see if it's a registered callback
		obj := e.pkg.TypesInfo.Uses[ident]
		if obj == nil {
			return "", CallSite{}
		}

		// Check if this identifier refers to a variable that was assigned from Metric/Log
		// We need to find where this variable was defined and extract the event name
		callsite := e.findCallSiteFromDefinition(obj.Pos())
		if callsite.EventName == "" {
			return "", CallSite{}
		}

		// Update the callsite to reflect this invocation location, not the definition
		pos := e.pkg.Fset.Position(call.Pos())
		funcName := e.findEnclosingFunc(call.Pos())

		return callsite.EventName, CallSite{
			EventName:    callsite.EventName,
			Filename:     pos.Filename,
			LineNo:       pos.Line,
			FuncName:     funcName,
			Package:      e.pkg.PkgPath,
			PropertyKeys: callsite.PropertyKeys,
			MetricType:   callsite.MetricType,
		}
	}

	// Try selector expression invocation (e.g., tem.event1(...))
	if selExpr, ok := call.Fun.(*ast.SelectorExpr); ok {
		return e.extractCallbackFromSelector(call, selExpr)
	}

	// Try index expression invocation (e.g., slice[0](...) or m["key"](...))
	if indexExpr, ok := call.Fun.(*ast.IndexExpr); ok {
		return e.extractCallbackFromIndex(call, indexExpr)
	}

	return "", CallSite{}
}

// extractCallbackFromSelector handles callback invocations via struct field access (e.g., tem.event1(...))
func (e *callSiteExtractor) extractCallbackFromSelector(call *ast.CallExpr, selExpr *ast.SelectorExpr) (string, CallSite) {
	// Check if this is a callback type
	tv, ok := e.pkg.TypesInfo.Types[selExpr]
	if !ok {
		return "", CallSite{}
	}

	// Check if it's a MetricEmitterFn or LogEmitterFn type
	if !e.isCallbackType(tv.Type) {
		return "", CallSite{}
	}

	// We need to trace back to find where this field was initialized
	// Use TypesInfo to find the exact field being referenced (handles name collisions)
	var callsite CallSite
	if obj := e.pkg.TypesInfo.Uses[selExpr.Sel]; obj != nil {
		// For struct fields, use type-aware lookup
		if fieldVar, ok := obj.(*types.Var); ok && fieldVar.IsField() {
			callsite = e.findCallbackFieldInitializationByType(fieldVar)
		} else {
			// For other cases, trace from the definition position
			callsite = e.findCallSiteFromDefinition(obj.Pos())
		}
	}

	// Fallback to name-based search if TypesInfo lookup failed
	if callsite.EventName == "" {
		fieldName := selExpr.Sel.Name
		callsite = e.findCallbackFieldInitialization(fieldName)
	}

	if callsite.EventName == "" {
		return "", CallSite{}
	}

	// Update the callsite to reflect this invocation location
	pos := e.pkg.Fset.Position(call.Pos())
	funcName := e.findEnclosingFunc(call.Pos())

	return callsite.EventName, CallSite{
		EventName:    callsite.EventName,
		Filename:     pos.Filename,
		LineNo:       pos.Line,
		FuncName:     funcName,
		Package:      e.pkg.PkgPath,
		PropertyKeys: callsite.PropertyKeys,
		MetricType:   callsite.MetricType,
	}
}

// extractCallbackFromIndex handles callback invocations via array/slice/map index (e.g., slice[0](...) or m["key"](...))
func (e *callSiteExtractor) extractCallbackFromIndex(call *ast.CallExpr, indexExpr *ast.IndexExpr) (string, CallSite) {
	pos := e.pkg.Fset.Position(call.Pos())
	var funcName string

	// Check the type of the indexed expression
	tv, ok := e.pkg.TypesInfo.Types[indexExpr]
	if !ok {
		return "", CallSite{}
	}

	// Verify the result is a callback type
	if !e.isCallbackType(tv.Type) {
		return "", CallSite{}
	}

	// Get the identifier being indexed (the slice/map variable)
	switch x := indexExpr.X.(type) {
	case *ast.Ident:
		// Look up where this variable was defined
		obj := e.pkg.TypesInfo.Uses[x]
		if obj == nil {
			return "", CallSite{}
		}

		// Find where this variable is initialized and check if it's from a function call or composite literal
		funcName, isFromFunction := e.findVariableInitFunction(obj.Pos())
		if isFromFunction {
			// Handle as function result
			callsite := e.findCallbackInFunctionReturn(funcName, indexExpr.Index)
			if callsite.EventName == "" {
				return "", CallSite{}
			}

			// Update the callsite to reflect this invocation location
			pos = e.pkg.Fset.Position(call.Pos())
			funcName = e.findEnclosingFunc(call.Pos())

			return callsite.EventName, CallSite{
				EventName:    callsite.EventName,
				Filename:     pos.Filename,
				LineNo:       pos.Line,
				FuncName:     funcName,
				Package:      e.pkg.PkgPath,
				PropertyKeys: callsite.PropertyKeys,
				MetricType:   callsite.MetricType,
			}
		}

		// Find the initialization of this slice/map from a composite literal
		callsite := e.findCallbackInCollection(obj.Pos(), indexExpr.Index)
		if callsite.EventName == "" {
			return "", CallSite{}
		}

		// Update the callsite to reflect this invocation location
		pos = e.pkg.Fset.Position(call.Pos())
		funcName = e.findEnclosingFunc(call.Pos())

		return callsite.EventName, CallSite{
			EventName:    callsite.EventName,
			Filename:     pos.Filename,
			LineNo:       pos.Line,
			FuncName:     funcName,
			Package:      e.pkg.PkgPath,
			PropertyKeys: callsite.PropertyKeys,
			MetricType:   callsite.MetricType,
		}

	case *ast.SelectorExpr:
		// Handle struct field arrays like obj.callbacks[0]()
		// We need to trace where this struct field was initialized

		fieldName := x.Sel.Name

		// Try to find the struct type and field initialization
		callsite := e.findCallbackInStructFieldArray(fieldName, indexExpr.Index)
		if callsite.EventName == "" {
			return "", CallSite{}
		}

		// Update the callsite to reflect this invocation location
		pos = e.pkg.Fset.Position(call.Pos())
		funcName = e.findEnclosingFunc(call.Pos())

		return callsite.EventName, CallSite{
			EventName:    callsite.EventName,
			Filename:     pos.Filename,
			LineNo:       pos.Line,
			FuncName:     funcName,
			Package:      e.pkg.PkgPath,
			PropertyKeys: callsite.PropertyKeys,
			MetricType:   callsite.MetricType,
		}

	case *ast.CallExpr:
		// Handle function call results like emittersInSlice(em)[0]
		// We need to find the function and analyze its return value
		return e.extractCallbackFromIndexedFunctionResult(call, indexExpr, x)
	default:
		return "", CallSite{}
	}
}

// extractCallbackFromIndexedFunctionResult handles cases like emittersInSlice(em)[0](...)
func (e *callSiteExtractor) extractCallbackFromIndexedFunctionResult(call *ast.CallExpr, indexExpr *ast.IndexExpr, funcCall *ast.CallExpr) (string, CallSite) {
	// Get the function being called
	var funcIdent *ast.Ident
	if selExpr, ok := funcCall.Fun.(*ast.SelectorExpr); ok {
		funcIdent = selExpr.Sel
	} else if ident, ok := funcCall.Fun.(*ast.Ident); ok {
		funcIdent = ident
	} else {
		return "", CallSite{}
	}

	// Find the function definition
	callsite := e.findCallbackInFunctionReturn(funcIdent.Name, indexExpr.Index)
	if callsite.EventName == "" {
		return "", CallSite{}
	}

	// Update the callsite to reflect this invocation location
	pos := e.pkg.Fset.Position(call.Pos())
	funcName := e.findEnclosingFunc(call.Pos())

	return callsite.EventName, CallSite{
		EventName:    callsite.EventName,
		Filename:     pos.Filename,
		LineNo:       pos.Line,
		FuncName:     funcName,
		Package:      e.pkg.PkgPath,
		PropertyKeys: callsite.PropertyKeys,
		MetricType:   callsite.MetricType,
	}
}

// findCallbackInCollection finds the callback at a specific index in a slice/array or key in a map
func (e *callSiteExtractor) findCallbackInCollection(defPos token.Pos, indexExpr ast.Expr) CallSite {
	// Find the file containing this definition
	var targetFile *ast.File
	for _, file := range e.pkg.Syntax {
		if e.pkg.Fset.Position(file.Pos()).Filename == e.pkg.Fset.Position(defPos).Filename {
			targetFile = file
			break
		}
	}

	if targetFile == nil {
		return CallSite{}
	}

	// Extract the index value
	var index int
	var mapKey string
	isMapLookup := false

	switch idx := indexExpr.(type) {
	case *ast.BasicLit:
		if idx.Kind == token.INT {
			fmt.Sscanf(idx.Value, "%d", &index)
		} else if idx.Kind == token.STRING {
			mapKey = strings.Trim(idx.Value, `"`)
			isMapLookup = true
		}
	default:
		// Can't handle dynamic indices
		return CallSite{}
	}

	// Find the assignment/declaration
	var result CallSite
	ast.Inspect(targetFile, func(n ast.Node) bool {
		if result.EventName != "" {
			return false
		}

		switch node := n.(type) {
		case *ast.AssignStmt:
			if node.Pos() <= defPos && defPos <= node.End() {
				result = e.extractCallbackFromCollectionAssignment(node, index, mapKey, isMapLookup)
			}
		case *ast.ValueSpec:
			if node.Pos() <= defPos && defPos <= node.End() {
				result = e.extractCallbackFromCollectionValueSpec(node, index, mapKey, isMapLookup)
			}
		}

		return result.EventName == ""
	})

	return result
}

// findVariableInitFunction checks if a variable is initialized from a function call
// Returns (functionName, true) if initialized from a function call, ("", false) otherwise
func (e *callSiteExtractor) findVariableInitFunction(varPos token.Pos) (string, bool) {
	// Find the file containing this variable definition
	var targetFile *ast.File
	for _, file := range e.pkg.Syntax {
		if e.pkg.Fset.Position(file.Pos()).Filename == e.pkg.Fset.Position(varPos).Filename {
			targetFile = file
			break
		}
	}

	if targetFile == nil {
		return "", false
	}

	var funcName string
	var found bool

	ast.Inspect(targetFile, func(n ast.Node) bool {
		if found {
			return false
		}

		switch node := n.(type) {
		case *ast.AssignStmt:
			// Check if this assignment contains our variable
			if node.Pos() <= varPos && varPos <= node.End() {
				// Check if RHS is a function call
				if len(node.Rhs) == 1 {
					if callExpr, ok := node.Rhs[0].(*ast.CallExpr); ok {
						// Extract function name
						funcName = e.extractFunctionName(callExpr)
						if funcName != "" {
							found = true
							return false
						}
					}
				}
			}
		case *ast.ValueSpec:
			// Check if this value spec contains our variable
			if node.Pos() <= varPos && varPos <= node.End() {
				// Check if value is a function call
				if len(node.Values) == 1 {
					if callExpr, ok := node.Values[0].(*ast.CallExpr); ok {
						// Extract function name
						funcName = e.extractFunctionName(callExpr)
						if funcName != "" {
							found = true
							return false
						}
					}
				}
			}
		}

		return true
	})

	return funcName, found
}

// extractFunctionName extracts the function name from a call expression
func (e *callSiteExtractor) extractFunctionName(callExpr *ast.CallExpr) string {
	switch fun := callExpr.Fun.(type) {
	case *ast.Ident:
		return fun.Name
	case *ast.SelectorExpr:
		return fun.Sel.Name
	default:
		return ""
	}
}

// findCallbackInFunctionReturn finds a callback in a function's return statement
func (e *callSiteExtractor) findCallbackInFunctionReturn(funcName string, indexExpr ast.Expr) CallSite {
	// Extract the index value
	var index int
	var mapKey string
	isMapLookup := false

	switch idx := indexExpr.(type) {
	case *ast.BasicLit:
		if idx.Kind == token.INT {
			fmt.Sscanf(idx.Value, "%d", &index)
		} else if idx.Kind == token.STRING {
			mapKey = strings.Trim(idx.Value, `"`)
			isMapLookup = true
		}
	default:
		return CallSite{}
	}

	// Search for the function definition
	for _, file := range e.pkg.Syntax {
		var result CallSite
		ast.Inspect(file, func(n ast.Node) bool {
			if result.EventName != "" {
				return false
			}

			if funcDecl, ok := n.(*ast.FuncDecl); ok && funcDecl.Name.Name == funcName {
				// Look for return statements
				ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
					if result.EventName != "" {
						return false
					}

					if retStmt, ok := n.(*ast.ReturnStmt); ok && len(retStmt.Results) > 0 {
						// Check if it's a composite literal (slice or map)
						if compLit, ok := retStmt.Results[0].(*ast.CompositeLit); ok {
							result = e.extractCallbackFromCompositeLiteral(compLit, index, mapKey, isMapLookup)
						}
					}
					return result.EventName == ""
				})
			}

			return result.EventName == ""
		})

		if result.EventName != "" {
			return result
		}
	}

	return CallSite{}
}

// extractCallbackFromCollectionAssignment extracts callback from slice/map assignment
func (e *callSiteExtractor) extractCallbackFromCollectionAssignment(assign *ast.AssignStmt, index int, mapKey string, isMapLookup bool) CallSite {
	if len(assign.Rhs) != 1 {
		return CallSite{}
	}

	compLit, ok := assign.Rhs[0].(*ast.CompositeLit)
	if !ok {
		return CallSite{}
	}

	return e.extractCallbackFromCompositeLiteral(compLit, index, mapKey, isMapLookup)
}

// extractCallbackFromCollectionValueSpec extracts callback from slice/map var declaration
func (e *callSiteExtractor) extractCallbackFromCollectionValueSpec(spec *ast.ValueSpec, index int, mapKey string, isMapLookup bool) CallSite {
	if len(spec.Values) != 1 {
		return CallSite{}
	}

	compLit, ok := spec.Values[0].(*ast.CompositeLit)
	if !ok {
		return CallSite{}
	}

	return e.extractCallbackFromCompositeLiteral(compLit, index, mapKey, isMapLookup)
}

// extractCallbackFromCompositeLiteral extracts a callback from a composite literal (slice/map)
func (e *callSiteExtractor) extractCallbackFromCompositeLiteral(compLit *ast.CompositeLit, index int, mapKey string, isMapLookup bool) CallSite {
	if isMapLookup {
		// Map lookup by key
		for _, elt := range compLit.Elts {
			if kv, ok := elt.(*ast.KeyValueExpr); ok {
				// Check if the key matches
				if lit, ok := kv.Key.(*ast.BasicLit); ok {
					if lit.Kind == token.STRING && strings.Trim(lit.Value, `"`) == mapKey {
						// Found the matching key
						if callExpr, ok := kv.Value.(*ast.CallExpr); ok {
							return e.extractCallbackFromCallExpr(callExpr)
						}
					}
				}
			}
		}
	} else {
		// Array/slice lookup by index
		if index < len(compLit.Elts) {
			elt := compLit.Elts[index]
			if callExpr, ok := elt.(*ast.CallExpr); ok {
				return e.extractCallbackFromCallExpr(callExpr)
			}
		}
	}

	return CallSite{}
}

// extractCallbackFromCallExpr extracts callback metadata from a call expression
func (e *callSiteExtractor) extractCallbackFromCallExpr(callExpr *ast.CallExpr) CallSite {
	selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return CallSite{}
	}

	methodName := selExpr.Sel.Name
	if methodName == "Metric" || methodName == "MetricWithProps" ||
		methodName == "Log" || methodName == "LogWithProps" {
		eventName := e.extractEventNameArg(callExpr, methodName)
		if eventName != "" {
			callsite := e.makeCallSite(callExpr, eventName)
			callsite.PropertyKeys = e.extractPropertyKeys(callExpr, methodName)
			callsite.MetricType = e.extractMetricType(callExpr, methodName)
			return callsite
		}
	}

	return CallSite{}
}

// findCallbackInStructFieldArray finds a callback in a struct field that's an array/slice/map
func (e *callSiteExtractor) findCallbackInStructFieldArray(fieldName string, indexExpr ast.Expr) CallSite {
	// Extract the index value
	var index int
	var mapKey string
	isMapLookup := false

	switch idx := indexExpr.(type) {
	case *ast.BasicLit:
		if idx.Kind == token.INT {
			fmt.Sscanf(idx.Value, "%d", &index)
		} else if idx.Kind == token.STRING {
			mapKey = strings.Trim(idx.Value, `"`)
			isMapLookup = true
		}
	default:
		// Can't handle dynamic indices
		return CallSite{}
	}

	// Search through all files in the package for struct field initializations
	for _, file := range e.pkg.Syntax {
		var result CallSite
		ast.Inspect(file, func(n ast.Node) bool {
			if result.EventName != "" {
				return false
			}

			// Look for composite literals (struct initialization)
			if compLit, ok := n.(*ast.CompositeLit); ok {
				for _, elt := range compLit.Elts {
					if kv, ok := elt.(*ast.KeyValueExpr); ok {
						// Check if the key matches our field name
						if ident, ok := kv.Key.(*ast.Ident); ok && ident.Name == fieldName {
							// Check if the value is a composite literal (array/slice/map)
							if arrayLit, ok := kv.Value.(*ast.CompositeLit); ok {
								result = e.extractCallbackFromCompositeLiteral(arrayLit, index, mapKey, isMapLookup)
								return false
							}
							// Check if the value is a function call
							if callExpr, ok := kv.Value.(*ast.CallExpr); ok {
								funcName := e.extractFunctionName(callExpr)
								if funcName != "" {
									result = e.findCallbackInFunctionReturn(funcName, indexExpr)
									return false
								}
							}
						}
					}
				}
			}

			return true
		})

		if result.EventName != "" {
			return result
		}
	}

	return CallSite{}
}

// isCallbackType checks if a type is MetricEmitterFn or LogEmitterFn
func (e *callSiteExtractor) isCallbackType(typ types.Type) bool {
	named, ok := typ.(*types.Named)
	if !ok {
		return false
	}

	obj := named.Obj()
	if obj == nil {
		return false
	}

	pkg := obj.Pkg()
	if pkg == nil || pkg.Path() != "github.com/pseudofunctor-ai/go-emitter/emitter/types" {
		return false
	}

	name := obj.Name()
	return name == "MetricEmitterFn" || name == "LogEmitterFn"
}

// findCallbackFieldInitializationByType searches for where a specific struct field (by type and name)
// was initialized with a callback (Metric/Log call). This handles field name collisions by using type info.
func (e *callSiteExtractor) findCallbackFieldInitializationByType(fieldObj *types.Var) CallSite {
	if fieldObj == nil {
		return CallSite{}
	}

	fieldName := fieldObj.Name()

	// Get the struct type that contains this field
	// fieldObj.Type() is the field's type (LogEmitterFn), we need the containing struct
	// We can identify the correct struct by checking if the field position matches
	fieldPos := fieldObj.Pos()

	// Search for composite literals that initialize this specific field
	for _, file := range e.pkg.Syntax {
		var result CallSite
		ast.Inspect(file, func(n ast.Node) bool {
			if result.EventName != "" {
				return false
			}

			if compLit, ok := n.(*ast.CompositeLit); ok {
				// Check if this composite literal's type matches the struct containing our field
				for _, elt := range compLit.Elts {
					if kv, ok := elt.(*ast.KeyValueExpr); ok {
						if ident, ok := kv.Key.(*ast.Ident); ok && ident.Name == fieldName {
							// Verify this is the right field by checking if it refers to the same object
							if obj := e.pkg.TypesInfo.Uses[ident]; obj != nil && obj.Pos() == fieldPos {
								// This is the correct field initialization
								if callExpr, ok := kv.Value.(*ast.CallExpr); ok {
									if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
										methodName := selExpr.Sel.Name
										if methodName == "Metric" || methodName == "MetricWithProps" ||
											methodName == "Log" || methodName == "LogWithProps" {
											eventName := e.extractEventNameArg(callExpr, methodName)
											if eventName != "" {
												result = e.makeCallSite(callExpr, eventName)
												result.PropertyKeys = e.extractPropertyKeys(callExpr, methodName)
												result.MetricType = e.extractMetricType(callExpr, methodName)
												return false
											}
										}
									}
								}
							}
						}
					}
				}
			}
			return true
		})

		if result.EventName != "" {
			return result
		}
	}

	return CallSite{}
}

// findCallbackFieldInitialization searches for where a struct field with the given name
// was initialized with a callback (Metric/Log call)
func (e *callSiteExtractor) findCallbackFieldInitialization(fieldName string) CallSite {
	// Search through all files in the package for composite literals or return statements
	// that initialize a field with this name
	for _, file := range e.pkg.Syntax {
		var result CallSite
		ast.Inspect(file, func(n ast.Node) bool {
			if result.EventName != "" {
				return false // Already found it
			}

			// Look for composite literals (struct initialization)
			if compLit, ok := n.(*ast.CompositeLit); ok {
				for _, elt := range compLit.Elts {
					if kv, ok := elt.(*ast.KeyValueExpr); ok {
						// Check if the key matches our field name
						if ident, ok := kv.Key.(*ast.Ident); ok && ident.Name == fieldName {
							// Check if the value is a call to Metric/Log
							if callExpr, ok := kv.Value.(*ast.CallExpr); ok {
								if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
									methodName := selExpr.Sel.Name
									if methodName == "Metric" || methodName == "MetricWithProps" ||
										methodName == "Log" || methodName == "LogWithProps" {
										eventName := e.extractEventNameArg(callExpr, methodName)
										if eventName != "" {
											result = e.makeCallSite(callExpr, eventName)
											result.PropertyKeys = e.extractPropertyKeys(callExpr, methodName)
											result.MetricType = e.extractMetricType(callExpr, methodName)
											return false
										}
									}
								}
							}
						}
					}
				}
			}

			return true
		})

		if result.EventName != "" {
			return result
		}
	}

	return CallSite{}
}

// extractFromCall extracts event name and call site from a function call
// Returns: eventName, callsite, isDecorator
func (e *callSiteExtractor) extractFromCall(call *ast.CallExpr) (string, CallSite, bool) {
	selExpr, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", CallSite{}, false
	}

	methodName := selExpr.Sel.Name

	// Handle timer calls first (timer.Time(...))
	if methodName == "Time" {
		eventName, callsite := e.extractTimerCall(call)
		if eventName != "" {
			return eventName, callsite, false
		}
		// Not a timer call, fall through to check emitter methods
	}

	// Verify this is actually an emitter method call by checking the receiver type
	if !e.isEmitterReceiver(selExpr.X) {
		return "", CallSite{}, false
	}

	// Handle direct emitter calls with event parameter
	if isEmitterMethod(methodName) {
		eventName, callsite := e.extractDirectCall(call, methodName)
		return eventName, callsite, false
	}

	// Handle *FnCallsite decorator calls
	if methodName == "LogFnCallsite" || methodName == "MetricFnCallsite" {
		eventName, callsite := e.extractCallsiteDecorator(call, methodName)
		return eventName, callsite, true
	}

	return "", CallSite{}, false
}

// extractDirectCall extracts event name from direct emitter method calls
func (e *callSiteExtractor) extractDirectCall(call *ast.CallExpr, methodName string) (string, CallSite) {
	eventName := e.extractEventNameArg(call, methodName)
	if eventName == "" {
		return "", CallSite{}
	}

	callsite := e.makeCallSite(call, eventName)

	// Extract property keys from props argument
	callsite.PropertyKeys = e.extractPropertyKeys(call, methodName)

	// Extract metric type
	callsite.MetricType = e.extractMetricType(call, methodName)

	return eventName, callsite
}

// extractTimerCall extracts event name from timer.Time() calls
func (e *callSiteExtractor) extractTimerCall(call *ast.CallExpr) (string, CallSite) {
	// Timer.Time signature: Time(ctx context.Context, event string, props map[string]interface{}, fn func() T) T
	// We need at least 4 arguments
	if len(call.Args) < 4 {
		return "", CallSite{}
	}

	// Extract event name from the second argument (index 1)
	lit, ok := call.Args[1].(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		// Event name must be a string literal
		return "", CallSite{}
	}
	eventName := strings.Trim(lit.Value, `"`)

	// Create the call site
	callsite := e.makeCallSite(call, eventName)

	// Extract property keys from the third argument (index 2) - the props map
	propsArg := call.Args[2]
	if compLit, ok := propsArg.(*ast.CompositeLit); ok {
		// Extract keys from the composite literal
		keys := make([]string, 0)
		for _, elt := range compLit.Elts {
			if kv, ok := elt.(*ast.KeyValueExpr); ok {
				if keyLit, ok := kv.Key.(*ast.BasicLit); ok && keyLit.Kind == token.STRING {
					keys = append(keys, strings.Trim(keyLit.Value, `"`))
				} else if keyIdent, ok := kv.Key.(*ast.Ident); ok {
					keys = append(keys, keyIdent.Name)
				}
			}
		}
		callsite.PropertyKeys = keys
	}

	// Timer calls always have TIMER metric type
	callsite.MetricType = "TIMER"

	return eventName, callsite
}

// extractCallsiteDecorator extracts event name from *FnCallsite decorator calls
func (e *callSiteExtractor) extractCallsiteDecorator(call *ast.CallExpr, decoratorName string) (string, CallSite) {
	if len(call.Args) != 1 {
		return "", CallSite{}
	}

	var callsite CallSite

	// The argument can be either an identifier (e.g., decoratedLog) or a selector (e.g., r.emitters.createSuccess)
	switch arg := call.Args[0].(type) {
	case *ast.Ident:
		// Simple identifier case
		obj := e.pkg.TypesInfo.Uses[arg]
		if obj == nil {
			return "", CallSite{}
		}
		callsite = e.findCallSiteFromDefinition(obj.Pos())

	case *ast.SelectorExpr:
		// Selector expression case (e.g., r.emitters.createSuccess)
		// Use TypesInfo to find the exact field being referenced (handles name collisions)
		if obj := e.pkg.TypesInfo.Uses[arg.Sel]; obj != nil {
			// For struct fields, use type-aware lookup
			if fieldVar, ok := obj.(*types.Var); ok && fieldVar.IsField() {
				callsite = e.findCallbackFieldInitializationByType(fieldVar)
			} else {
				// For other cases, trace from the definition position
				callsite = e.findCallSiteFromDefinition(obj.Pos())
			}
		}

		// Fallback to name-based search if TypesInfo lookup failed
		if callsite.EventName == "" {
			fieldName := arg.Sel.Name
			callsite = e.findCallbackFieldInitialization(fieldName)
		}

	default:
		return "", CallSite{}
	}

	if callsite.EventName == "" {
		return "", CallSite{}
	}

	// Update the call site location to the decorator (but keep the metadata from definition)
	pos := e.pkg.Fset.Position(call.Pos())
	funcName := e.findEnclosingFunc(call.Pos())
	callsite.Filename = pos.Filename
	callsite.LineNo = pos.Line
	callsite.FuncName = funcName

	return callsite.EventName, callsite
}

// findCallSiteFromDefinition finds the full callsite metadata from where a variable is defined
func (e *callSiteExtractor) findCallSiteFromDefinition(pos token.Pos) CallSite {
	// Find the file containing this position
	var targetFile *ast.File
	for _, file := range e.pkg.Syntax {
		if e.pkg.Fset.Position(file.Pos()).Filename == e.pkg.Fset.Position(pos).Filename {
			targetFile = file
			break
		}
	}

	if targetFile == nil {
		return CallSite{}
	}

	// Find the assignment or declaration at this position
	var callsite CallSite
	ast.Inspect(targetFile, func(n ast.Node) bool {
		if callsite.EventName != "" {
			return false
		}

		switch node := n.(type) {
		case *ast.AssignStmt:
			if node.Pos() <= pos && pos <= node.End() {
				callsite = e.extractCallSiteFromAssignment(node)
			}
		case *ast.ValueSpec:
			if node.Pos() <= pos && pos <= node.End() {
				callsite = e.extractCallSiteFromValueSpec(node)
			}
		}

		return callsite.EventName == ""
	})

	return callsite
}

// findEventNameFromDefinition finds the event name from where a variable is defined
func (e *callSiteExtractor) findEventNameFromDefinition(pos token.Pos) string {
	callsite := e.findCallSiteFromDefinition(pos)
	return callsite.EventName
}

// extractCallSiteFromAssignment extracts full callsite from an assignment like `foo := emitter.Metric("event_name", COUNT)`
// or traces through callback references like `cb := s.emitters.validateSuccess`
func (e *callSiteExtractor) extractCallSiteFromAssignment(assign *ast.AssignStmt) CallSite {
	if len(assign.Rhs) != 1 {
		return CallSite{}
	}

	// Case 1: RHS is a call expression that creates a new callback
	if call, ok := assign.Rhs[0].(*ast.CallExpr); ok {
		selExpr, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return CallSite{}
		}

		methodName := selExpr.Sel.Name
		if methodName == "Metric" || methodName == "Log" || methodName == "MetricWithProps" || methodName == "LogWithProps" {
			eventName := e.extractEventNameArg(call, methodName)
			if eventName == "" {
				return CallSite{}
			}

			callsite := e.makeCallSite(call, eventName)
			callsite.PropertyKeys = e.extractPropertyKeys(call, methodName)
			callsite.MetricType = e.extractMetricType(call, methodName)
			return callsite
		}
	}

	// Case 2: RHS is a reference to an existing callback (identifier or selector)
	// Examples: `cb := existingCallback` or `cb := s.emitters.validateSuccess`
	// We need to trace what the RHS refers to
	rhsExpr := assign.Rhs[0]

	// For identifiers, use TypesInfo to find what they refer to
	if ident, ok := rhsExpr.(*ast.Ident); ok {
		obj := e.pkg.TypesInfo.Uses[ident]
		if obj != nil {
			// Recursively trace the callback definition
			return e.findCallSiteFromDefinition(obj.Pos())
		}
	}

	// For selector expressions, trace the field initialization
	if selExpr, ok := rhsExpr.(*ast.SelectorExpr); ok {
		// Use TypesInfo to find the exact field being referenced (handles name collisions)
		if obj := e.pkg.TypesInfo.Uses[selExpr.Sel]; obj != nil {
			// For struct fields, use type-aware lookup
			if fieldVar, ok := obj.(*types.Var); ok && fieldVar.IsField() {
				if callsite := e.findCallbackFieldInitializationByType(fieldVar); callsite.EventName != "" {
					return callsite
				}
			} else {
				// For other cases, trace from the definition position
				if callsite := e.findCallSiteFromDefinition(obj.Pos()); callsite.EventName != "" {
					return callsite
				}
			}
		}

		// Fallback to name-based search if TypesInfo lookup failed
		fieldName := selExpr.Sel.Name
		return e.findCallbackFieldInitialization(fieldName)
	}

	return CallSite{}
}

// extractEventNameFromAssignment extracts event name from an assignment like `foo := emitter.Metric("event_name", COUNT)`
func (e *callSiteExtractor) extractEventNameFromAssignment(assign *ast.AssignStmt) string {
	callsite := e.extractCallSiteFromAssignment(assign)
	return callsite.EventName
}

// extractCallSiteFromValueSpec extracts full callsite from a var declaration
// or traces through callback references like `var cb = s.emitters.validateSuccess`
func (e *callSiteExtractor) extractCallSiteFromValueSpec(spec *ast.ValueSpec) CallSite {
	if len(spec.Values) != 1 {
		return CallSite{}
	}

	// Case 1: Value is a call expression that creates a new callback
	if call, ok := spec.Values[0].(*ast.CallExpr); ok {
		selExpr, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return CallSite{}
		}

		methodName := selExpr.Sel.Name
		if methodName == "Metric" || methodName == "Log" || methodName == "MetricWithProps" || methodName == "LogWithProps" {
			eventName := e.extractEventNameArg(call, methodName)
			if eventName == "" {
				return CallSite{}
			}

			callsite := e.makeCallSite(call, eventName)
			callsite.PropertyKeys = e.extractPropertyKeys(call, methodName)
			callsite.MetricType = e.extractMetricType(call, methodName)
			return callsite
		}
	}

	// Case 2: Value is a reference to an existing callback
	// Examples: `var cb = existingCallback` or `var cb = s.emitters.validateSuccess`
	valueExpr := spec.Values[0]

	// For identifiers, use TypesInfo to find what they refer to
	if ident, ok := valueExpr.(*ast.Ident); ok {
		obj := e.pkg.TypesInfo.Uses[ident]
		if obj != nil {
			// Recursively trace the callback definition
			return e.findCallSiteFromDefinition(obj.Pos())
		}
	}

	// For selector expressions, trace the field initialization
	if selExpr, ok := valueExpr.(*ast.SelectorExpr); ok {
		// Use TypesInfo to find the exact field being referenced (handles name collisions)
		if obj := e.pkg.TypesInfo.Uses[selExpr.Sel]; obj != nil {
			// For struct fields, use type-aware lookup
			if fieldVar, ok := obj.(*types.Var); ok && fieldVar.IsField() {
				if callsite := e.findCallbackFieldInitializationByType(fieldVar); callsite.EventName != "" {
					return callsite
				}
			} else {
				// For other cases, trace from the definition position
				if callsite := e.findCallSiteFromDefinition(obj.Pos()); callsite.EventName != "" {
					return callsite
				}
			}
		}

		// Fallback to name-based search if TypesInfo lookup failed
		fieldName := selExpr.Sel.Name
		return e.findCallbackFieldInitialization(fieldName)
	}

	return CallSite{}
}

// extractEventNameFromValueSpec extracts event name from a var declaration
func (e *callSiteExtractor) extractEventNameFromValueSpec(spec *ast.ValueSpec) string {
	callsite := e.extractCallSiteFromValueSpec(spec)
	return callsite.EventName
}

// extractEventNameArg extracts the event name string literal from a call expression
func (e *callSiteExtractor) extractEventNameArg(call *ast.CallExpr, methodName string) string {
	argIndex := getEventNameArgIndex(methodName)
	if argIndex >= len(call.Args) {
		return ""
	}

	lit, ok := call.Args[argIndex].(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		// Event name must be a string literal
		e.err = fmt.Errorf("event name at %s must be a string literal, got %T",
			e.pkg.Fset.Position(call.Args[argIndex].Pos()), call.Args[argIndex])
		return ""
	}

	// Remove quotes from string literal
	return strings.Trim(lit.Value, `"`)
}

// extractPropertyKeys extracts property keys from the props map argument or from WithProps calls
func (e *callSiteExtractor) extractPropertyKeys(call *ast.CallExpr, methodName string) []string {
	// For MetricWithProps and LogWithProps, extract from the propKeys argument
	if methodName == "MetricWithProps" || methodName == "LogWithProps" {
		return e.extractPropKeysFromWithPropsCall(call)
	}

	// For other methods, extract from the props map literal
	propsIndex := getPropsArgIndex(methodName)
	if propsIndex >= len(call.Args) {
		return nil
	}

	propsArg := call.Args[propsIndex]

	// Handle composite literal (map[string]interface{}{...})
	compLit, ok := propsArg.(*ast.CompositeLit)
	if !ok {
		// Props might be a variable reference or nil - we can't extract keys statically
		return nil
	}

	// Extract keys from the composite literal
	keys := make(map[string]struct{})
	for _, elt := range compLit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		// Key should be a string literal or identifier
		var keyStr string
		switch key := kv.Key.(type) {
		case *ast.BasicLit:
			if key.Kind == token.STRING {
				keyStr = strings.Trim(key.Value, `"`)
			}
		case *ast.Ident:
			keyStr = key.Name
		}

		if keyStr != "" {
			keys[keyStr] = struct{}{}
		}
	}

	// Convert map to sorted slice
	result := make([]string, 0, len(keys))
	for key := range keys {
		result = append(result, key)
	}
	sortStrings(result)

	return result
}

// extractMetricType extracts the metric type from method calls
func (e *callSiteExtractor) extractMetricType(call *ast.CallExpr, methodName string) string {
	// For Metric() and MetricWithProps(), the metric type is an explicit argument
	if methodName == "Metric" || methodName == "MetricWithProps" {
		metricTypeIndex := 1 // Second argument
		if metricTypeIndex >= len(call.Args) {
			return ""
		}

		// The metric type should be an identifier like COUNT, GAUGE, etc.
		if sel, ok := call.Args[metricTypeIndex].(*ast.SelectorExpr); ok {
			return sel.Sel.Name
		}
		if ident, ok := call.Args[metricTypeIndex].(*ast.Ident); ok {
			return ident.Name
		}
	}

	// For direct method calls, infer from method name
	return inferMetricTypeFromMethod(methodName)
}

// extractPropKeysFromWithPropsCall extracts property keys from MetricWithProps or LogWithProps
// These methods have propKeys as a slice literal in the third argument (index 2)
func (e *callSiteExtractor) extractPropKeysFromWithPropsCall(call *ast.CallExpr) []string {
	// propKeys is the third argument (index 2) for MetricWithProps/LogWithProps
	if len(call.Args) < 3 {
		return nil
	}

	propsArg := call.Args[2]

	// Handle slice literal []string{...}
	compLit, ok := propsArg.(*ast.CompositeLit)
	if !ok {
		// Props might be a variable reference - we can't extract keys statically
		return nil
	}

	// Extract string values from the slice literal
	keys := make([]string, 0, len(compLit.Elts))
	for _, elt := range compLit.Elts {
		if lit, ok := elt.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			keys = append(keys, strings.Trim(lit.Value, `"`))
		}
	}

	// Sort for stable output
	sortStrings(keys)

	return keys
}

// inferMetricTypeFromMethod maps method names to metric types
func inferMetricTypeFromMethod(methodName string) string {
	switch methodName {
	case "Count":
		return "COUNT"
	case "Gauge":
		return "GAUGE"
	case "Histogram":
		return "HISTOGRAM"
	case "Meter":
		return "METER"
	case "Set":
		return "SET"
	case "Event":
		return "EVENT"
	case "EmitInt", "EmitFloat", "EmitDuration":
		// For Emit methods, we can't infer the type statically
		return ""
	case "Info", "Warn", "Error", "Fatal", "Debug", "Trace",
		"Infof", "Warnf", "Errorf", "Fatalf", "Debugf", "Tracef",
		"InfoContext", "WarnContext", "ErrorContext", "FatalContext", "DebugContext", "TraceContext",
		"InfofContext", "WarnfContext", "ErrorfContext", "FatalfContext", "DebugfContext", "TracefContext",
		"Log", "LogWithProps":
		return "COUNT"
	default:
		return ""
	}
}

// makeCallSite creates a CallSite from an AST node
func (e *callSiteExtractor) makeCallSite(node ast.Node, eventName string) CallSite {
	pos := e.pkg.Fset.Position(node.Pos())

	// Extract function name from position
	funcName := e.findEnclosingFunc(node.Pos())

	return CallSite{
		EventName: eventName,
		Filename:  pos.Filename,
		LineNo:    pos.Line,
		FuncName:  funcName,
		Package:   e.pkg.PkgPath,
	}
}

// isEmitterReceiver checks if the given expression refers to an emitter.Emitter type or CombinedEmitter interface
func (e *callSiteExtractor) isEmitterReceiver(expr ast.Expr) bool {
	// Get the type of the receiver expression
	tv, ok := e.pkg.TypesInfo.Types[expr]
	if !ok {
		// If we don't have type info, we can't verify - skip this call
		return false
	}

	typ := tv.Type

	// Check if it's a pointer type (*Emitter)
	ptr, ok := typ.(*types.Pointer)
	if ok {
		typ = ptr.Elem()
	}

	// Get the named type
	named, ok := typ.(*types.Named)
	if !ok {
		return false
	}

	// Check if it's from the emitter package and is named "Emitter"
	obj := named.Obj()
	if obj == nil {
		return false
	}

	pkg := obj.Pkg()
	if pkg == nil {
		return false
	}

	// Check if this is github.com/pseudofunctor-ai/go-emitter/emitter.Emitter (concrete type)
	if pkg.Path() == "github.com/pseudofunctor-ai/go-emitter/emitter" && obj.Name() == "Emitter" {
		return true
	}

	// Check if this is github.com/pseudofunctor-ai/go-emitter/emitter/types.CombinedEmitter (interface)
	if pkg.Path() == "github.com/pseudofunctor-ai/go-emitter/emitter/types" && obj.Name() == "CombinedEmitter" {
		return true
	}

	return false
}

// findEnclosingFunc finds the name of the function enclosing the given position
func (e *callSiteExtractor) findEnclosingFunc(pos token.Pos) string {
	for _, file := range e.pkg.Syntax {
		var funcName string
		ast.Inspect(file, func(n ast.Node) bool {
			if funcDecl, ok := n.(*ast.FuncDecl); ok {
				if funcDecl.Pos() <= pos && pos <= funcDecl.End() {
					funcName = e.pkg.PkgPath + "." + funcDecl.Name.Name
					return false
				}
			}
			return true
		})
		if funcName != "" {
			return funcName
		}
	}
	return ""
}

// isEmitterMethod returns true if the method name is an emitter method that takes an event name
func isEmitterMethod(name string) bool {
	methods := []string{
		// Direct metric/log methods
		"Count", "Gauge", "Histogram", "Meter", "Set", "Event",
		// Logger methods
		"Info", "Warn", "Error", "Fatal", "Debug", "Trace",
		"Infof", "Warnf", "Errorf", "Fatalf", "Debugf", "Tracef",
		"InfoContext", "WarnContext", "ErrorContext", "FatalContext", "DebugContext", "TraceContext",
		"InfofContext", "WarnfContext", "ErrorfContext", "FatalfContext", "DebugfContext", "TracefContext",
		// Backend methods
		"EmitInt", "EmitFloat", "EmitDuration",
		// Registration methods
		"Metric", "Log", "MetricWithProps", "LogWithProps",
	}

	for _, m := range methods {
		if name == m {
			return true
		}
	}
	return false
}

// getEventNameArgIndex returns the argument index for the event name parameter
func getEventNameArgIndex(methodName string) int {
	// For most methods, event name is the first parameter after context (if any)
	// Context methods have ctx as first param, so event is at index 1
	contextMethods := map[string]bool{
		"Count": true, "Gauge": true, "Histogram": true, "Meter": true, "Set": true, "Event": true,
		"InfoContext": true, "WarnContext": true, "ErrorContext": true, "FatalContext": true,
		"DebugContext": true, "TraceContext": true,
		"InfofContext": true, "WarnfContext": true, "ErrorfContext": true, "FatalfContext": true,
		"DebugfContext": true, "TracefContext": true,
		"EmitInt": true, "EmitFloat": true, "EmitDuration": true,
	}

	if contextMethods[methodName] {
		return 1 // event is second arg (after context)
	}

	// For Metric, Log, and non-context methods, event is first arg
	return 0
}

// getPropsArgIndex returns the argument index for the props parameter
func getPropsArgIndex(methodName string) int {
	// Context methods: Count(ctx, event, props, value)
	// Props is at index 2 for context methods
	contextMethods := map[string]bool{
		"Count": true, "Gauge": true, "Histogram": true, "Meter": true, "Set": true, "Event": true,
		"InfoContext": true, "WarnContext": true, "ErrorContext": true, "FatalContext": true,
		"DebugContext": true, "TraceContext": true,
		"InfofContext": true, "WarnfContext": true, "ErrorfContext": true, "FatalfContext": true,
		"DebugfContext": true, "TracefContext": true,
		"EmitInt": true, "EmitFloat": true, "EmitDuration": true,
	}

	if contextMethods[methodName] {
		return 2 // props is third arg (after context and event)
	}

	// Non-context methods: Info(event, props, msg)
	// Props is at index 1
	return 1
}

// writeOutputFile writes the generated Go code to the output file
func writeOutputFile(outputPath, pkgName, varName string, callsites map[string]CallSite) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer f.Close()

	w := &codeWriter{file: f}

	w.writeLine("// Code generated by go-emitter. DO NOT EDIT.")
	w.writeLine("")
	w.writeLine("package %s", pkgName)
	w.writeLine("")
	w.writeLine("import (")
	w.writeLine("\t\"github.com/pseudofunctor-ai/go-emitter/emitter/types\"")
	w.writeLine(")")
	w.writeLine("")
	w.writeLine("var %s = map[string]types.CallSiteDetails{", varName)

	// Sort event names by filename, then by line number for deterministic, readable output
	eventNames := make([]string, 0, len(callsites))
	for eventName := range callsites {
		eventNames = append(eventNames, eventName)
	}

	// Sort by filename first, then by line number
	sortCallsites(eventNames, callsites)

	for _, eventName := range eventNames {
		cs := callsites[eventName]
		w.writeLine("\t%q: {", eventName)
		w.writeLine("\t\tFilename:     %q,", cs.Filename)
		w.writeLine("\t\tLineNo:       %d,", cs.LineNo)
		w.writeLine("\t\tFuncName:     %q,", cs.FuncName)
		w.writeLine("\t\tPackage:      %q,", cs.Package)
		if len(cs.PropertyKeys) > 0 {
			w.writeLine("\t\tPropertyKeys: []string{%s},", formatStringSlice(cs.PropertyKeys))
		} else {
			w.writeLine("\t\tPropertyKeys: nil,")
		}
		if cs.MetricType != "" {
			w.writeLine("\t\tMetricType:   %q,", cs.MetricType)
		} else {
			w.writeLine("\t\tMetricType:   \"\",")
		}
		w.writeLine("\t},")
	}

	w.writeLine("}")
	w.writeLine("")

	return w.err
}

type codeWriter struct {
	file *os.File
	err  error
}

func (w *codeWriter) writeLine(format string, args ...interface{}) {
	if w.err != nil {
		return
	}

	line := fmt.Sprintf(format, args...)
	_, w.err = fmt.Fprintf(w.file, "%s\n", line)
}

// sortCallsites sorts event names by filename first, then by line number
func sortCallsites(eventNames []string, callsites map[string]CallSite) {
	sort.Slice(eventNames, func(i, j int) bool {
		if callsites[eventNames[i]].Filename == callsites[eventNames[j]].Filename {
			return callsites[eventNames[i]].LineNo < callsites[eventNames[j]].LineNo
		}
		return callsites[eventNames[i]].Filename < callsites[eventNames[j]].Filename
	})
}

// sortStrings is a simple insertion sort for string slices
func sortStrings(strs []string) {
	for i := 1; i < len(strs); i++ {
		key := strs[i]
		j := i - 1
		for j >= 0 && strs[j] > key {
			strs[j+1] = strs[j]
			j--
		}
		strs[j+1] = key
	}
}

// formatStringSlice formats a string slice for code generation
func formatStringSlice(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	quoted := make([]string, len(strs))
	for i, s := range strs {
		quoted[i] = fmt.Sprintf("%q", s)
	}
	return strings.Join(quoted, ", ")
}
