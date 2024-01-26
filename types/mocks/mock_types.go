// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/pseudofunctor-ai/go-emitter/types (interfaces: SimpleLogger,SimpleLoggerFmt,ContextLogger,ContextLoggerFmt,MetricsEmitter,EmitterBackend)
//
// Generated by this command:
//
//	mockgen -destination=mocks/mock_types.go -package mocks github.com/pseudofunctor-ai/go-emitter/types SimpleLogger,SimpleLoggerFmt,ContextLogger,ContextLoggerFmt,MetricsEmitter,EmitterBackend
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"
	time "time"

	types "github.com/pseudofunctor-ai/go-emitter/types"
	gomock "go.uber.org/mock/gomock"
)

// MockSimpleLogger is a mock of SimpleLogger interface.
type MockSimpleLogger struct {
	ctrl     *gomock.Controller
	recorder *MockSimpleLoggerMockRecorder
}

// MockSimpleLoggerMockRecorder is the mock recorder for MockSimpleLogger.
type MockSimpleLoggerMockRecorder struct {
	mock *MockSimpleLogger
}

// NewMockSimpleLogger creates a new mock instance.
func NewMockSimpleLogger(ctrl *gomock.Controller) *MockSimpleLogger {
	mock := &MockSimpleLogger{ctrl: ctrl}
	mock.recorder = &MockSimpleLoggerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSimpleLogger) EXPECT() *MockSimpleLoggerMockRecorder {
	return m.recorder
}

// Debug mocks base method.
func (m *MockSimpleLogger) Debug(arg0 string, arg1 map[string]any, arg2 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Debug", arg0, arg1, arg2)
}

// Debug indicates an expected call of Debug.
func (mr *MockSimpleLoggerMockRecorder) Debug(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Debug", reflect.TypeOf((*MockSimpleLogger)(nil).Debug), arg0, arg1, arg2)
}

// Error mocks base method.
func (m *MockSimpleLogger) Error(arg0 string, arg1 map[string]any, arg2 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Error", arg0, arg1, arg2)
}

// Error indicates an expected call of Error.
func (mr *MockSimpleLoggerMockRecorder) Error(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Error", reflect.TypeOf((*MockSimpleLogger)(nil).Error), arg0, arg1, arg2)
}

// Fatal mocks base method.
func (m *MockSimpleLogger) Fatal(arg0 string, arg1 map[string]any, arg2 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Fatal", arg0, arg1, arg2)
}

// Fatal indicates an expected call of Fatal.
func (mr *MockSimpleLoggerMockRecorder) Fatal(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Fatal", reflect.TypeOf((*MockSimpleLogger)(nil).Fatal), arg0, arg1, arg2)
}

// Info mocks base method.
func (m *MockSimpleLogger) Info(arg0 string, arg1 map[string]any, arg2 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Info", arg0, arg1, arg2)
}

// Info indicates an expected call of Info.
func (mr *MockSimpleLoggerMockRecorder) Info(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Info", reflect.TypeOf((*MockSimpleLogger)(nil).Info), arg0, arg1, arg2)
}

// Trace mocks base method.
func (m *MockSimpleLogger) Trace(arg0 string, arg1 map[string]any, arg2 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Trace", arg0, arg1, arg2)
}

// Trace indicates an expected call of Trace.
func (mr *MockSimpleLoggerMockRecorder) Trace(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Trace", reflect.TypeOf((*MockSimpleLogger)(nil).Trace), arg0, arg1, arg2)
}

// Warn mocks base method.
func (m *MockSimpleLogger) Warn(arg0 string, arg1 map[string]any, arg2 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Warn", arg0, arg1, arg2)
}

// Warn indicates an expected call of Warn.
func (mr *MockSimpleLoggerMockRecorder) Warn(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Warn", reflect.TypeOf((*MockSimpleLogger)(nil).Warn), arg0, arg1, arg2)
}

// MockSimpleLoggerFmt is a mock of SimpleLoggerFmt interface.
type MockSimpleLoggerFmt struct {
	ctrl     *gomock.Controller
	recorder *MockSimpleLoggerFmtMockRecorder
}

// MockSimpleLoggerFmtMockRecorder is the mock recorder for MockSimpleLoggerFmt.
type MockSimpleLoggerFmtMockRecorder struct {
	mock *MockSimpleLoggerFmt
}

// NewMockSimpleLoggerFmt creates a new mock instance.
func NewMockSimpleLoggerFmt(ctrl *gomock.Controller) *MockSimpleLoggerFmt {
	mock := &MockSimpleLoggerFmt{ctrl: ctrl}
	mock.recorder = &MockSimpleLoggerFmtMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSimpleLoggerFmt) EXPECT() *MockSimpleLoggerFmtMockRecorder {
	return m.recorder
}

// Debugf mocks base method.
func (m *MockSimpleLoggerFmt) Debugf(arg0 string, arg1 map[string]any, arg2 string, arg3 ...any) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Debugf", varargs...)
}

// Debugf indicates an expected call of Debugf.
func (mr *MockSimpleLoggerFmtMockRecorder) Debugf(arg0, arg1, arg2 any, arg3 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Debugf", reflect.TypeOf((*MockSimpleLoggerFmt)(nil).Debugf), varargs...)
}

// Errorf mocks base method.
func (m *MockSimpleLoggerFmt) Errorf(arg0 string, arg1 map[string]any, arg2 string, arg3 ...any) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Errorf", varargs...)
}

// Errorf indicates an expected call of Errorf.
func (mr *MockSimpleLoggerFmtMockRecorder) Errorf(arg0, arg1, arg2 any, arg3 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Errorf", reflect.TypeOf((*MockSimpleLoggerFmt)(nil).Errorf), varargs...)
}

// Fatalf mocks base method.
func (m *MockSimpleLoggerFmt) Fatalf(arg0 string, arg1 map[string]any, arg2 string, arg3 ...any) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Fatalf", varargs...)
}

// Fatalf indicates an expected call of Fatalf.
func (mr *MockSimpleLoggerFmtMockRecorder) Fatalf(arg0, arg1, arg2 any, arg3 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Fatalf", reflect.TypeOf((*MockSimpleLoggerFmt)(nil).Fatalf), varargs...)
}

// Infof mocks base method.
func (m *MockSimpleLoggerFmt) Infof(arg0 string, arg1 map[string]any, arg2 string, arg3 ...any) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Infof", varargs...)
}

// Infof indicates an expected call of Infof.
func (mr *MockSimpleLoggerFmtMockRecorder) Infof(arg0, arg1, arg2 any, arg3 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Infof", reflect.TypeOf((*MockSimpleLoggerFmt)(nil).Infof), varargs...)
}

// Tracef mocks base method.
func (m *MockSimpleLoggerFmt) Tracef(arg0 string, arg1 map[string]any, arg2 string, arg3 ...any) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Tracef", varargs...)
}

// Tracef indicates an expected call of Tracef.
func (mr *MockSimpleLoggerFmtMockRecorder) Tracef(arg0, arg1, arg2 any, arg3 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Tracef", reflect.TypeOf((*MockSimpleLoggerFmt)(nil).Tracef), varargs...)
}

// Warnf mocks base method.
func (m *MockSimpleLoggerFmt) Warnf(arg0 string, arg1 map[string]any, arg2 string, arg3 ...any) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Warnf", varargs...)
}

// Warnf indicates an expected call of Warnf.
func (mr *MockSimpleLoggerFmtMockRecorder) Warnf(arg0, arg1, arg2 any, arg3 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Warnf", reflect.TypeOf((*MockSimpleLoggerFmt)(nil).Warnf), varargs...)
}

// MockContextLogger is a mock of ContextLogger interface.
type MockContextLogger struct {
	ctrl     *gomock.Controller
	recorder *MockContextLoggerMockRecorder
}

// MockContextLoggerMockRecorder is the mock recorder for MockContextLogger.
type MockContextLoggerMockRecorder struct {
	mock *MockContextLogger
}

// NewMockContextLogger creates a new mock instance.
func NewMockContextLogger(ctrl *gomock.Controller) *MockContextLogger {
	mock := &MockContextLogger{ctrl: ctrl}
	mock.recorder = &MockContextLoggerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockContextLogger) EXPECT() *MockContextLoggerMockRecorder {
	return m.recorder
}

// DebugContext mocks base method.
func (m *MockContextLogger) DebugContext(arg0 context.Context, arg1 string, arg2 map[string]any, arg3 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "DebugContext", arg0, arg1, arg2, arg3)
}

// DebugContext indicates an expected call of DebugContext.
func (mr *MockContextLoggerMockRecorder) DebugContext(arg0, arg1, arg2, arg3 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DebugContext", reflect.TypeOf((*MockContextLogger)(nil).DebugContext), arg0, arg1, arg2, arg3)
}

// ErrorContext mocks base method.
func (m *MockContextLogger) ErrorContext(arg0 context.Context, arg1 string, arg2 map[string]any, arg3 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ErrorContext", arg0, arg1, arg2, arg3)
}

// ErrorContext indicates an expected call of ErrorContext.
func (mr *MockContextLoggerMockRecorder) ErrorContext(arg0, arg1, arg2, arg3 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ErrorContext", reflect.TypeOf((*MockContextLogger)(nil).ErrorContext), arg0, arg1, arg2, arg3)
}

// FatalContext mocks base method.
func (m *MockContextLogger) FatalContext(arg0 context.Context, arg1 string, arg2 map[string]any, arg3 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "FatalContext", arg0, arg1, arg2, arg3)
}

// FatalContext indicates an expected call of FatalContext.
func (mr *MockContextLoggerMockRecorder) FatalContext(arg0, arg1, arg2, arg3 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FatalContext", reflect.TypeOf((*MockContextLogger)(nil).FatalContext), arg0, arg1, arg2, arg3)
}

// InfoContext mocks base method.
func (m *MockContextLogger) InfoContext(arg0 context.Context, arg1 string, arg2 map[string]any, arg3 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "InfoContext", arg0, arg1, arg2, arg3)
}

// InfoContext indicates an expected call of InfoContext.
func (mr *MockContextLoggerMockRecorder) InfoContext(arg0, arg1, arg2, arg3 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InfoContext", reflect.TypeOf((*MockContextLogger)(nil).InfoContext), arg0, arg1, arg2, arg3)
}

// TraceContext mocks base method.
func (m *MockContextLogger) TraceContext(arg0 context.Context, arg1 string, arg2 map[string]any, arg3 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "TraceContext", arg0, arg1, arg2, arg3)
}

// TraceContext indicates an expected call of TraceContext.
func (mr *MockContextLoggerMockRecorder) TraceContext(arg0, arg1, arg2, arg3 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TraceContext", reflect.TypeOf((*MockContextLogger)(nil).TraceContext), arg0, arg1, arg2, arg3)
}

// WarnContext mocks base method.
func (m *MockContextLogger) WarnContext(arg0 context.Context, arg1 string, arg2 map[string]any, arg3 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "WarnContext", arg0, arg1, arg2, arg3)
}

// WarnContext indicates an expected call of WarnContext.
func (mr *MockContextLoggerMockRecorder) WarnContext(arg0, arg1, arg2, arg3 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WarnContext", reflect.TypeOf((*MockContextLogger)(nil).WarnContext), arg0, arg1, arg2, arg3)
}

// MockContextLoggerFmt is a mock of ContextLoggerFmt interface.
type MockContextLoggerFmt struct {
	ctrl     *gomock.Controller
	recorder *MockContextLoggerFmtMockRecorder
}

// MockContextLoggerFmtMockRecorder is the mock recorder for MockContextLoggerFmt.
type MockContextLoggerFmtMockRecorder struct {
	mock *MockContextLoggerFmt
}

// NewMockContextLoggerFmt creates a new mock instance.
func NewMockContextLoggerFmt(ctrl *gomock.Controller) *MockContextLoggerFmt {
	mock := &MockContextLoggerFmt{ctrl: ctrl}
	mock.recorder = &MockContextLoggerFmtMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockContextLoggerFmt) EXPECT() *MockContextLoggerFmtMockRecorder {
	return m.recorder
}

// DebugfContext mocks base method.
func (m *MockContextLoggerFmt) DebugfContext(arg0 context.Context, arg1 string, arg2 map[string]any, arg3 string, arg4 ...any) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1, arg2, arg3}
	for _, a := range arg4 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "DebugfContext", varargs...)
}

// DebugfContext indicates an expected call of DebugfContext.
func (mr *MockContextLoggerFmtMockRecorder) DebugfContext(arg0, arg1, arg2, arg3 any, arg4 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1, arg2, arg3}, arg4...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DebugfContext", reflect.TypeOf((*MockContextLoggerFmt)(nil).DebugfContext), varargs...)
}

// ErrorfContext mocks base method.
func (m *MockContextLoggerFmt) ErrorfContext(arg0 context.Context, arg1 string, arg2 map[string]any, arg3 string, arg4 ...any) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1, arg2, arg3}
	for _, a := range arg4 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "ErrorfContext", varargs...)
}

// ErrorfContext indicates an expected call of ErrorfContext.
func (mr *MockContextLoggerFmtMockRecorder) ErrorfContext(arg0, arg1, arg2, arg3 any, arg4 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1, arg2, arg3}, arg4...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ErrorfContext", reflect.TypeOf((*MockContextLoggerFmt)(nil).ErrorfContext), varargs...)
}

// FatalfContext mocks base method.
func (m *MockContextLoggerFmt) FatalfContext(arg0 context.Context, arg1 string, arg2 map[string]any, arg3 string, arg4 ...any) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1, arg2, arg3}
	for _, a := range arg4 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "FatalfContext", varargs...)
}

// FatalfContext indicates an expected call of FatalfContext.
func (mr *MockContextLoggerFmtMockRecorder) FatalfContext(arg0, arg1, arg2, arg3 any, arg4 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1, arg2, arg3}, arg4...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FatalfContext", reflect.TypeOf((*MockContextLoggerFmt)(nil).FatalfContext), varargs...)
}

// InfofContext mocks base method.
func (m *MockContextLoggerFmt) InfofContext(arg0 context.Context, arg1 string, arg2 map[string]any, arg3 string, arg4 ...any) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1, arg2, arg3}
	for _, a := range arg4 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "InfofContext", varargs...)
}

// InfofContext indicates an expected call of InfofContext.
func (mr *MockContextLoggerFmtMockRecorder) InfofContext(arg0, arg1, arg2, arg3 any, arg4 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1, arg2, arg3}, arg4...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InfofContext", reflect.TypeOf((*MockContextLoggerFmt)(nil).InfofContext), varargs...)
}

// TracefContext mocks base method.
func (m *MockContextLoggerFmt) TracefContext(arg0 context.Context, arg1 string, arg2 map[string]any, arg3 string, arg4 ...any) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1, arg2, arg3}
	for _, a := range arg4 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "TracefContext", varargs...)
}

// TracefContext indicates an expected call of TracefContext.
func (mr *MockContextLoggerFmtMockRecorder) TracefContext(arg0, arg1, arg2, arg3 any, arg4 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1, arg2, arg3}, arg4...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TracefContext", reflect.TypeOf((*MockContextLoggerFmt)(nil).TracefContext), varargs...)
}

// WarnfContext mocks base method.
func (m *MockContextLoggerFmt) WarnfContext(arg0 context.Context, arg1 string, arg2 map[string]any, arg3 string, arg4 ...any) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1, arg2, arg3}
	for _, a := range arg4 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "WarnfContext", varargs...)
}

// WarnfContext indicates an expected call of WarnfContext.
func (mr *MockContextLoggerFmtMockRecorder) WarnfContext(arg0, arg1, arg2, arg3 any, arg4 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1, arg2, arg3}, arg4...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WarnfContext", reflect.TypeOf((*MockContextLoggerFmt)(nil).WarnfContext), varargs...)
}

// MockMetricsEmitter is a mock of MetricsEmitter interface.
type MockMetricsEmitter struct {
	ctrl     *gomock.Controller
	recorder *MockMetricsEmitterMockRecorder
}

// MockMetricsEmitterMockRecorder is the mock recorder for MockMetricsEmitter.
type MockMetricsEmitterMockRecorder struct {
	mock *MockMetricsEmitter
}

// NewMockMetricsEmitter creates a new mock instance.
func NewMockMetricsEmitter(ctrl *gomock.Controller) *MockMetricsEmitter {
	mock := &MockMetricsEmitter{ctrl: ctrl}
	mock.recorder = &MockMetricsEmitterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMetricsEmitter) EXPECT() *MockMetricsEmitterMockRecorder {
	return m.recorder
}

// Count mocks base method.
func (m *MockMetricsEmitter) Count(arg0 context.Context, arg1 string, arg2 map[string]any, arg3 float64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Count", arg0, arg1, arg2, arg3)
}

// Count indicates an expected call of Count.
func (mr *MockMetricsEmitterMockRecorder) Count(arg0, arg1, arg2, arg3 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Count", reflect.TypeOf((*MockMetricsEmitter)(nil).Count), arg0, arg1, arg2, arg3)
}

// Gauge mocks base method.
func (m *MockMetricsEmitter) Gauge(arg0 context.Context, arg1 string, arg2 map[string]any, arg3 float64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Gauge", arg0, arg1, arg2, arg3)
}

// Gauge indicates an expected call of Gauge.
func (mr *MockMetricsEmitterMockRecorder) Gauge(arg0, arg1, arg2, arg3 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Gauge", reflect.TypeOf((*MockMetricsEmitter)(nil).Gauge), arg0, arg1, arg2, arg3)
}

// Histogram mocks base method.
func (m *MockMetricsEmitter) Histogram(arg0 context.Context, arg1 string, arg2 map[string]any) types.Histogram {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Histogram", arg0, arg1, arg2)
	ret0, _ := ret[0].(types.Histogram)
	return ret0
}

// Histogram indicates an expected call of Histogram.
func (mr *MockMetricsEmitterMockRecorder) Histogram(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Histogram", reflect.TypeOf((*MockMetricsEmitter)(nil).Histogram), arg0, arg1, arg2)
}

// MockEmitterBackend is a mock of EmitterBackend interface.
type MockEmitterBackend struct {
	ctrl     *gomock.Controller
	recorder *MockEmitterBackendMockRecorder
}

// MockEmitterBackendMockRecorder is the mock recorder for MockEmitterBackend.
type MockEmitterBackendMockRecorder struct {
	mock *MockEmitterBackend
}

// NewMockEmitterBackend creates a new mock instance.
func NewMockEmitterBackend(ctrl *gomock.Controller) *MockEmitterBackend {
	mock := &MockEmitterBackend{ctrl: ctrl}
	mock.recorder = &MockEmitterBackendMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockEmitterBackend) EXPECT() *MockEmitterBackendMockRecorder {
	return m.recorder
}

// EmitDuration mocks base method.
func (m *MockEmitterBackend) EmitDuration(arg0 context.Context, arg1 string, arg2 map[string]any, arg3 time.Duration, arg4 types.MetricType) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "EmitDuration", arg0, arg1, arg2, arg3, arg4)
}

// EmitDuration indicates an expected call of EmitDuration.
func (mr *MockEmitterBackendMockRecorder) EmitDuration(arg0, arg1, arg2, arg3, arg4 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EmitDuration", reflect.TypeOf((*MockEmitterBackend)(nil).EmitDuration), arg0, arg1, arg2, arg3, arg4)
}

// EmitFloat mocks base method.
func (m *MockEmitterBackend) EmitFloat(arg0 context.Context, arg1 string, arg2 map[string]any, arg3 float64, arg4 types.MetricType) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "EmitFloat", arg0, arg1, arg2, arg3, arg4)
}

// EmitFloat indicates an expected call of EmitFloat.
func (mr *MockEmitterBackendMockRecorder) EmitFloat(arg0, arg1, arg2, arg3, arg4 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EmitFloat", reflect.TypeOf((*MockEmitterBackend)(nil).EmitFloat), arg0, arg1, arg2, arg3, arg4)
}

// EmitInt mocks base method.
func (m *MockEmitterBackend) EmitInt(arg0 context.Context, arg1 string, arg2 map[string]any, arg3 int64, arg4 types.MetricType) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "EmitInt", arg0, arg1, arg2, arg3, arg4)
}

// EmitInt indicates an expected call of EmitInt.
func (mr *MockEmitterBackendMockRecorder) EmitInt(arg0, arg1, arg2, arg3, arg4 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EmitInt", reflect.TypeOf((*MockEmitterBackend)(nil).EmitInt), arg0, arg1, arg2, arg3, arg4)
}
