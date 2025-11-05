package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
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
	EventName string
	Filename  string
	LineNo    int
	FuncName  string
	Package   string
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

	// Check if this is an emitter method call
	eventName, callsite, isDecorator := e.extractFromCall(callExpr)
	if eventName == "" {
		return e
	}

	// Check for duplicate event names
	if existing, found := e.callsites[eventName]; found {
		// Allow decorators to override existing call sites
		if !isDecorator {
			e.err = fmt.Errorf("duplicate event name %q: already defined at %s:%d, found again at %s:%d",
				eventName, existing.Filename, existing.LineNo, callsite.Filename, callsite.LineNo)
			return nil
		}
		// Decorator overrides the existing call site
	}

	e.callsites[eventName] = callsite
	return e
}

// extractFromCall extracts event name and call site from a function call
// Returns: eventName, callsite, isDecorator
func (e *callSiteExtractor) extractFromCall(call *ast.CallExpr) (string, CallSite, bool) {
	selExpr, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", CallSite{}, false
	}

	methodName := selExpr.Sel.Name

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

	return eventName, e.makeCallSite(call, eventName)
}

// extractCallsiteDecorator extracts event name from *FnCallsite decorator calls
func (e *callSiteExtractor) extractCallsiteDecorator(call *ast.CallExpr, decoratorName string) (string, CallSite) {
	if len(call.Args) != 1 {
		return "", CallSite{}
	}

	// The argument should be an identifier referring to a MetricEmitterFn or LogEmitterFn
	ident, ok := call.Args[0].(*ast.Ident)
	if !ok {
		return "", CallSite{}
	}

	// Look up where this identifier is defined
	obj := e.pkg.TypesInfo.Uses[ident]
	if obj == nil {
		return "", CallSite{}
	}

	// Find the assignment/declaration
	eventName := e.findEventNameFromDefinition(obj.Pos())
	if eventName == "" {
		return "", CallSite{}
	}

	// Return the call site at the decorator location (not the definition location)
	return eventName, e.makeCallSite(call, eventName)
}

// findEventNameFromDefinition finds the event name from where a variable is defined
func (e *callSiteExtractor) findEventNameFromDefinition(pos token.Pos) string {
	// Find the file containing this position
	var targetFile *ast.File
	for _, file := range e.pkg.Syntax {
		if e.pkg.Fset.Position(file.Pos()).Filename == e.pkg.Fset.Position(pos).Filename {
			targetFile = file
			break
		}
	}

	if targetFile == nil {
		return ""
	}

	// Find the assignment or declaration at this position
	var eventName string
	ast.Inspect(targetFile, func(n ast.Node) bool {
		if eventName != "" {
			return false
		}

		switch node := n.(type) {
		case *ast.AssignStmt:
			if node.Pos() <= pos && pos <= node.End() {
				eventName = e.extractEventNameFromAssignment(node)
			}
		case *ast.ValueSpec:
			if node.Pos() <= pos && pos <= node.End() {
				eventName = e.extractEventNameFromValueSpec(node)
			}
		}

		return eventName == ""
	})

	return eventName
}

// extractEventNameFromAssignment extracts event name from an assignment like `foo := emitter.Metric("event_name", COUNT)`
func (e *callSiteExtractor) extractEventNameFromAssignment(assign *ast.AssignStmt) string {
	if len(assign.Rhs) != 1 {
		return ""
	}

	call, ok := assign.Rhs[0].(*ast.CallExpr)
	if !ok {
		return ""
	}

	selExpr, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return ""
	}

	methodName := selExpr.Sel.Name
	if methodName == "Metric" || methodName == "Log" {
		return e.extractEventNameArg(call, methodName)
	}

	return ""
}

// extractEventNameFromValueSpec extracts event name from a var declaration
func (e *callSiteExtractor) extractEventNameFromValueSpec(spec *ast.ValueSpec) string {
	if len(spec.Values) != 1 {
		return ""
	}

	call, ok := spec.Values[0].(*ast.CallExpr)
	if !ok {
		return ""
	}

	selExpr, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return ""
	}

	methodName := selExpr.Sel.Name
	if methodName == "Metric" || methodName == "Log" {
		return e.extractEventNameArg(call, methodName)
	}

	return ""
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
		"Metric", "Log",
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

// writeOutputFile writes the generated Go code to the output file
func writeOutputFile(outputPath, pkgName, varName string, callsites map[string]CallSite) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer f.Close()

	w := &codeWriter{file: f}

	w.writeLine("// Code generated by emitter-gen. DO NOT EDIT.")
	w.writeLine("")
	w.writeLine("package %s", pkgName)
	w.writeLine("")
	w.writeLine("import (")
	w.writeLine("\t\"github.com/pseudofunctor-ai/go-emitter/emitter/types\"")
	w.writeLine(")")
	w.writeLine("")
	w.writeLine("var %s = map[string]types.CallSiteDetails{", varName)

	// Sort event names for deterministic output
	eventNames := make([]string, 0, len(callsites))
	for eventName := range callsites {
		eventNames = append(eventNames, eventName)
	}
	sortStrings(eventNames)

	for _, eventName := range eventNames {
		cs := callsites[eventName]
		w.writeLine("\t%q: {", eventName)
		w.writeLine("\t\tFilename: %q,", cs.Filename)
		w.writeLine("\t\tLineNo:   %d,", cs.LineNo)
		w.writeLine("\t\tFuncName: %q,", cs.FuncName)
		w.writeLine("\t\tPackage:  %q,", cs.Package)
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
