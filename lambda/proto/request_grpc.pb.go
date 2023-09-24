// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v4.24.3
// source: request.proto

package proto

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

// LambdaServerClient is the client API for LambdaServer service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type LambdaServerClient interface {
	MakeRequest(ctx context.Context, in *Payload, opts ...grpc.CallOption) (*Payload, error)
}

type lambdaServerClient struct {
	cc grpc.ClientConnInterface
}

func NewLambdaServerClient(cc grpc.ClientConnInterface) LambdaServerClient {
	return &lambdaServerClient{cc}
}

func (c *lambdaServerClient) MakeRequest(ctx context.Context, in *Payload, opts ...grpc.CallOption) (*Payload, error) {
	out := new(Payload)
	err := c.cc.Invoke(ctx, "/lambda.LambdaServer/MakeRequest", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LambdaServerServer is the server API for LambdaServer service.
// All implementations must embed UnimplementedLambdaServerServer
// for forward compatibility
type LambdaServerServer interface {
	MakeRequest(context.Context, *Payload) (*Payload, error)
	mustEmbedUnimplementedLambdaServerServer()
}

// UnimplementedLambdaServerServer must be embedded to have forward compatible implementations.
type UnimplementedLambdaServerServer struct {
}

func (UnimplementedLambdaServerServer) MakeRequest(context.Context, *Payload) (*Payload, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MakeRequest not implemented")
}
func (UnimplementedLambdaServerServer) mustEmbedUnimplementedLambdaServerServer() {}

// UnsafeLambdaServerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to LambdaServerServer will
// result in compilation errors.
type UnsafeLambdaServerServer interface {
	mustEmbedUnimplementedLambdaServerServer()
}

func RegisterLambdaServerServer(s grpc.ServiceRegistrar, srv LambdaServerServer) {
	s.RegisterService(&LambdaServer_ServiceDesc, srv)
}

func _LambdaServer_MakeRequest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Payload)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LambdaServerServer).MakeRequest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/lambda.LambdaServer/MakeRequest",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LambdaServerServer).MakeRequest(ctx, req.(*Payload))
	}
	return interceptor(ctx, in, info, handler)
}

// LambdaServer_ServiceDesc is the grpc.ServiceDesc for LambdaServer service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var LambdaServer_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "lambda.LambdaServer",
	HandlerType: (*LambdaServerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "MakeRequest",
			Handler:    _LambdaServer_MakeRequest_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "request.proto",
}