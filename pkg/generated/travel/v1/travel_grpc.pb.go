package travelv1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

const _ = grpc.SupportPackageIsVersion9

const (
	TravelService_PlanTrip_FullMethodName      = "/godnota.travel.v1.TravelService/PlanTrip"
	TravelService_SubmitAnswers_FullMethodName = "/godnota.travel.v1.TravelService/SubmitAnswers"
	TravelService_SelectIdea_FullMethodName    = "/godnota.travel.v1.TravelService/SelectIdea"
	TravelService_GetConfig_FullMethodName     = "/godnota.travel.v1.TravelService/GetConfig"
	TravelService_UpdateConfig_FullMethodName  = "/godnota.travel.v1.TravelService/UpdateConfig"
	TravelService_ConfirmConfig_FullMethodName = "/godnota.travel.v1.TravelService/ConfirmConfig"
	TravelService_SearchRoutes_FullMethodName  = "/godnota.travel.v1.TravelService/SearchRoutes"
	TravelService_GetRoutes_FullMethodName     = "/godnota.travel.v1.TravelService/GetRoutes"
)

type TravelServiceClient interface {
	PlanTrip(ctx context.Context, in *PlanTripRequest, opts ...grpc.CallOption) (*PlanTripResponse, error)
	SubmitAnswers(ctx context.Context, in *SubmitAnswersRequest, opts ...grpc.CallOption) (*SubmitAnswersResponse, error)
	SelectIdea(ctx context.Context, in *SelectIdeaRequest, opts ...grpc.CallOption) (*SelectIdeaResponse, error)
	GetConfig(ctx context.Context, in *GetConfigRequest, opts ...grpc.CallOption) (*GetConfigResponse, error)
	UpdateConfig(ctx context.Context, in *UpdateConfigRequest, opts ...grpc.CallOption) (*UpdateConfigResponse, error)
	ConfirmConfig(ctx context.Context, in *ConfirmConfigRequest, opts ...grpc.CallOption) (*ConfirmConfigResponse, error)
	SearchRoutes(ctx context.Context, in *SearchRoutesRequest, opts ...grpc.CallOption) (*SearchRoutesResponse, error)
	GetRoutes(ctx context.Context, in *GetRoutesRequest, opts ...grpc.CallOption) (*GetRoutesResponse, error)
}

type travelServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewTravelServiceClient(cc grpc.ClientConnInterface) TravelServiceClient {
	return &travelServiceClient{cc}
}

func (c *travelServiceClient) PlanTrip(ctx context.Context, in *PlanTripRequest, opts ...grpc.CallOption) (*PlanTripResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PlanTripResponse)
	err := c.cc.Invoke(ctx, TravelService_PlanTrip_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *travelServiceClient) SubmitAnswers(ctx context.Context, in *SubmitAnswersRequest, opts ...grpc.CallOption) (*SubmitAnswersResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SubmitAnswersResponse)
	err := c.cc.Invoke(ctx, TravelService_SubmitAnswers_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *travelServiceClient) SelectIdea(ctx context.Context, in *SelectIdeaRequest, opts ...grpc.CallOption) (*SelectIdeaResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SelectIdeaResponse)
	err := c.cc.Invoke(ctx, TravelService_SelectIdea_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *travelServiceClient) GetConfig(ctx context.Context, in *GetConfigRequest, opts ...grpc.CallOption) (*GetConfigResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetConfigResponse)
	err := c.cc.Invoke(ctx, TravelService_GetConfig_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *travelServiceClient) UpdateConfig(ctx context.Context, in *UpdateConfigRequest, opts ...grpc.CallOption) (*UpdateConfigResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UpdateConfigResponse)
	err := c.cc.Invoke(ctx, TravelService_UpdateConfig_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *travelServiceClient) ConfirmConfig(ctx context.Context, in *ConfirmConfigRequest, opts ...grpc.CallOption) (*ConfirmConfigResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ConfirmConfigResponse)
	err := c.cc.Invoke(ctx, TravelService_ConfirmConfig_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *travelServiceClient) SearchRoutes(ctx context.Context, in *SearchRoutesRequest, opts ...grpc.CallOption) (*SearchRoutesResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SearchRoutesResponse)
	err := c.cc.Invoke(ctx, TravelService_SearchRoutes_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *travelServiceClient) GetRoutes(ctx context.Context, in *GetRoutesRequest, opts ...grpc.CallOption) (*GetRoutesResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetRoutesResponse)
	err := c.cc.Invoke(ctx, TravelService_GetRoutes_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type TravelServiceServer interface {
	PlanTrip(context.Context, *PlanTripRequest) (*PlanTripResponse, error)
	SubmitAnswers(context.Context, *SubmitAnswersRequest) (*SubmitAnswersResponse, error)
	SelectIdea(context.Context, *SelectIdeaRequest) (*SelectIdeaResponse, error)
	GetConfig(context.Context, *GetConfigRequest) (*GetConfigResponse, error)
	UpdateConfig(context.Context, *UpdateConfigRequest) (*UpdateConfigResponse, error)
	ConfirmConfig(context.Context, *ConfirmConfigRequest) (*ConfirmConfigResponse, error)
	SearchRoutes(context.Context, *SearchRoutesRequest) (*SearchRoutesResponse, error)
	GetRoutes(context.Context, *GetRoutesRequest) (*GetRoutesResponse, error)
}
type UnimplementedTravelServiceServer struct{}

func (UnimplementedTravelServiceServer) PlanTrip(context.Context, *PlanTripRequest) (*PlanTripResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PlanTrip not implemented")
}
func (UnimplementedTravelServiceServer) SubmitAnswers(context.Context, *SubmitAnswersRequest) (*SubmitAnswersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitAnswers not implemented")
}
func (UnimplementedTravelServiceServer) SelectIdea(context.Context, *SelectIdeaRequest) (*SelectIdeaResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SelectIdea not implemented")
}
func (UnimplementedTravelServiceServer) GetConfig(context.Context, *GetConfigRequest) (*GetConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetConfig not implemented")
}
func (UnimplementedTravelServiceServer) UpdateConfig(context.Context, *UpdateConfigRequest) (*UpdateConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateConfig not implemented")
}
func (UnimplementedTravelServiceServer) ConfirmConfig(context.Context, *ConfirmConfigRequest) (*ConfirmConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ConfirmConfig not implemented")
}
func (UnimplementedTravelServiceServer) SearchRoutes(context.Context, *SearchRoutesRequest) (*SearchRoutesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SearchRoutes not implemented")
}
func (UnimplementedTravelServiceServer) GetRoutes(context.Context, *GetRoutesRequest) (*GetRoutesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRoutes not implemented")
}
func (UnimplementedTravelServiceServer) testEmbeddedByValue() {}

type UnsafeTravelServiceServer interface {
	mustEmbedUnimplementedTravelServiceServer()
}

func RegisterTravelServiceServer(s grpc.ServiceRegistrar, srv TravelServiceServer) {
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&TravelService_ServiceDesc, srv)
}

func _TravelService_PlanTrip_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PlanTripRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TravelServiceServer).PlanTrip(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TravelService_PlanTrip_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TravelServiceServer).PlanTrip(ctx, req.(*PlanTripRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TravelService_SubmitAnswers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SubmitAnswersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TravelServiceServer).SubmitAnswers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TravelService_SubmitAnswers_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TravelServiceServer).SubmitAnswers(ctx, req.(*SubmitAnswersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TravelService_SelectIdea_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SelectIdeaRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TravelServiceServer).SelectIdea(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TravelService_SelectIdea_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TravelServiceServer).SelectIdea(ctx, req.(*SelectIdeaRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TravelService_GetConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TravelServiceServer).GetConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TravelService_GetConfig_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TravelServiceServer).GetConfig(ctx, req.(*GetConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TravelService_UpdateConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TravelServiceServer).UpdateConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TravelService_UpdateConfig_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TravelServiceServer).UpdateConfig(ctx, req.(*UpdateConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TravelService_ConfirmConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ConfirmConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TravelServiceServer).ConfirmConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TravelService_ConfirmConfig_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TravelServiceServer).ConfirmConfig(ctx, req.(*ConfirmConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TravelService_SearchRoutes_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SearchRoutesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TravelServiceServer).SearchRoutes(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TravelService_SearchRoutes_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TravelServiceServer).SearchRoutes(ctx, req.(*SearchRoutesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TravelService_GetRoutes_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRoutesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TravelServiceServer).GetRoutes(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TravelService_GetRoutes_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TravelServiceServer).GetRoutes(ctx, req.(*GetRoutesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var TravelService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "godnota.travel.v1.TravelService",
	HandlerType: (*TravelServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PlanTrip",
			Handler:    _TravelService_PlanTrip_Handler,
		},
		{
			MethodName: "SubmitAnswers",
			Handler:    _TravelService_SubmitAnswers_Handler,
		},
		{
			MethodName: "SelectIdea",
			Handler:    _TravelService_SelectIdea_Handler,
		},
		{
			MethodName: "GetConfig",
			Handler:    _TravelService_GetConfig_Handler,
		},
		{
			MethodName: "UpdateConfig",
			Handler:    _TravelService_UpdateConfig_Handler,
		},
		{
			MethodName: "ConfirmConfig",
			Handler:    _TravelService_ConfirmConfig_Handler,
		},
		{
			MethodName: "SearchRoutes",
			Handler:    _TravelService_SearchRoutes_Handler,
		},
		{
			MethodName: "GetRoutes",
			Handler:    _TravelService_GetRoutes_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/travel/v1/travel.proto",
}
