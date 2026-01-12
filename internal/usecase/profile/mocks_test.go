// Package profile_test provides mocks for profile usecase tests.
package profile_test

import (
	"context"
	"io"
	"reflect"
	"time"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo/storage"

	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

// MockProfileRepo is a mock of ProfileRepo interface.
type MockProfileRepo struct {
	ctrl     *gomock.Controller
	recorder *MockProfileRepoMockRecorder
}

// MockProfileRepoMockRecorder is the mock recorder for MockProfileRepo.
type MockProfileRepoMockRecorder struct {
	mock *MockProfileRepo
}

// NewMockProfileRepo creates a new mock instance.
func NewMockProfileRepo(ctrl *gomock.Controller) *MockProfileRepo {
	mock := &MockProfileRepo{ctrl: ctrl}
	mock.recorder = &MockProfileRepoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProfileRepo) EXPECT() *MockProfileRepoMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockProfileRepo) Create(ctx context.Context, profile *entity.Profile) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, profile)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create.
func (mr *MockProfileRepoMockRecorder) Create(ctx, profile any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockProfileRepo)(nil).Create), ctx, profile)
}

// GetByUserID mocks base method.
func (m *MockProfileRepo) GetByUserID(ctx context.Context, userID uint) (*entity.Profile, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByUserID", ctx, userID)
	ret0, _ := ret[0].(*entity.Profile)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByUserID indicates an expected call of GetByUserID.
func (mr *MockProfileRepoMockRecorder) GetByUserID(ctx, userID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByUserID", reflect.TypeOf((*MockProfileRepo)(nil).GetByUserID), ctx, userID)
}

// Update mocks base method.
func (m *MockProfileRepo) Update(ctx context.Context, profile *entity.Profile) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", ctx, profile)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *MockProfileRepoMockRecorder) Update(ctx, profile any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockProfileRepo)(nil).Update), ctx, profile)
}

// Upsert mocks base method.
func (m *MockProfileRepo) Upsert(ctx context.Context, profile *entity.Profile) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Upsert", ctx, profile)
	ret0, _ := ret[0].(error)
	return ret0
}

// Upsert indicates an expected call of Upsert.
func (mr *MockProfileRepoMockRecorder) Upsert(ctx, profile any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Upsert", reflect.TypeOf((*MockProfileRepo)(nil).Upsert), ctx, profile)
}

// MockMediaRepo is a mock of MediaRepo interface.
type MockMediaRepo struct {
	ctrl     *gomock.Controller
	recorder *MockMediaRepoMockRecorder
}

// MockMediaRepoMockRecorder is the mock recorder for MockMediaRepo.
type MockMediaRepoMockRecorder struct {
	mock *MockMediaRepo
}

// NewMockMediaRepo creates a new mock instance.
func NewMockMediaRepo(ctrl *gomock.Controller) *MockMediaRepo {
	mock := &MockMediaRepo{ctrl: ctrl}
	mock.recorder = &MockMediaRepoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMediaRepo) EXPECT() *MockMediaRepoMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockMediaRepo) Create(ctx context.Context, media *entity.Media) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, media)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create.
func (mr *MockMediaRepoMockRecorder) Create(ctx, media any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockMediaRepo)(nil).Create), ctx, media)
}

// GetByID mocks base method.
func (m *MockMediaRepo) GetByID(ctx context.Context, id uint) (*entity.Media, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByID", ctx, id)
	ret0, _ := ret[0].(*entity.Media)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByID indicates an expected call of GetByID.
func (mr *MockMediaRepoMockRecorder) GetByID(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockMediaRepo)(nil).GetByID), ctx, id)
}

// GetByAttachable mocks base method.
func (m *MockMediaRepo) GetByAttachable(ctx context.Context, attachableType string, attachableID uint, collection string) ([]*entity.Media, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByAttachable", ctx, attachableType, attachableID, collection)
	ret0, _ := ret[0].([]*entity.Media)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByAttachable indicates an expected call of GetByAttachable.
func (mr *MockMediaRepoMockRecorder) GetByAttachable(ctx, attachableType, attachableID, collection any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByAttachable", reflect.TypeOf((*MockMediaRepo)(nil).GetByAttachable), ctx, attachableType, attachableID, collection)
}

// Update mocks base method.
func (m *MockMediaRepo) Update(ctx context.Context, media *entity.Media) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", ctx, media)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *MockMediaRepoMockRecorder) Update(ctx, media any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockMediaRepo)(nil).Update), ctx, media)
}

// Delete mocks base method.
func (m *MockMediaRepo) Delete(ctx context.Context, id uint) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockMediaRepoMockRecorder) Delete(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockMediaRepo)(nil).Delete), ctx, id)
}

// DeleteByAttachable mocks base method.
func (m *MockMediaRepo) DeleteByAttachable(ctx context.Context, attachableType string, attachableID uint) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteByAttachable", ctx, attachableType, attachableID)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteByAttachable indicates an expected call of DeleteByAttachable.
func (mr *MockMediaRepoMockRecorder) DeleteByAttachable(ctx, attachableType, attachableID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteByAttachable", reflect.TypeOf((*MockMediaRepo)(nil).DeleteByAttachable), ctx, attachableType, attachableID)
}

// MockStorageProvider is a mock of storage.Provider interface.
type MockStorageProvider struct {
	ctrl     *gomock.Controller
	recorder *MockStorageProviderMockRecorder
}

// MockStorageProviderMockRecorder is the mock recorder for MockStorageProvider.
type MockStorageProviderMockRecorder struct {
	mock *MockStorageProvider
}

// NewMockStorageProvider creates a new mock instance.
func NewMockStorageProvider(ctrl *gomock.Controller) *MockStorageProvider {
	mock := &MockStorageProvider{ctrl: ctrl}
	mock.recorder = &MockStorageProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStorageProvider) EXPECT() *MockStorageProviderMockRecorder {
	return m.recorder
}

// Put mocks base method.
func (m *MockStorageProvider) Put(ctx context.Context, path string, reader io.Reader, size int64, mimeType string) (*storage.FileInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Put", ctx, path, reader, size, mimeType)
	ret0, _ := ret[0].(*storage.FileInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Put indicates an expected call of Put.
func (mr *MockStorageProviderMockRecorder) Put(ctx, path, reader, size, mimeType any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Put", reflect.TypeOf((*MockStorageProvider)(nil).Put), ctx, path, reader, size, mimeType)
}

// Get mocks base method.
func (m *MockStorageProvider) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, path)
	ret0, _ := ret[0].(io.ReadCloser)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockStorageProviderMockRecorder) Get(ctx, path any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockStorageProvider)(nil).Get), ctx, path)
}

// Delete mocks base method.
func (m *MockStorageProvider) Delete(ctx context.Context, path string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, path)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockStorageProviderMockRecorder) Delete(ctx, path any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockStorageProvider)(nil).Delete), ctx, path)
}

// Exists mocks base method.
func (m *MockStorageProvider) Exists(ctx context.Context, path string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Exists", ctx, path)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Exists indicates an expected call of Exists.
func (mr *MockStorageProviderMockRecorder) Exists(ctx, path any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exists", reflect.TypeOf((*MockStorageProvider)(nil).Exists), ctx, path)
}

// URL mocks base method.
func (m *MockStorageProvider) URL(ctx context.Context, path string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "URL", ctx, path)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// URL indicates an expected call of URL.
func (mr *MockStorageProviderMockRecorder) URL(ctx, path any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "URL", reflect.TypeOf((*MockStorageProvider)(nil).URL), ctx, path)
}

// TemporaryURL mocks base method.
func (m *MockStorageProvider) TemporaryURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TemporaryURL", ctx, path, expiry)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TemporaryURL indicates an expected call of TemporaryURL.
func (mr *MockStorageProviderMockRecorder) TemporaryURL(ctx, path, expiry any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TemporaryURL", reflect.TypeOf((*MockStorageProvider)(nil).TemporaryURL), ctx, path, expiry)
}

// PresignedUploadURL mocks base method.
func (m *MockStorageProvider) PresignedUploadURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PresignedUploadURL", ctx, path, expiry)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PresignedUploadURL indicates an expected call of PresignedUploadURL.
func (mr *MockStorageProviderMockRecorder) PresignedUploadURL(ctx, path, expiry any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PresignedUploadURL", reflect.TypeOf((*MockStorageProvider)(nil).PresignedUploadURL), ctx, path, expiry)
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
