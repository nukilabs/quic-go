// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/nukilabs/quic-go/http3 (interfaces: QUICListener)
//
// Generated by this command:
//
//	mockgen -typed -package http3 -destination mock_quic_listener_test.go github.com/nukilabs/quic-go/http3 QUICListener
//

// Package http3 is a generated GoMock package.
package http3

import (
	context "context"
	net "net"
	reflect "reflect"

	quic "github.com/nukilabs/quic-go"
	gomock "go.uber.org/mock/gomock"
)

// MockQUICListener is a mock of QUICListener interface.
type MockQUICListener struct {
	ctrl     *gomock.Controller
	recorder *MockQUICListenerMockRecorder
	isgomock struct{}
}

// MockQUICListenerMockRecorder is the mock recorder for MockQUICListener.
type MockQUICListenerMockRecorder struct {
	mock *MockQUICListener
}

// NewMockQUICListener creates a new mock instance.
func NewMockQUICListener(ctrl *gomock.Controller) *MockQUICListener {
	mock := &MockQUICListener{ctrl: ctrl}
	mock.recorder = &MockQUICListenerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockQUICListener) EXPECT() *MockQUICListenerMockRecorder {
	return m.recorder
}

// Accept mocks base method.
func (m *MockQUICListener) Accept(arg0 context.Context) (*quic.Conn, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Accept", arg0)
	ret0, _ := ret[0].(*quic.Conn)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Accept indicates an expected call of Accept.
func (mr *MockQUICListenerMockRecorder) Accept(arg0 any) *MockQUICListenerAcceptCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Accept", reflect.TypeOf((*MockQUICListener)(nil).Accept), arg0)
	return &MockQUICListenerAcceptCall{Call: call}
}

// MockQUICListenerAcceptCall wrap *gomock.Call
type MockQUICListenerAcceptCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockQUICListenerAcceptCall) Return(arg0 *quic.Conn, arg1 error) *MockQUICListenerAcceptCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockQUICListenerAcceptCall) Do(f func(context.Context) (*quic.Conn, error)) *MockQUICListenerAcceptCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockQUICListenerAcceptCall) DoAndReturn(f func(context.Context) (*quic.Conn, error)) *MockQUICListenerAcceptCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Addr mocks base method.
func (m *MockQUICListener) Addr() net.Addr {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Addr")
	ret0, _ := ret[0].(net.Addr)
	return ret0
}

// Addr indicates an expected call of Addr.
func (mr *MockQUICListenerMockRecorder) Addr() *MockQUICListenerAddrCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Addr", reflect.TypeOf((*MockQUICListener)(nil).Addr))
	return &MockQUICListenerAddrCall{Call: call}
}

// MockQUICListenerAddrCall wrap *gomock.Call
type MockQUICListenerAddrCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockQUICListenerAddrCall) Return(arg0 net.Addr) *MockQUICListenerAddrCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockQUICListenerAddrCall) Do(f func() net.Addr) *MockQUICListenerAddrCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockQUICListenerAddrCall) DoAndReturn(f func() net.Addr) *MockQUICListenerAddrCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Close mocks base method.
func (m *MockQUICListener) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockQUICListenerMockRecorder) Close() *MockQUICListenerCloseCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockQUICListener)(nil).Close))
	return &MockQUICListenerCloseCall{Call: call}
}

// MockQUICListenerCloseCall wrap *gomock.Call
type MockQUICListenerCloseCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockQUICListenerCloseCall) Return(arg0 error) *MockQUICListenerCloseCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockQUICListenerCloseCall) Do(f func() error) *MockQUICListenerCloseCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockQUICListenerCloseCall) DoAndReturn(f func() error) *MockQUICListenerCloseCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
