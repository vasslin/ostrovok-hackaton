package travel

import (
	"context"

	commonv1 "github.com/godnota/pkg/generated/common/v1"
	pb "github.com/godnota/pkg/generated/travel/v1"
	usecase "github.com/godnota/travel-service/internal/usecase/travel"
)

type travelService interface {
	PlanTrip(context.Context, usecase.PlanTripRequest) (usecase.Result, error)
	SubmitAnswers(context.Context, usecase.SubmitAnswersRequest) (usecase.Result, error)
	SelectIdea(context.Context, usecase.SelectIdeaRequest) (usecase.Result, error)
	GetConfig(context.Context, usecase.SessionRequest) (usecase.Result, error)
	UpdateConfig(context.Context, usecase.ConfigRequest) (usecase.Result, error)
	ConfirmConfig(context.Context, usecase.ConfigRequest) (usecase.Result, error)
	SearchRoutes(context.Context, usecase.SessionRequest) (usecase.Result, error)
	GetRoutes(context.Context, usecase.SessionRequest) (usecase.Result, error)
}

type Controller struct {
	service travelService
}

func New(service travelService) *Controller { return &Controller{service: service} }

var _ pb.TravelServiceServer = (*Controller)(nil)

func (c *Controller) PlanTrip(ctx context.Context, req *pb.PlanTripRequest) (*pb.PlanTripResponse, error) {
	if req != nil {
		_, _ = c.service.PlanTrip(ctx, usecase.PlanTripRequest{
			SessionID: req.GetSessionId(),
			Text:      req.GetText(),
		})
	}
	return &pb.PlanTripResponse{}, nil
}

func (c *Controller) SubmitAnswers(ctx context.Context, req *pb.SubmitAnswersRequest) (*pb.SubmitAnswersResponse, error) {
	if req != nil {
		answers := make([]usecase.Answer, 0, len(req.GetAnswers()))
		for _, answer := range req.GetAnswers() {
			answers = append(answers, usecase.Answer{
				QuestionID: answer.GetQuestionId(),
				Text:       answer.GetText(),
			})
		}
		_, _ = c.service.SubmitAnswers(ctx, usecase.SubmitAnswersRequest{
			SessionID: req.GetSessionId(),
			Answers:   answers,
		})
	}
	return &pb.SubmitAnswersResponse{}, nil
}

func (c *Controller) SelectIdea(ctx context.Context, req *pb.SelectIdeaRequest) (*pb.SelectIdeaResponse, error) {
	if req != nil {
		_, _ = c.service.SelectIdea(ctx, usecase.SelectIdeaRequest{
			SessionID: req.GetSessionId(),
			IdeaID:    req.GetIdeaId(),
			Comment:   req.GetComment(),
		})
	}
	return &pb.SelectIdeaResponse{}, nil
}

func (c *Controller) GetConfig(ctx context.Context, req *pb.GetConfigRequest) (*pb.GetConfigResponse, error) {
	if req != nil {
		_, _ = c.service.GetConfig(ctx, usecase.SessionRequest{SessionID: req.GetSessionId()})
	}
	return &pb.GetConfigResponse{}, nil
}

func (c *Controller) UpdateConfig(ctx context.Context, req *pb.UpdateConfigRequest) (*pb.UpdateConfigResponse, error) {
	if req != nil {
		_, _ = c.service.UpdateConfig(ctx, configRequest(req.GetSessionId(), req.GetConfig()))
	}
	return &pb.UpdateConfigResponse{}, nil
}

func (c *Controller) ConfirmConfig(ctx context.Context, req *pb.ConfirmConfigRequest) (*pb.ConfirmConfigResponse, error) {
	if req != nil {
		_, _ = c.service.ConfirmConfig(ctx, configRequest(req.GetSessionId(), req.GetConfig()))
	}
	return &pb.ConfirmConfigResponse{}, nil
}

func (c *Controller) SearchRoutes(ctx context.Context, req *pb.SearchRoutesRequest) (*pb.SearchRoutesResponse, error) {
	if req != nil {
		_, _ = c.service.SearchRoutes(ctx, usecase.SessionRequest{SessionID: req.GetSessionId()})
	}
	return &pb.SearchRoutesResponse{}, nil
}

func (c *Controller) GetRoutes(ctx context.Context, req *pb.GetRoutesRequest) (*pb.GetRoutesResponse, error) {
	if req != nil {
		_, _ = c.service.GetRoutes(ctx, usecase.SessionRequest{SessionID: req.GetSessionId()})
	}
	return &pb.GetRoutesResponse{}, nil
}

func configRequest(sessionID string, config *commonv1.TripConfig) usecase.ConfigRequest {
	return usecase.ConfigRequest{
		SessionID: sessionID,
		Version:   config.GetVersion(),
		Origin:    config.GetTrip().GetOrigin().GetCity(),
	}
}
