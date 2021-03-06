// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/darren-west/app/user-service/controller (interfaces: UserRepository)

// Package mocks is a generated GoMock package.
package mocks

import (
	models "github.com/darren-west/app/user-service/models"
	repository "github.com/darren-west/app/user-service/repository"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockUserRepository is a mock of UserRepository interface
type MockUserRepository struct {
	ctrl     *gomock.Controller
	recorder *MockUserRepositoryMockRecorder
}

// MockUserRepositoryMockRecorder is the mock recorder for MockUserRepository
type MockUserRepositoryMockRecorder struct {
	mock *MockUserRepository
}

// NewMockUserRepository creates a new mock instance
func NewMockUserRepository(ctrl *gomock.Controller) *MockUserRepository {
	mock := &MockUserRepository{ctrl: ctrl}
	mock.recorder = &MockUserRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockUserRepository) EXPECT() *MockUserRepositoryMockRecorder {
	return m.recorder
}

// CreateUser mocks base method
func (m *MockUserRepository) CreateUser(arg0 models.UserInfo) error {
	ret := m.ctrl.Call(m, "CreateUser", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateUser indicates an expected call of CreateUser
func (mr *MockUserRepositoryMockRecorder) CreateUser(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUser", reflect.TypeOf((*MockUserRepository)(nil).CreateUser), arg0)
}

// FindUser mocks base method
func (m *MockUserRepository) FindUser(arg0 repository.Matcher) (models.UserInfo, error) {
	ret := m.ctrl.Call(m, "FindUser", arg0)
	ret0, _ := ret[0].(models.UserInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindUser indicates an expected call of FindUser
func (mr *MockUserRepositoryMockRecorder) FindUser(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindUser", reflect.TypeOf((*MockUserRepository)(nil).FindUser), arg0)
}

// ListUsers mocks base method
func (m *MockUserRepository) ListUsers(arg0 repository.Matcher) ([]models.UserInfo, error) {
	ret := m.ctrl.Call(m, "ListUsers", arg0)
	ret0, _ := ret[0].([]models.UserInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListUsers indicates an expected call of ListUsers
func (mr *MockUserRepositoryMockRecorder) ListUsers(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListUsers", reflect.TypeOf((*MockUserRepository)(nil).ListUsers), arg0)
}

// RemoveUser mocks base method
func (m *MockUserRepository) RemoveUser(arg0 repository.Matcher) error {
	ret := m.ctrl.Call(m, "RemoveUser", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveUser indicates an expected call of RemoveUser
func (mr *MockUserRepositoryMockRecorder) RemoveUser(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveUser", reflect.TypeOf((*MockUserRepository)(nil).RemoveUser), arg0)
}

// UpdateUser mocks base method
func (m *MockUserRepository) UpdateUser(arg0 models.UserInfo) error {
	ret := m.ctrl.Call(m, "UpdateUser", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateUser indicates an expected call of UpdateUser
func (mr *MockUserRepositoryMockRecorder) UpdateUser(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateUser", reflect.TypeOf((*MockUserRepository)(nil).UpdateUser), arg0)
}
