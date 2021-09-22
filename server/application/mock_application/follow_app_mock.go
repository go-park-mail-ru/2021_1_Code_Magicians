// Code generated by MockGen. DO NOT EDIT.
// Source: application/follow_app.go

// Package mock_application is a generated GoMock package.
package mock_application

import (
	entity "pinterest/domain/entity"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockFollowAppInterface is a mock of FollowAppInterface interface.
type MockFollowAppInterface struct {
	ctrl     *gomock.Controller
	recorder *MockFollowAppInterfaceMockRecorder
}

// MockFollowAppInterfaceMockRecorder is the mock recorder for MockFollowAppInterface.
type MockFollowAppInterfaceMockRecorder struct {
	mock *MockFollowAppInterface
}

// NewMockFollowAppInterface creates a new mock instance.
func NewMockFollowAppInterface(ctrl *gomock.Controller) *MockFollowAppInterface {
	mock := &MockFollowAppInterface{ctrl: ctrl}
	mock.recorder = &MockFollowAppInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFollowAppInterface) EXPECT() *MockFollowAppInterfaceMockRecorder {
	return m.recorder
}

// CheckIfFollowed mocks base method.
func (m *MockFollowAppInterface) CheckIfFollowed(followerID, followedID int) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckIfFollowed", followerID, followedID)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CheckIfFollowed indicates an expected call of CheckIfFollowed.
func (mr *MockFollowAppInterfaceMockRecorder) CheckIfFollowed(followerID, followedID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckIfFollowed", reflect.TypeOf((*MockFollowAppInterface)(nil).CheckIfFollowed), followerID, followedID)
}

// Follow mocks base method.
func (m *MockFollowAppInterface) Follow(followerID, followedID int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Follow", followerID, followedID)
	ret0, _ := ret[0].(error)
	return ret0
}

// Follow indicates an expected call of Follow.
func (mr *MockFollowAppInterfaceMockRecorder) Follow(followerID, followedID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Follow", reflect.TypeOf((*MockFollowAppInterface)(nil).Follow), followerID, followedID)
}

// GetAllFollowed mocks base method.
func (m *MockFollowAppInterface) GetAllFollowed(followerID int) ([]entity.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllFollowed", followerID)
	ret0, _ := ret[0].([]entity.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllFollowed indicates an expected call of GetAllFollowed.
func (mr *MockFollowAppInterfaceMockRecorder) GetAllFollowed(followerID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllFollowed", reflect.TypeOf((*MockFollowAppInterface)(nil).GetAllFollowed), followerID)
}

// GetAllFollowers mocks base method.
func (m *MockFollowAppInterface) GetAllFollowers(followedID int) ([]entity.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllFollowers", followedID)
	ret0, _ := ret[0].([]entity.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllFollowers indicates an expected call of GetAllFollowers.
func (mr *MockFollowAppInterfaceMockRecorder) GetAllFollowers(followedID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllFollowers", reflect.TypeOf((*MockFollowAppInterface)(nil).GetAllFollowers), followedID)
}

// GetPinsOfFollowedUsers mocks base method.
func (m *MockFollowAppInterface) GetPinsOfFollowedUsers(userID int) ([]entity.Pin, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPinsOfFollowedUsers", userID)
	ret0, _ := ret[0].([]entity.Pin)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPinsOfFollowedUsers indicates an expected call of GetPinsOfFollowedUsers.
func (mr *MockFollowAppInterfaceMockRecorder) GetPinsOfFollowedUsers(userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPinsOfFollowedUsers", reflect.TypeOf((*MockFollowAppInterface)(nil).GetPinsOfFollowedUsers), userID)
}

// Unfollow mocks base method.
func (m *MockFollowAppInterface) Unfollow(followerID, followedID int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Unfollow", followerID, followedID)
	ret0, _ := ret[0].(error)
	return ret0
}

// Unfollow indicates an expected call of Unfollow.
func (mr *MockFollowAppInterfaceMockRecorder) Unfollow(followerID, followedID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unfollow", reflect.TypeOf((*MockFollowAppInterface)(nil).Unfollow), followerID, followedID)
}
