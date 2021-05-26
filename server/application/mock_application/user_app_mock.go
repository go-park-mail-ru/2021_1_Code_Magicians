// Code generated by MockGen. DO NOT EDIT.
// Source: application/user_app.go

// Package mock_application is a generated GoMock package.
package mock_application

import (
	io "io"
	entity "pinterest/domain/entity"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockUserAppInterface is a mock of UserAppInterface interface.
type MockUserAppInterface struct {
	ctrl     *gomock.Controller
	recorder *MockUserAppInterfaceMockRecorder
}

// MockUserAppInterfaceMockRecorder is the mock recorder for MockUserAppInterface.
type MockUserAppInterfaceMockRecorder struct {
	mock *MockUserAppInterface
}

// NewMockUserAppInterface creates a new mock instance.
func NewMockUserAppInterface(ctrl *gomock.Controller) *MockUserAppInterface {
	mock := &MockUserAppInterface{ctrl: ctrl}
	mock.recorder = &MockUserAppInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserAppInterface) EXPECT() *MockUserAppInterfaceMockRecorder {
	return m.recorder
}

// ChangePassword mocks base method.
func (m *MockUserAppInterface) ChangePassword(user *entity.User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ChangePassword", user)
	ret0, _ := ret[0].(error)
	return ret0
}

// ChangePassword indicates an expected call of ChangePassword.
func (mr *MockUserAppInterfaceMockRecorder) ChangePassword(user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ChangePassword", reflect.TypeOf((*MockUserAppInterface)(nil).ChangePassword), user)
}

// CreateUser mocks base method.
func (m *MockUserAppInterface) CreateUser(user *entity.User) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateUser", user)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateUser indicates an expected call of CreateUser.
func (mr *MockUserAppInterfaceMockRecorder) CreateUser(user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUser", reflect.TypeOf((*MockUserAppInterface)(nil).CreateUser), user)
}

// DeleteUser mocks base method.
func (m *MockUserAppInterface) DeleteUser(userID int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteUser", userID)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteUser indicates an expected call of DeleteUser.
func (mr *MockUserAppInterfaceMockRecorder) DeleteUser(userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteUser", reflect.TypeOf((*MockUserAppInterface)(nil).DeleteUser), userID)
}

// GetUser mocks base method.
func (m *MockUserAppInterface) GetUser(userID int) (*entity.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUser", userID)
	ret0, _ := ret[0].(*entity.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUser indicates an expected call of GetUser.
func (mr *MockUserAppInterfaceMockRecorder) GetUser(userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUser", reflect.TypeOf((*MockUserAppInterface)(nil).GetUser), userID)
}

// GetUserByUsername mocks base method.
func (m *MockUserAppInterface) GetUserByUsername(username string) (*entity.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserByUsername", username)
	ret0, _ := ret[0].(*entity.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserByUsername indicates an expected call of GetUserByUsername.
func (mr *MockUserAppInterfaceMockRecorder) GetUserByUsername(username interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserByUsername", reflect.TypeOf((*MockUserAppInterface)(nil).GetUserByUsername), username)
}

// GetUsers mocks base method.
func (m *MockUserAppInterface) GetUsers() ([]entity.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUsers")
	ret0, _ := ret[0].([]entity.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUsers indicates an expected call of GetUsers.
func (mr *MockUserAppInterfaceMockRecorder) GetUsers() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUsers", reflect.TypeOf((*MockUserAppInterface)(nil).GetUsers))
}

// SaveUser mocks base method.
func (m *MockUserAppInterface) SaveUser(user *entity.User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveUser", user)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveUser indicates an expected call of SaveUser.
func (mr *MockUserAppInterfaceMockRecorder) SaveUser(user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveUser", reflect.TypeOf((*MockUserAppInterface)(nil).SaveUser), user)
}

// SearchUsers mocks base method.
func (m *MockUserAppInterface) SearchUsers(keywords string) ([]entity.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SearchUsers", keywords)
	ret0, _ := ret[0].([]entity.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SearchUsers indicates an expected call of SearchUsers.
func (mr *MockUserAppInterfaceMockRecorder) SearchUsers(keywords interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SearchUsers", reflect.TypeOf((*MockUserAppInterface)(nil).SearchUsers), keywords)
}

// UpdateAvatar mocks base method.
func (m *MockUserAppInterface) UpdateAvatar(userID int, file io.Reader, extension string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateAvatar", userID, file, extension)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateAvatar indicates an expected call of UpdateAvatar.
func (mr *MockUserAppInterfaceMockRecorder) UpdateAvatar(userID, file, extension interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateAvatar", reflect.TypeOf((*MockUserAppInterface)(nil).UpdateAvatar), userID, file, extension)
}
