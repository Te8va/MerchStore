// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/Te8va/MerchStore/internal/domain (interfaces: MerchService)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"

	domain "github.com/Te8va/MerchStore/internal/domain"
)

// MockMerchService is a mock of MerchService interface.
type MockMerchService struct {
	ctrl     *gomock.Controller
	recorder *MockMerchServiceMockRecorder
}

// MockMerchServiceMockRecorder is the mock recorder for MockMerchService.
type MockMerchServiceMockRecorder struct {
	mock *MockMerchService
}

// NewMockMerchService creates a new mock instance.
func NewMockMerchService(ctrl *gomock.Controller) *MockMerchService {
	mock := &MockMerchService{ctrl: ctrl}
	mock.recorder = &MockMerchServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMerchService) EXPECT() *MockMerchServiceMockRecorder {
	return m.recorder
}

// BuyMerch mocks base method.
func (m *MockMerchService) BuyMerch(arg0 context.Context, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BuyMerch", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// BuyMerch indicates an expected call of BuyMerch.
func (mr *MockMerchServiceMockRecorder) BuyMerch(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BuyMerch", reflect.TypeOf((*MockMerchService)(nil).BuyMerch), arg0, arg1, arg2)
}

// GetUserInfo mocks base method.
func (m *MockMerchService) GetUserInfo(arg0 context.Context, arg1 string) (domain.UserInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserInfo", arg0, arg1)
	ret0, _ := ret[0].(domain.UserInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserInfo indicates an expected call of GetUserInfo.
func (mr *MockMerchServiceMockRecorder) GetUserInfo(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserInfo", reflect.TypeOf((*MockMerchService)(nil).GetUserInfo), arg0, arg1)
}

// SendCoin mocks base method.
func (m *MockMerchService) SendCoin(arg0 context.Context, arg1, arg2 string, arg3 int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendCoin", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendCoin indicates an expected call of SendCoin.
func (mr *MockMerchServiceMockRecorder) SendCoin(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendCoin", reflect.TypeOf((*MockMerchService)(nil).SendCoin), arg0, arg1, arg2, arg3)
}
