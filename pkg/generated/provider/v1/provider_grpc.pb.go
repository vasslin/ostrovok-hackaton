package providerv1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

const _ = grpc.SupportPackageIsVersion9

const (
	ProviderService_GetCapabilities_FullMethodName  = "/godnota.provider.v1.ProviderService/GetCapabilities"
	ProviderService_SearchParts_FullMethodName      = "/godnota.provider.v1.ProviderService/SearchParts"
	ProviderService_SearchTransport_FullMethodName  = "/godnota.provider.v1.ProviderService/SearchTransport"
	ProviderService_SearchHotels_FullMethodName     = "/godnota.provider.v1.ProviderService/SearchHotels"
	ProviderService_SearchActivities_FullMethodName = "/godnota.provider.v1.ProviderService/SearchActivities"
)

type ProviderServiceClient interface {
	GetCapabilities(ctx context.Context, in *structpb.Struct, opts ...grpc.CallOption) (*structpb.Struct, error)
	SearchParts(ctx context.Context, in *SearchPartsRequest, opts ...grpc.CallOption) (*SearchPartsResponse, error)
	SearchTransport(ctx context.Context, in *SearchTransportRequest, opts ...grpc.CallOption) (*SearchTransportResponse, error)
	SearchHotels(ctx context.Context, in *SearchHotelsRequest, opts ...grpc.CallOption) (*SearchHotelsResponse, error)
	SearchActivities(ctx context.Context, in *SearchActivitiesRequest, opts ...grpc.CallOption) (*SearchActivitiesResponse, error)
}

type providerServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewProviderServiceClient(cc grpc.ClientConnInterface) ProviderServiceClient {
	return &providerServiceClient{cc}
}

func (c *providerServiceClient) GetCapabilities(ctx context.Context, in *structpb.Struct, opts ...grpc.CallOption) (*structpb.Struct, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(structpb.Struct)
	err := c.cc.Invoke(ctx, ProviderService_GetCapabilities_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *providerServiceClient) SearchParts(ctx context.Context, in *SearchPartsRequest, opts ...grpc.CallOption) (*SearchPartsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SearchPartsResponse)
	err := c.cc.Invoke(ctx, ProviderService_SearchParts_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *providerServiceClient) SearchTransport(ctx context.Context, in *SearchTransportRequest, opts ...grpc.CallOption) (*SearchTransportResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SearchTransportResponse)
	err := c.cc.Invoke(ctx, ProviderService_SearchTransport_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *providerServiceClient) SearchHotels(ctx context.Context, in *SearchHotelsRequest, opts ...grpc.CallOption) (*SearchHotelsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SearchHotelsResponse)
	err := c.cc.Invoke(ctx, ProviderService_SearchHotels_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *providerServiceClient) SearchActivities(ctx context.Context, in *SearchActivitiesRequest, opts ...grpc.CallOption) (*SearchActivitiesResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SearchActivitiesResponse)
	err := c.cc.Invoke(ctx, ProviderService_SearchActivities_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type ProviderServiceServer interface {
	GetCapabilities(context.Context, *structpb.Struct) (*structpb.Struct, error)
	SearchParts(context.Context, *SearchPartsRequest) (*SearchPartsResponse, error)
	SearchTransport(context.Context, *SearchTransportRequest) (*SearchTransportResponse, error)
	SearchHotels(context.Context, *SearchHotelsRequest) (*SearchHotelsResponse, error)
	SearchActivities(context.Context, *SearchActivitiesRequest) (*SearchActivitiesResponse, error)
}

type UnimplementedProviderServiceServer struct{}

func (UnimplementedProviderServiceServer) GetCapabilities(context.Context, *structpb.Struct) (*structpb.Struct, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCapabilities not implemented")
}
func (UnimplementedProviderServiceServer) SearchParts(context.Context, *SearchPartsRequest) (*SearchPartsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SearchParts not implemented")
}
func (UnimplementedProviderServiceServer) SearchTransport(context.Context, *SearchTransportRequest) (*SearchTransportResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SearchTransport not implemented")
}
func (UnimplementedProviderServiceServer) SearchHotels(context.Context, *SearchHotelsRequest) (*SearchHotelsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SearchHotels not implemented")
}
func (UnimplementedProviderServiceServer) SearchActivities(context.Context, *SearchActivitiesRequest) (*SearchActivitiesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SearchActivities not implemented")
}
func (UnimplementedProviderServiceServer) testEmbeddedByValue() {}

type UnsafeProviderServiceServer interface {
	mustEmbedUnimplementedProviderServiceServer()
}

func RegisterProviderServiceServer(s grpc.ServiceRegistrar, srv ProviderServiceServer) {

	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&ProviderService_ServiceDesc, srv)
}

func _ProviderService_GetCapabilities_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(structpb.Struct)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProviderServiceServer).GetCapabilities(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ProviderService_GetCapabilities_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProviderServiceServer).GetCapabilities(ctx, req.(*structpb.Struct))
	}
	return interceptor(ctx, in, info, handler)
}

func _ProviderService_SearchParts_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SearchPartsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProviderServiceServer).SearchParts(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ProviderService_SearchParts_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProviderServiceServer).SearchParts(ctx, req.(*SearchPartsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ProviderService_SearchTransport_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SearchTransportRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProviderServiceServer).SearchTransport(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ProviderService_SearchTransport_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProviderServiceServer).SearchTransport(ctx, req.(*SearchTransportRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ProviderService_SearchHotels_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SearchHotelsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProviderServiceServer).SearchHotels(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ProviderService_SearchHotels_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProviderServiceServer).SearchHotels(ctx, req.(*SearchHotelsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ProviderService_SearchActivities_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SearchActivitiesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProviderServiceServer).SearchActivities(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ProviderService_SearchActivities_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProviderServiceServer).SearchActivities(ctx, req.(*SearchActivitiesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var ProviderService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "godnota.provider.v1.ProviderService",
	HandlerType: (*ProviderServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetCapabilities",
			Handler:    _ProviderService_GetCapabilities_Handler,
		},
		{
			MethodName: "SearchParts",
			Handler:    _ProviderService_SearchParts_Handler,
		},
		{
			MethodName: "SearchTransport",
			Handler:    _ProviderService_SearchTransport_Handler,
		},
		{
			MethodName: "SearchHotels",
			Handler:    _ProviderService_SearchHotels_Handler,
		},
		{
			MethodName: "SearchActivities",
			Handler:    _ProviderService_SearchActivities_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/provider/v1/provider.proto",
}
