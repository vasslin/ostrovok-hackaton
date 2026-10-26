package travelv1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

const _ = grpc.SupportPackageIsVersion9

const (
	TravelAI_ExtractConfig_FullMethodName = "/travel.v1.TravelAI/ExtractConfig"
	TravelAI_RankHotels_FullMethodName    = "/travel.v1.TravelAI/RankHotels"
)

type TravelAIClient interface {
	ExtractConfig(ctx context.Context, in *ExtractConfigRequest, opts ...grpc.CallOption) (*ExtractConfigResponse, error)
	RankHotels(ctx context.Context, in *RankHotelsRequest, opts ...grpc.CallOption) (*RankHotelsResponse, error)
}

type travelAIClient struct {
	cc grpc.ClientConnInterface
}

func NewTravelAIClient(cc grpc.ClientConnInterface) TravelAIClient {
	return &travelAIClient{cc}
}

func (c *travelAIClient) ExtractConfig(ctx context.Context, in *ExtractConfigRequest, opts ...grpc.CallOption) (*ExtractConfigResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ExtractConfigResponse)
	err := c.cc.Invoke(ctx, TravelAI_ExtractConfig_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *travelAIClient) RankHotels(ctx context.Context, in *RankHotelsRequest, opts ...grpc.CallOption) (*RankHotelsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RankHotelsResponse)
	err := c.cc.Invoke(ctx, TravelAI_RankHotels_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type TravelAIServer interface {
	ExtractConfig(context.Context, *ExtractConfigRequest) (*ExtractConfigResponse, error)
	RankHotels(context.Context, *RankHotelsRequest) (*RankHotelsResponse, error)
	mustEmbedUnimplementedTravelAIServer()
}

type UnimplementedTravelAIServer struct{}

func (UnimplementedTravelAIServer) ExtractConfig(context.Context, *ExtractConfigRequest) (*ExtractConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ExtractConfig not implemented")
}
func (UnimplementedTravelAIServer) RankHotels(context.Context, *RankHotelsRequest) (*RankHotelsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RankHotels not implemented")
}
func (UnimplementedTravelAIServer) mustEmbedUnimplementedTravelAIServer() {}
func (UnimplementedTravelAIServer) testEmbeddedByValue()                  {}

type UnsafeTravelAIServer interface {
	mustEmbedUnimplementedTravelAIServer()
}

func RegisterTravelAIServer(s grpc.ServiceRegistrar, srv TravelAIServer) {

	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&TravelAI_ServiceDesc, srv)
}

func _TravelAI_ExtractConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExtractConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TravelAIServer).ExtractConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TravelAI_ExtractConfig_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TravelAIServer).ExtractConfig(ctx, req.(*ExtractConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TravelAI_RankHotels_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RankHotelsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TravelAIServer).RankHotels(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TravelAI_RankHotels_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TravelAIServer).RankHotels(ctx, req.(*RankHotelsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var TravelAI_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "travel.v1.TravelAI",
	HandlerType: (*TravelAIServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ExtractConfig",
			Handler:    _TravelAI_ExtractConfig_Handler,
		},
		{
			MethodName: "RankHotels",
			Handler:    _TravelAI_RankHotels_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/travel/v1/travel.proto",
}
