// Code generated by MockGen. DO NOT EDIT.
// Source: usage/board_app.go

// Package mock_application is a generated GoMock package.
package mock_application

import (
	entity "pinterest/domain/entity"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockBoardAppInterface is a mock of BoardAppInterface interface.
type MockBoardAppInterface struct {
	ctrl     *gomock.Controller
	recorder *MockBoardAppInterfaceMockRecorder
}

// MockBoardAppInterfaceMockRecorder is the mock recorder for MockBoardAppInterface.
type MockBoardAppInterfaceMockRecorder struct {
	mock *MockBoardAppInterface
}

// NewMockBoardAppInterface creates a new mock instance.
func NewMockBoardAppInterface(ctrl *gomock.Controller) *MockBoardAppInterface {
	mock := &MockBoardAppInterface{ctrl: ctrl}
	mock.recorder = &MockBoardAppInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBoardAppInterface) EXPECT() *MockBoardAppInterfaceMockRecorder {
	return m.recorder
}

// AddBoard mocks base method.
func (m *MockBoardAppInterface) AddBoard(arg0 *entity.Board) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddBoard", arg0)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddBoard indicates an expected call of AddBoard.
func (mr *MockBoardAppInterfaceMockRecorder) AddBoard(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddBoard", reflect.TypeOf((*MockBoardAppInterface)(nil).AddBoard), arg0)
}

// CheckBoard mocks base method.
func (m *MockBoardAppInterface) CheckBoard(arg0, arg1 int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckBoard", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// CheckBoard indicates an expected call of CheckBoard.
func (mr *MockBoardAppInterfaceMockRecorder) CheckBoard(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckBoard", reflect.TypeOf((*MockBoardAppInterface)(nil).CheckBoard), arg0, arg1)
}

// DeleteBoard mocks base method.
func (m *MockBoardAppInterface) DeleteBoard(arg0, arg1 int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteBoard", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteBoard indicates an expected call of DeleteBoard.
func (mr *MockBoardAppInterfaceMockRecorder) DeleteBoard(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteBoard", reflect.TypeOf((*MockBoardAppInterface)(nil).DeleteBoard), arg0, arg1)
}

// GetBoard mocks base method.
func (m *MockBoardAppInterface) GetBoard(arg0 int) (*entity.BoardInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBoard", arg0)
	ret0, _ := ret[0].(*entity.BoardInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBoard indicates an expected call of GetBoard.
func (mr *MockBoardAppInterfaceMockRecorder) GetBoard(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBoard", reflect.TypeOf((*MockBoardAppInterface)(nil).GetBoard), arg0)
}

// GetBoards mocks base method.
func (m *MockBoardAppInterface) GetBoards(arg0 int) ([]entity.Board, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBoards", arg0)
	ret0, _ := ret[0].([]entity.Board)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBoards indicates an expected call of GetBoards.
func (mr *MockBoardAppInterfaceMockRecorder) GetBoards(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBoards", reflect.TypeOf((*MockBoardAppInterface)(nil).GetBoards), arg0)
}

// GetInitUserBoard mocks base method.
func (m *MockBoardAppInterface) GetInitUserBoard(arg0 int) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetInitUserBoard", arg0)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetInitUserBoard indicates an expected call of GetInitUserBoard.
func (mr *MockBoardAppInterfaceMockRecorder) GetInitUserBoard(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInitUserBoard", reflect.TypeOf((*MockBoardAppInterface)(nil).GetInitUserBoard), arg0)
}

// UploadBoardAvatar mocks base method.
func (m *MockBoardAppInterface) UploadBoardAvatar(arg0 int, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UploadBoardAvatar", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UploadBoardAvatar indicates an expected call of UploadBoardAvatar.
func (mr *MockBoardAppInterfaceMockRecorder) UploadBoardAvatar(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UploadBoardAvatar", reflect.TypeOf((*MockBoardAppInterface)(nil).UploadBoardAvatar), arg0, arg1)
}