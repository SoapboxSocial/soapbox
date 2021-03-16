// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/rooms/pb/room_api_grpc.pb.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	pb "github.com/soapboxsocial/soapbox/pkg/rooms/pb"
	grpc "google.golang.org/grpc"
	reflect "reflect"
)

// MockRoomServiceClient is a mock of RoomServiceClient interface
type MockRoomServiceClient struct {
	ctrl     *gomock.Controller
	recorder *MockRoomServiceClientMockRecorder
}

// MockRoomServiceClientMockRecorder is the mock recorder for MockRoomServiceClient
type MockRoomServiceClientMockRecorder struct {
	mock *MockRoomServiceClient
}

// NewMockRoomServiceClient creates a new mock instance
func NewMockRoomServiceClient(ctrl *gomock.Controller) *MockRoomServiceClient {
	mock := &MockRoomServiceClient{ctrl: ctrl}
	mock.recorder = &MockRoomServiceClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockRoomServiceClient) EXPECT() *MockRoomServiceClientMockRecorder {
	return m.recorder
}

// GetRoom mocks base method
func (m *MockRoomServiceClient) GetRoom(ctx context.Context, in *pb.GetRoomRequest, opts ...grpc.CallOption) (*pb.GetRoomResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetRoom", varargs...)
	ret0, _ := ret[0].(*pb.GetRoomResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRoom indicates an expected call of GetRoom
func (mr *MockRoomServiceClientMockRecorder) GetRoom(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRoom", reflect.TypeOf((*MockRoomServiceClient)(nil).GetRoom), varargs...)
}

// RegisterWelcomeRoom mocks base method
func (m *MockRoomServiceClient) RegisterWelcomeRoom(ctx context.Context, in *pb.RegisterWelcomeRoomRequest, opts ...grpc.CallOption) (*pb.RegisterWelcomeRoomResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "RegisterWelcomeRoom", varargs...)
	ret0, _ := ret[0].(*pb.RegisterWelcomeRoomResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RegisterWelcomeRoom indicates an expected call of RegisterWelcomeRoom
func (mr *MockRoomServiceClientMockRecorder) RegisterWelcomeRoom(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterWelcomeRoom", reflect.TypeOf((*MockRoomServiceClient)(nil).RegisterWelcomeRoom), varargs...)
}

// MockRoomServiceServer is a mock of RoomServiceServer interface
type MockRoomServiceServer struct {
	ctrl     *gomock.Controller
	recorder *MockRoomServiceServerMockRecorder
}

// MockRoomServiceServerMockRecorder is the mock recorder for MockRoomServiceServer
type MockRoomServiceServerMockRecorder struct {
	mock *MockRoomServiceServer
}

// NewMockRoomServiceServer creates a new mock instance
func NewMockRoomServiceServer(ctrl *gomock.Controller) *MockRoomServiceServer {
	mock := &MockRoomServiceServer{ctrl: ctrl}
	mock.recorder = &MockRoomServiceServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockRoomServiceServer) EXPECT() *MockRoomServiceServerMockRecorder {
	return m.recorder
}

// GetRoom mocks base method
func (m *MockRoomServiceServer) GetRoom(arg0 context.Context, arg1 *pb.GetRoomRequest) (*pb.GetRoomResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRoom", arg0, arg1)
	ret0, _ := ret[0].(*pb.GetRoomResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRoom indicates an expected call of GetRoom
func (mr *MockRoomServiceServerMockRecorder) GetRoom(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRoom", reflect.TypeOf((*MockRoomServiceServer)(nil).GetRoom), arg0, arg1)
}

// RegisterWelcomeRoom mocks base method
func (m *MockRoomServiceServer) RegisterWelcomeRoom(arg0 context.Context, arg1 *pb.RegisterWelcomeRoomRequest) (*pb.RegisterWelcomeRoomResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RegisterWelcomeRoom", arg0, arg1)
	ret0, _ := ret[0].(*pb.RegisterWelcomeRoomResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RegisterWelcomeRoom indicates an expected call of RegisterWelcomeRoom
func (mr *MockRoomServiceServerMockRecorder) RegisterWelcomeRoom(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterWelcomeRoom", reflect.TypeOf((*MockRoomServiceServer)(nil).RegisterWelcomeRoom), arg0, arg1)
}

// mustEmbedUnimplementedRoomServiceServer mocks base method
func (m *MockRoomServiceServer) mustEmbedUnimplementedRoomServiceServer() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "mustEmbedUnimplementedRoomServiceServer")
}

// mustEmbedUnimplementedRoomServiceServer indicates an expected call of mustEmbedUnimplementedRoomServiceServer
func (mr *MockRoomServiceServerMockRecorder) mustEmbedUnimplementedRoomServiceServer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "mustEmbedUnimplementedRoomServiceServer", reflect.TypeOf((*MockRoomServiceServer)(nil).mustEmbedUnimplementedRoomServiceServer))
}

// MockUnsafeRoomServiceServer is a mock of UnsafeRoomServiceServer interface
type MockUnsafeRoomServiceServer struct {
	ctrl     *gomock.Controller
	recorder *MockUnsafeRoomServiceServerMockRecorder
}

// MockUnsafeRoomServiceServerMockRecorder is the mock recorder for MockUnsafeRoomServiceServer
type MockUnsafeRoomServiceServerMockRecorder struct {
	mock *MockUnsafeRoomServiceServer
}

// NewMockUnsafeRoomServiceServer creates a new mock instance
func NewMockUnsafeRoomServiceServer(ctrl *gomock.Controller) *MockUnsafeRoomServiceServer {
	mock := &MockUnsafeRoomServiceServer{ctrl: ctrl}
	mock.recorder = &MockUnsafeRoomServiceServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockUnsafeRoomServiceServer) EXPECT() *MockUnsafeRoomServiceServerMockRecorder {
	return m.recorder
}

// mustEmbedUnimplementedRoomServiceServer mocks base method
func (m *MockUnsafeRoomServiceServer) mustEmbedUnimplementedRoomServiceServer() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "mustEmbedUnimplementedRoomServiceServer")
}

// mustEmbedUnimplementedRoomServiceServer indicates an expected call of mustEmbedUnimplementedRoomServiceServer
func (mr *MockUnsafeRoomServiceServerMockRecorder) mustEmbedUnimplementedRoomServiceServer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "mustEmbedUnimplementedRoomServiceServer", reflect.TypeOf((*MockUnsafeRoomServiceServer)(nil).mustEmbedUnimplementedRoomServiceServer))
}
