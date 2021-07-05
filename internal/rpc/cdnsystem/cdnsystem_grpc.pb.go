// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package cdnsystem

import (
	context "context"
	base "d7y.io/dragonfly/v2/internal/rpc/base"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// SeederClient is the client API for Seeder service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SeederClient interface {
	// generate seeds and return to scheduler
	ObtainSeeds(ctx context.Context, in *SeedRequest, opts ...grpc.CallOption) (Seeder_ObtainSeedsClient, error)
	// get piece tasks from cdn
	GetPieceTasks(ctx context.Context, in *base.PieceTaskRequest, opts ...grpc.CallOption) (*base.PiecePacket, error)
}

type seederClient struct {
	cc grpc.ClientConnInterface
}

func NewSeederClient(cc grpc.ClientConnInterface) SeederClient {
	return &seederClient{cc}
}

func (c *seederClient) ObtainSeeds(ctx context.Context, in *SeedRequest, opts ...grpc.CallOption) (Seeder_ObtainSeedsClient, error) {
	stream, err := c.cc.NewStream(ctx, &Seeder_ServiceDesc.Streams[0], "/cdnsystem.Seeder/ObtainSeeds", opts...)
	if err != nil {
		return nil, err
	}
	x := &seederObtainSeedsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Seeder_ObtainSeedsClient interface {
	Recv() (*PieceSeed, error)
	grpc.ClientStream
}

type seederObtainSeedsClient struct {
	grpc.ClientStream
}

func (x *seederObtainSeedsClient) Recv() (*PieceSeed, error) {
	m := new(PieceSeed)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *seederClient) GetPieceTasks(ctx context.Context, in *base.PieceTaskRequest, opts ...grpc.CallOption) (*base.PiecePacket, error) {
	out := new(base.PiecePacket)
	err := c.cc.Invoke(ctx, "/cdnsystem.Seeder/GetPieceTasks", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SeederServer is the server API for Seeder service.
// All implementations must embed UnimplementedSeederServer
// for forward compatibility
type SeederServer interface {
	// generate seeds and return to scheduler
	ObtainSeeds(*SeedRequest, Seeder_ObtainSeedsServer) error
	// get piece tasks from cdn
	GetPieceTasks(context.Context, *base.PieceTaskRequest) (*base.PiecePacket, error)
	mustEmbedUnimplementedSeederServer()
}

// UnimplementedSeederServer must be embedded to have forward compatible implementations.
type UnimplementedSeederServer struct {
}

func (UnimplementedSeederServer) ObtainSeeds(*SeedRequest, Seeder_ObtainSeedsServer) error {
	return status.Errorf(codes.Unimplemented, "method ObtainSeeds not implemented")
}
func (UnimplementedSeederServer) GetPieceTasks(context.Context, *base.PieceTaskRequest) (*base.PiecePacket, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPieceTasks not implemented")
}
func (UnimplementedSeederServer) mustEmbedUnimplementedSeederServer() {}

// UnsafeSeederServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SeederServer will
// result in compilation errors.
type UnsafeSeederServer interface {
	mustEmbedUnimplementedSeederServer()
}

func RegisterSeederServer(s grpc.ServiceRegistrar, srv SeederServer) {
	s.RegisterService(&Seeder_ServiceDesc, srv)
}

func _Seeder_ObtainSeeds_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(SeedRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(SeederServer).ObtainSeeds(m, &seederObtainSeedsServer{stream})
}

type Seeder_ObtainSeedsServer interface {
	Send(*PieceSeed) error
	grpc.ServerStream
}

type seederObtainSeedsServer struct {
	grpc.ServerStream
}

func (x *seederObtainSeedsServer) Send(m *PieceSeed) error {
	return x.ServerStream.SendMsg(m)
}

func _Seeder_GetPieceTasks_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(base.PieceTaskRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SeederServer).GetPieceTasks(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cdnsystem.Seeder/GetPieceTasks",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SeederServer).GetPieceTasks(ctx, req.(*base.PieceTaskRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Seeder_ServiceDesc is the grpc.ServiceDesc for Seeder service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Seeder_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "cdnsystem.Seeder",
	HandlerType: (*SeederServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetPieceTasks",
			Handler:    _Seeder_GetPieceTasks_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ObtainSeeds",
			Handler:       _Seeder_ObtainSeeds_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "internal/rpc/cdnsystem/cdnsystem.proto",
}
