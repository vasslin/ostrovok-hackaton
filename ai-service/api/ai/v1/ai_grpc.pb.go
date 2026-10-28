package aiv1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

const _ = grpc.SupportPackageIsVersion9

const (
	AIService_GenerateOrUpdateConfig_FullMethodName  = "/godnota.ai.v1.AIService/GenerateOrUpdateConfig"
	AIService_UpdateConfig_FullMethodName            = "/godnota.ai.v1.AIService/UpdateConfig"
	AIService_RefineConfig_FullMethodName            = "/godnota.ai.v1.AIService/RefineConfig"
	AIService_GenerateRecommendations_FullMethodName = "/godnota.ai.v1.AIService/GenerateRecommendations"
	AIService_RankRoutes_FullMethodName              = "/godnota.ai.v1.AIService/RankRoutes"
	AIService_RankJoinOptions_FullMethodName         = "/godnota.ai.v1.AIService/RankJoinOptions"
	AIService_RankFinalRoutes_FullMethodName         = "/godnota.ai.v1.AIService/RankFinalRoutes"
)

type AIServiceClient interface {
	GenerateOrUpdateConfig(ctx context.Context, in *GenerateOrUpdateConfigRequest, opts ...grpc.CallOption) (*GenerateOrUpdateConfigResponse, error)
	UpdateConfig(ctx context.Context, in *UpdateConfigRequest, opts ...grpc.CallOption) (*UpdateConfigResponse, error)
	RefineConfig(ctx context.Context, in *RefineConfigRequest, opts ...grpc.CallOption) (*RefineConfigResponse, error)
	GenerateRecommendations(ctx context.Context, in *GenerateRecommendationsRequest, opts ...grpc.CallOption) (*GenerateRecommendationsResponse, error)
	RankRoutes(ctx context.Context, in *RankRoutesRequest, opts ...grpc.CallOption) (*RankRoutesResponse, error)
	RankJoinOptions(ctx context.Context, in *RankRoutesRequest, opts ...grpc.CallOption) (*RankRoutesResponse, error)
	RankFinalRoutes(ctx context.Context, in *RankRoutesRequest, opts ...grpc.CallOption) (*RankRoutesResponse, error)
}

type aIServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAIServiceClient(cc grpc.ClientConnInterface) AIServiceClient {
	return &aIServiceClient{cc}
}

func (c *aIServiceClient) GenerateOrUpdateConfig(ctx context.Context, in *GenerateOrUpdateConfigRequest, opts ...grpc.CallOption) (*GenerateOrUpdateConfigResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GenerateOrUpdateConfigResponse)
	err := c.cc.Invoke(ctx, AIService_GenerateOrUpdateConfig_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aIServiceClient) UpdateConfig(ctx context.Context, in *UpdateConfigRequest, opts ...grpc.CallOption) (*UpdateConfigResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UpdateConfigResponse)
	err := c.cc.Invoke(ctx, AIService_UpdateConfig_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aIServiceClient) RefineConfig(ctx context.Context, in *RefineConfigRequest, opts ...grpc.CallOption) (*RefineConfigResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RefineConfigResponse)
	err := c.cc.Invoke(ctx, AIService_RefineConfig_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aIServiceClient) GenerateRecommendations(ctx context.Context, in *GenerateRecommendationsRequest, opts ...grpc.CallOption) (*GenerateRecommendationsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GenerateRecommendationsResponse)
	err := c.cc.Invoke(ctx, AIService_GenerateRecommendations_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aIServiceClient) RankRoutes(ctx context.Context, in *RankRoutesRequest, opts ...grpc.CallOption) (*RankRoutesResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RankRoutesResponse)
	err := c.cc.Invoke(ctx, AIService_RankRoutes_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aIServiceClient) RankJoinOptions(ctx context.Context, in *RankRoutesRequest, opts ...grpc.CallOption) (*RankRoutesResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RankRoutesResponse)
	err := c.cc.Invoke(ctx, AIService_RankJoinOptions_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aIServiceClient) RankFinalRoutes(ctx context.Context, in *RankRoutesRequest, opts ...grpc.CallOption) (*RankRoutesResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RankRoutesResponse)
	err := c.cc.Invoke(ctx, AIService_RankFinalRoutes_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type AIServiceServer interface {
	GenerateOrUpdateConfig(context.Context, *GenerateOrUpdateConfigRequest) (*GenerateOrUpdateConfigResponse, error)
	UpdateConfig(context.Context, *UpdateConfigRequest) (*UpdateConfigResponse, error)
	RefineConfig(context.Context, *RefineConfigRequest) (*RefineConfigResponse, error)
	GenerateRecommendations(context.Context, *GenerateRecommendationsRequest) (*GenerateRecommendationsResponse, error)
	RankRoutes(context.Context, *RankRoutesRequest) (*RankRoutesResponse, error)
	RankJoinOptions(context.Context, *RankRoutesRequest) (*RankRoutesResponse, error)
	RankFinalRoutes(context.Context, *RankRoutesRequest) (*RankRoutesResponse, error)
	mustEmbedUnimplementedAIServiceServer()
}

type UnimplementedAIServiceServer struct{}

func (UnimplementedAIServiceServer) GenerateOrUpdateConfig(context.Context, *GenerateOrUpdateConfigRequest) (*GenerateOrUpdateConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GenerateOrUpdateConfig not implemented")
}
func (UnimplementedAIServiceServer) UpdateConfig(context.Context, *UpdateConfigRequest) (*UpdateConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateConfig not implemented")
}
func (UnimplementedAIServiceServer) RefineConfig(context.Context, *RefineConfigRequest) (*RefineConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RefineConfig not implemented")
}
func (UnimplementedAIServiceServer) GenerateRecommendations(context.Context, *GenerateRecommendationsRequest) (*GenerateRecommendationsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GenerateRecommendations not implemented")
}
func (UnimplementedAIServiceServer) RankRoutes(context.Context, *RankRoutesRequest) (*RankRoutesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RankRoutes not implemented")
}
func (UnimplementedAIServiceServer) RankJoinOptions(context.Context, *RankRoutesRequest) (*RankRoutesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RankJoinOptions not implemented")
}
func (UnimplementedAIServiceServer) RankFinalRoutes(context.Context, *RankRoutesRequest) (*RankRoutesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RankFinalRoutes not implemented")
}
func (UnimplementedAIServiceServer) mustEmbedUnimplementedAIServiceServer() {}
func (UnimplementedAIServiceServer) testEmbeddedByValue()                   {}

type UnsafeAIServiceServer interface {
	mustEmbedUnimplementedAIServiceServer()
}

func RegisterAIServiceServer(s grpc.ServiceRegistrar, srv AIServiceServer) {

	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&AIService_ServiceDesc, srv)
}

func _AIService_GenerateOrUpdateConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GenerateOrUpdateConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AIServiceServer).GenerateOrUpdateConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AIService_GenerateOrUpdateConfig_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AIServiceServer).GenerateOrUpdateConfig(ctx, req.(*GenerateOrUpdateConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AIService_UpdateConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AIServiceServer).UpdateConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AIService_UpdateConfig_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AIServiceServer).UpdateConfig(ctx, req.(*UpdateConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AIService_RefineConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RefineConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AIServiceServer).RefineConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AIService_RefineConfig_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AIServiceServer).RefineConfig(ctx, req.(*RefineConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AIService_GenerateRecommendations_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GenerateRecommendationsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AIServiceServer).GenerateRecommendations(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AIService_GenerateRecommendations_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AIServiceServer).GenerateRecommendations(ctx, req.(*GenerateRecommendationsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AIService_RankRoutes_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RankRoutesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AIServiceServer).RankRoutes(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AIService_RankRoutes_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AIServiceServer).RankRoutes(ctx, req.(*RankRoutesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AIService_RankJoinOptions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RankRoutesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AIServiceServer).RankJoinOptions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AIService_RankJoinOptions_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AIServiceServer).RankJoinOptions(ctx, req.(*RankRoutesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AIService_RankFinalRoutes_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RankRoutesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AIServiceServer).RankFinalRoutes(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AIService_RankFinalRoutes_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AIServiceServer).RankFinalRoutes(ctx, req.(*RankRoutesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var AIService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "godnota.ai.v1.AIService",
	HandlerType: (*AIServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GenerateOrUpdateConfig",
			Handler:    _AIService_GenerateOrUpdateConfig_Handler,
		},
		{
			MethodName: "UpdateConfig",
			Handler:    _AIService_UpdateConfig_Handler,
		},
		{
			MethodName: "RefineConfig",
			Handler:    _AIService_RefineConfig_Handler,
		},
		{
			MethodName: "GenerateRecommendations",
			Handler:    _AIService_GenerateRecommendations_Handler,
		},
		{
			MethodName: "RankRoutes",
			Handler:    _AIService_RankRoutes_Handler,
		},
		{
			MethodName: "RankJoinOptions",
			Handler:    _AIService_RankJoinOptions_Handler,
		},
		{
			MethodName: "RankFinalRoutes",
			Handler:    _AIService_RankFinalRoutes_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/ai/v1/ai.proto",
}
