// Code generated by MockGen. DO NOT EDIT.
// Source: domain/repository/user_repository.go

// Package mock_repository is a generated GoMock package.
package mock_repository

import (
	entity "pinterest/domain/entity"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockUserRepository is a mock of UserRepository interface.
type MockUserRepository struct {
	ctrl     *gomock.Controller
	recorder *MockUserRepositoryMockRecorder
}

// MockUserRepositoryMockRecorder is the mock recorder for MockUserRepository.
type MockUserRepositoryMockRecorder struct {
	mock *MockUserRepository
}

// NewMockUserRepository creates a new mock instance.
func NewMockUserRepository(ctrl *gomock.Controller) *MockUserRepository {
	mock := &MockUserRepository{ctrl: ctrl}
	mock.recorder = &MockUserRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserRepository) EXPECT() *MockUserRepositoryMockRecorder {
	return m.recorder
}

// DeleteUser mocks base method.
func (m *MockUserRepository) DeleteUser(arg0 int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteUser", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteUser indicates an expected call of DeleteUser.
func (mr *MockUserRepositoryMockRecorder) DeleteUser(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteUser", reflect.TypeOf((*MockUserRepository)(nil).DeleteUser), arg0)
}

// GetUser mocks base method.
func (m *MockUserRepository) GetUser(arg0 int) (*entity.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUser", arg0)
	ret0, _ := ret[0].(*entity.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUser indicates an expected call of GetUser.
func (mr *MockUserRepositoryMockRecorder) GetUser(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUser", reflect.TypeOf((*MockUserRepository)(nil).GetUser), arg0)
}

// GetUserByUsername mocks base method.
func (m *MockUserRepository) GetUserByUsername(arg0 string) (*entity.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserByUsername", arg0)
	ret0, _ := ret[0].(*entity.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserByUsername indicates an expected call of GetUserByUsername.
func (mr *MockUserRepositoryMockRecorder) GetUserByUsername(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserByUsername", reflect.TypeOf((*MockUserRepository)(nil).GetUserByUsername), arg0)
}

// GetUsers mocks base method.
func (m *MockUserRepository) GetUsers() ([]entity.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUsers")
	ret0, _ := ret[0].([]entity.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUsers indicates an expected call of GetUsers.
func (mr *MockUserRepositoryMockRecorder) GetUsers() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUsers", reflect.TypeOf((*MockUserRepository)(nil).GetUsers))
}

// SaveUser mocks base method.
func (m *MockUserRepository) SaveUser(arg0 *entity.User) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveUser", arg0)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SaveUser indicates an expected call of SaveUser.
func (mr *MockUserRepositoryMockRecorder) SaveUser(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveUser", reflect.TypeOf((*MockUserRepository)(nil).SaveUser), arg0)
}
