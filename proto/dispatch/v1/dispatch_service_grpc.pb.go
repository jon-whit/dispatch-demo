// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: dispatch/v1/dispatch_service.proto

package dispatchv1

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
	DispatchService_DispatchCheck_FullMethodName = "/dispatch.v1.DispatchService/DispatchCheck"
)

// DispatchServiceClient is the client API for DispatchService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DispatchServiceClient interface {
	DispatchCheck(ctx context.Context, in *DispatchCheckRequest, opts ...grpc.CallOption) (*DispatchCheckResponse, error)
}

type dispatchServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewDispatchServiceClient(cc grpc.ClientConnInterface) DispatchServiceClient {
	return &dispatchServiceClient{cc}
}

func (c *dispatchServiceClient) DispatchCheck(ctx context.Context, in *DispatchCheckRequest, opts ...grpc.CallOption) (*DispatchCheckResponse, error) {
	out := new(DispatchCheckResponse)
	err := c.cc.Invoke(ctx, DispatchService_DispatchCheck_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DispatchServiceServer is the server API for DispatchService service.
// All implementations must embed UnimplementedDispatchServiceServer
// for forward compatibility
type DispatchServiceServer interface {
	DispatchCheck(context.Context, *DispatchCheckRequest) (*DispatchCheckResponse, error)
	mustEmbedUnimplementedDispatchServiceServer()
}

// UnimplementedDispatchServiceServer must be embedded to have forward compatible implementations.
type UnimplementedDispatchServiceServer struct {
}

func (UnimplementedDispatchServiceServer) DispatchCheck(context.Context, *DispatchCheckRequest) (*DispatchCheckResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DispatchCheck not implemented")
}
func (UnimplementedDispatchServiceServer) mustEmbedUnimplementedDispatchServiceServer() {}

// UnsafeDispatchServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DispatchServiceServer will
// result in compilation errors.
type UnsafeDispatchServiceServer interface {
	mustEmbedUnimplementedDispatchServiceServer()
}

func RegisterDispatchServiceServer(s grpc.ServiceRegistrar, srv DispatchServiceServer) {
	s.RegisterService(&DispatchService_ServiceDesc, srv)
}

func _DispatchService_DispatchCheck_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DispatchCheckRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DispatchServiceServer).DispatchCheck(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: DispatchService_DispatchCheck_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DispatchServiceServer).DispatchCheck(ctx, req.(*DispatchCheckRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// DispatchService_ServiceDesc is the grpc.ServiceDesc for DispatchService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DispatchService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "dispatch.v1.DispatchService",
	HandlerType: (*DispatchServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "DispatchCheck",
			Handler:    _DispatchService_DispatchCheck_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "dispatch/v1/dispatch_service.proto",
}
