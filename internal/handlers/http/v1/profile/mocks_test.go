// Package profile_test provides mocks for profile handler tests.
package profile_test

import (
	"context"
	"reflect"

	"go-boilerplate/internal/dto/profile"

	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

// MockProfile is a mock of Profile interface.
type MockProfile struct {
	ctrl     *gomock.Controller
	recorder *MockProfileMockRecorder
}

// MockProfileMockRecorder is the mock recorder for MockProfile.
type MockProfileMockRecorder struct {
	mock *MockProfile
}

// NewMockProfile creates a new mock instance.
func NewMockProfile(ctrl *gomock.Controller) *MockProfile {
	mock := &MockProfile{ctrl: ctrl}
	mock.recorder = &MockProfileMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProfile) EXPECT() *MockProfileMockRecorder {
	return m.recorder
}

// GetProfile mocks base method.
func (m *MockProfile) GetProfile(ctx context.Context, userID uint) (*profiledto.ProfileResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProfile", ctx, userID)
	ret0, _ := ret[0].(*profiledto.ProfileResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProfile indicates an expected call of GetProfile.
func (mr *MockProfileMockRecorder) GetProfile(ctx, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProfile", reflect.TypeOf((*MockProfile)(nil).GetProfile), ctx, userID)
}

// UpdateProfile mocks base method.
func (m *MockProfile) UpdateProfile(ctx context.Context, userID uint, req profiledto.UpdateProfileRequest) (*profiledto.ProfileResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateProfile", ctx, userID, req)
	ret0, _ := ret[0].(*profiledto.ProfileResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateProfile indicates an expected call of UpdateProfile.
func (mr *MockProfileMockRecorder) UpdateProfile(ctx, userID, req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateProfile", reflect.TypeOf((*MockProfile)(nil).UpdateProfile), ctx, userID, req)
}

// MockLogger is a mock of logger.Interface.
type MockLogger struct{}

// NewMockLogger creates a new mock logger.
func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

func (m *MockLogger) Debug(message interface{}, args ...interface{}) {}
func (m *MockLogger) Info(message string, args ...interface{})       {}
func (m *MockLogger) Warn(message string, args ...interface{})       {}
func (m *MockLogger) Error(message interface{}, args ...interface{}) {}
func (m *MockLogger) Fatal(message interface{}, args ...interface{}) {}
func (m *MockLogger) GetZapLogger() *zap.Logger                      { return nil }
