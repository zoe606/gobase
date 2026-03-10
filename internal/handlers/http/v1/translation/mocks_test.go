package translation_test

import (
	"context"
	"reflect"

	translationdto "go-boilerplate/internal/dto/translation"

	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

// MockTranslation is a mock of Translation interface.
type MockTranslation struct {
	ctrl     *gomock.Controller
	recorder *MockTranslationMockRecorder
}

// MockTranslationMockRecorder is the mock recorder for MockTranslation.
type MockTranslationMockRecorder struct {
	mock *MockTranslation
}

// NewMockTranslation creates a new mock instance.
func NewMockTranslation(ctrl *gomock.Controller) *MockTranslation {
	mock := &MockTranslation{ctrl: ctrl}
	mock.recorder = &MockTranslationMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTranslation) EXPECT() *MockTranslationMockRecorder {
	return m.recorder
}

// Translate mocks base method.
func (m *MockTranslation) Translate(ctx context.Context, req translationdto.TranslateRequest) (*translationdto.TranslationResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Translate", ctx, req)
	ret0, _ := ret[0].(*translationdto.TranslationResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Translate indicates an expected call of Translate.
func (mr *MockTranslationMockRecorder) Translate(ctx, req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Translate", reflect.TypeOf((*MockTranslation)(nil).Translate), ctx, req)
}

// History mocks base method.
func (m *MockTranslation) History(ctx context.Context, req translationdto.HistoryRequest) (*translationdto.HistoryResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "History", ctx, req)
	ret0, _ := ret[0].(*translationdto.HistoryResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// History indicates an expected call of History.
func (mr *MockTranslationMockRecorder) History(ctx, req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "History", reflect.TypeOf((*MockTranslation)(nil).History), ctx, req)
}

// MockLogger is a mock of logger.Interface.
type MockLogger struct{}

func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

func (m *MockLogger) Debug(message interface{}, args ...interface{}) {}
func (m *MockLogger) Info(message string, args ...interface{})       {}
func (m *MockLogger) Warn(message string, args ...interface{})       {}
func (m *MockLogger) Error(message interface{}, args ...interface{}) {}
func (m *MockLogger) Fatal(message interface{}, args ...interface{}) {}
func (m *MockLogger) GetZapLogger() *zap.Logger                      { return nil }
