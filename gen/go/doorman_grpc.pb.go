// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.19.6
// source: doorman.proto

package _go

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	Doorman_Check_FullMethodName  = "/doorman.Doorman/Check"
	Doorman_Grant_FullMethodName  = "/doorman.Doorman/Grant"
	Doorman_Revoke_FullMethodName = "/doorman.Doorman/Revoke"
)

// DoormanClient is the client API for Doorman service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DoormanClient interface {
	Check(ctx context.Context, in *CheckRequest, opts ...grpc.CallOption) (*CheckResponse, error)
	Grant(ctx context.Context, in *GrantRequest, opts ...grpc.CallOption) (*GrantResponse, error)
	Revoke(ctx context.Context, in *RevokeRequest, opts ...grpc.CallOption) (*RevokeResponse, error)
}

type doormanClient struct {
	cc grpc.ClientConnInterface
}

func NewDoormanClient(cc grpc.ClientConnInterface) DoormanClient {
	return &doormanClient{cc}
}

func (c *doormanClient) Check(ctx context.Context, in *CheckRequest, opts ...grpc.CallOption) (*CheckResponse, error) {
	out := new(CheckResponse)
	err := c.cc.Invoke(ctx, Doorman_Check_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *doormanClient) Grant(ctx context.Context, in *GrantRequest, opts ...grpc.CallOption) (*GrantResponse, error) {
	out := new(GrantResponse)
	err := c.cc.Invoke(ctx, Doorman_Grant_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *doormanClient) Revoke(ctx context.Context, in *RevokeRequest, opts ...grpc.CallOption) (*RevokeResponse, error) {
	out := new(RevokeResponse)
	err := c.cc.Invoke(ctx, Doorman_Revoke_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DoormanServer is the server API for Doorman service.
// All implementations must embed UnimplementedDoormanServer
// for forward compatibility
type DoormanServer interface {
	Check(context.Context, *CheckRequest) (*CheckResponse, error)
	Grant(context.Context, *GrantRequest) (*GrantResponse, error)
	Revoke(context.Context, *RevokeRequest) (*RevokeResponse, error)
	mustEmbedUnimplementedDoormanServer()
}

// UnimplementedDoormanServer must be embedded to have forward compatible implementations.
type UnimplementedDoormanServer struct {
}

func (UnimplementedDoormanServer) Check(context.Context, *CheckRequest) (*CheckResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Check not implemented")
}
func (UnimplementedDoormanServer) Grant(context.Context, *GrantRequest) (*GrantResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Grant not implemented")
}
func (UnimplementedDoormanServer) Revoke(context.Context, *RevokeRequest) (*RevokeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Revoke not implemented")
}
func (UnimplementedDoormanServer) mustEmbedUnimplementedDoormanServer() {}

// UnsafeDoormanServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DoormanServer will
// result in compilation errors.
type UnsafeDoormanServer interface {
	mustEmbedUnimplementedDoormanServer()
}

func RegisterDoormanServer(s grpc.ServiceRegistrar, srv DoormanServer) {
	s.RegisterService(&Doorman_ServiceDesc, srv)
}

func _Doorman_Check_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CheckRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DoormanServer).Check(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Doorman_Check_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DoormanServer).Check(ctx, req.(*CheckRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Doorman_Grant_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GrantRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DoormanServer).Grant(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Doorman_Grant_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DoormanServer).Grant(ctx, req.(*GrantRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Doorman_Revoke_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RevokeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DoormanServer).Revoke(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Doorman_Revoke_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DoormanServer).Revoke(ctx, req.(*RevokeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Doorman_ServiceDesc is the grpc.ServiceDesc for Doorman service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Doorman_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "doorman.Doorman",
	HandlerType: (*DoormanServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Check",
			Handler:    _Doorman_Check_Handler,
		},
		{
			MethodName: "Grant",
			Handler:    _Doorman_Grant_Handler,
		},
		{
			MethodName: "Revoke",
			Handler:    _Doorman_Revoke_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "doorman.proto",
}