// Code generated by MockGen. DO NOT EDIT.
// Source: domain/task_permission.go
//
// Generated by this command:
//
//	mockgen -source=domain/task_permission.go -destination=tests/mock/mock_task_permission.go -package=mock
//

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	domain "github.com/keitatwr/task-management-app/domain"
	gomock "go.uber.org/mock/gomock"
)

// MockTaskPermissionRepository is a mock of TaskPermissionRepository interface.
type MockTaskPermissionRepository struct {
	ctrl     *gomock.Controller
	recorder *MockTaskPermissionRepositoryMockRecorder
	isgomock struct{}
}

// MockTaskPermissionRepositoryMockRecorder is the mock recorder for MockTaskPermissionRepository.
type MockTaskPermissionRepositoryMockRecorder struct {
	mock *MockTaskPermissionRepository
}

// NewMockTaskPermissionRepository creates a new mock instance.
func NewMockTaskPermissionRepository(ctrl *gomock.Controller) *MockTaskPermissionRepository {
	mock := &MockTaskPermissionRepository{ctrl: ctrl}
	mock.recorder = &MockTaskPermissionRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTaskPermissionRepository) EXPECT() *MockTaskPermissionRepositoryMockRecorder {
	return m.recorder
}

// FetchPermissionByTaskID mocks base method.
func (m *MockTaskPermissionRepository) FetchPermissionByTaskID(ctx context.Context, taskID, userID int) (*domain.TaskPermission, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchPermissionByTaskID", ctx, taskID, userID)
	ret0, _ := ret[0].(*domain.TaskPermission)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchPermissionByTaskID indicates an expected call of FetchPermissionByTaskID.
func (mr *MockTaskPermissionRepositoryMockRecorder) FetchPermissionByTaskID(ctx, taskID, userID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchPermissionByTaskID", reflect.TypeOf((*MockTaskPermissionRepository)(nil).FetchPermissionByTaskID), ctx, taskID, userID)
}

// FetchTaskIDByUserID mocks base method.
func (m *MockTaskPermissionRepository) FetchTaskIDByUserID(ctx context.Context, id int, canEdit, canRead bool) ([]int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchTaskIDByUserID", ctx, id, canEdit, canRead)
	ret0, _ := ret[0].([]int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchTaskIDByUserID indicates an expected call of FetchTaskIDByUserID.
func (mr *MockTaskPermissionRepositoryMockRecorder) FetchTaskIDByUserID(ctx, id, canEdit, canRead any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchTaskIDByUserID", reflect.TypeOf((*MockTaskPermissionRepository)(nil).FetchTaskIDByUserID), ctx, id, canEdit, canRead)
}

// GrantPermission mocks base method.
func (m *MockTaskPermissionRepository) GrantPermission(ctx context.Context, taskPermission *domain.TaskPermission) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GrantPermission", ctx, taskPermission)
	ret0, _ := ret[0].(error)
	return ret0
}

// GrantPermission indicates an expected call of GrantPermission.
func (mr *MockTaskPermissionRepositoryMockRecorder) GrantPermission(ctx, taskPermission any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GrantPermission", reflect.TypeOf((*MockTaskPermissionRepository)(nil).GrantPermission), ctx, taskPermission)
}
