package grpcapi

import (
	aiv1 "ai-service/api/ai/v1"
	"ai-service/internal/service"
	"context"

	"google.golang.org/protobuf/proto"
)

type planningServer struct {
	aiv1.UnimplementedAIServiceServer
	service *service.Service
}

func (s *planningServer) analyze(ctx context.Context, req, out proto.Message, refine bool) error {
	var in service.PlanningInput
	if err := input(req, &in); err != nil {
		return err
	}
	if refine {
		in.ForceGenerate = true
	}
	result, err := s.service.AnalyzeCanonical(ctx, in)
	if err != nil {
		return convertError(ctx, err)
	}
	return output(result, out)
}
func (s *planningServer) GenerateOrUpdateConfig(ctx context.Context, req *aiv1.GenerateOrUpdateConfigRequest) (*aiv1.GenerateOrUpdateConfigResponse, error) {
	out := new(aiv1.GenerateOrUpdateConfigResponse)
	if err := s.analyze(ctx, req, out, false); err != nil {
		return nil, err
	}
	return out, nil
}
func (s *planningServer) UpdateConfig(ctx context.Context, req *aiv1.UpdateConfigRequest) (*aiv1.UpdateConfigResponse, error) {
	out := new(aiv1.UpdateConfigResponse)
	if err := s.analyze(ctx, req, out, false); err != nil {
		return nil, err
	}
	return out, nil
}
func (s *planningServer) RefineConfig(ctx context.Context, req *aiv1.RefineConfigRequest) (*aiv1.RefineConfigResponse, error) {
	out := new(aiv1.RefineConfigResponse)
	if err := s.analyze(ctx, req, out, true); err != nil {
		return nil, err
	}
	return out, nil
}
func (s *planningServer) GenerateRecommendations(ctx context.Context, req *aiv1.GenerateRecommendationsRequest) (*aiv1.GenerateRecommendationsResponse, error) {
	var in service.RecommendationsInput
	if err := input(req, &in); err != nil {
		return nil, err
	}
	result, err := s.service.RecommendationsCanonical(ctx, in)
	if err != nil {
		return nil, convertError(ctx, err)
	}
	out := new(aiv1.GenerateRecommendationsResponse)
	if err := output(result, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *planningServer) RankRoutes(ctx context.Context, req *aiv1.RankRoutesRequest) (*aiv1.RankRoutesResponse, error) {
	var in service.RankRoutesInput
	if err := input(req, &in); err != nil {
		return nil, err
	}
	result, err := s.service.RankRoutes(ctx, in)
	if err != nil {
		return nil, convertError(ctx, err)
	}
	out := new(aiv1.RankRoutesResponse)
	if err := output(result, out); err != nil {
		return nil, err
	}
	return out, nil
}
func (s *planningServer) RankJoinOptions(ctx context.Context, req *aiv1.RankRoutesRequest) (*aiv1.RankRoutesResponse, error) {
	return s.RankRoutes(ctx, req)
}
func (s *planningServer) RankFinalRoutes(ctx context.Context, req *aiv1.RankRoutesRequest) (*aiv1.RankRoutesResponse, error) {
	return s.RankRoutes(ctx, req)
}
