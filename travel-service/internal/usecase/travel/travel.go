package travel

import (
	"context"
	"log/slog"

	"github.com/godnota/travel-service/internal/entity"
)

type Repository interface {
	Save(ctx context.Context, plan entity.Plan) error
}

type SessionPort interface {
	GetContext(ctx context.Context, sessionID string) error
	SaveContext(ctx context.Context, sessionID string) error
}

type AIPort interface {
	GenerateOrUpdateConfig(ctx context.Context) error
	UpdateConfig(ctx context.Context) error
	RefineConfig(ctx context.Context) error
	RankRoutes(ctx context.Context) error
}

type MapPort interface {
	MapToProviderConfigs(ctx context.Context) error
	MapRoutesToFrontend(ctx context.Context) error
}

type RoutePort interface {
	SearchRoutes(ctx context.Context) error
}

type Service struct {
	repo     Repository
	sessions SessionPort
	ai       AIPort
	maps     MapPort
	routes   RoutePort
}

func New(repo Repository, sessions SessionPort, ai AIPort, maps MapPort, routes RoutePort) *Service {
	return &Service{repo: repo, sessions: sessions, ai: ai, maps: maps, routes: routes}
}

type Result struct{}

type PlanTripRequest struct {
	SessionID string
	Text      string
}

type Answer struct {
	QuestionID string
	Text       string
}

type SubmitAnswersRequest struct {
	SessionID string
	Answers   []Answer
}

type SelectIdeaRequest struct {
	SessionID string
	IdeaID    string
	Comment   string
}

type ConfigRequest struct {
	SessionID string
	Version   string
	Origin    string
}

type SessionRequest struct {
	SessionID string
}

func (s *Service) PlanTrip(ctx context.Context, req PlanTripRequest) (Result, error) {
	slog.Info("travel-service: plan trip", "session_id", req.SessionID, "text", req.Text)
	_ = s.repo.Save(ctx, entity.Plan{SessionID: req.SessionID, Text: req.Text})
	s.call(ctx, "session.GetContext", func(ctx context.Context) error { return s.sessions.GetContext(ctx, req.SessionID) })
	s.call(ctx, "ai.GenerateOrUpdateConfig", s.ai.GenerateOrUpdateConfig)
	s.call(ctx, "session.SaveContext", func(ctx context.Context) error { return s.sessions.SaveContext(ctx, req.SessionID) })
	return Result{}, nil
}

func (s *Service) SubmitAnswers(ctx context.Context, req SubmitAnswersRequest) (Result, error) {
	slog.Info("travel-service: submit answers", "session_id", req.SessionID, "answers", req.Answers)
	_ = s.repo.Save(ctx, entity.Plan{SessionID: req.SessionID})
	s.call(ctx, "session.GetContext", func(ctx context.Context) error { return s.sessions.GetContext(ctx, req.SessionID) })
	s.call(ctx, "ai.UpdateConfig", s.ai.UpdateConfig)
	s.call(ctx, "session.SaveContext", func(ctx context.Context) error { return s.sessions.SaveContext(ctx, req.SessionID) })
	return Result{}, nil
}

func (s *Service) SelectIdea(ctx context.Context, req SelectIdeaRequest) (Result, error) {
	slog.Info("travel-service: select idea", "session_id", req.SessionID, "idea_id", req.IdeaID, "comment", req.Comment)
	_ = s.repo.Save(ctx, entity.Plan{SessionID: req.SessionID, IdeaID: req.IdeaID, Comment: req.Comment})
	s.call(ctx, "session.GetContext", func(ctx context.Context) error { return s.sessions.GetContext(ctx, req.SessionID) })
	s.call(ctx, "ai.RefineConfig", s.ai.RefineConfig)
	s.call(ctx, "session.SaveContext", func(ctx context.Context) error { return s.sessions.SaveContext(ctx, req.SessionID) })
	return Result{}, nil
}

func (s *Service) GetConfig(ctx context.Context, req SessionRequest) (Result, error) {
	slog.Info("travel-service: get config", "session_id", req.SessionID)
	s.call(ctx, "session.GetContext", func(ctx context.Context) error { return s.sessions.GetContext(ctx, req.SessionID) })
	return Result{}, nil
}

func (s *Service) UpdateConfig(ctx context.Context, req ConfigRequest) (Result, error) {
	slog.Info("travel-service: update config", "session_id", req.SessionID, "version", req.Version, "origin", req.Origin)
	_ = s.repo.Save(ctx, entity.Plan{SessionID: req.SessionID})
	s.call(ctx, "session.SaveContext", func(ctx context.Context) error { return s.sessions.SaveContext(ctx, req.SessionID) })
	return Result{}, nil
}

func (s *Service) ConfirmConfig(ctx context.Context, req ConfigRequest) (Result, error) {
	slog.Info("travel-service: confirm config", "session_id", req.SessionID, "version", req.Version, "origin", req.Origin)
	_ = s.repo.Save(ctx, entity.Plan{SessionID: req.SessionID})
	s.call(ctx, "session.SaveContext", func(ctx context.Context) error { return s.sessions.SaveContext(ctx, req.SessionID) })
	return Result{}, nil
}

func (s *Service) SearchRoutes(ctx context.Context, req SessionRequest) (Result, error) {
	slog.Info("travel-service: search routes", "session_id", req.SessionID)
	_ = s.repo.Save(ctx, entity.Plan{SessionID: req.SessionID})
	s.call(ctx, "map.MapToProviderConfigs", s.maps.MapToProviderConfigs)
	s.call(ctx, "route.SearchRoutes", s.routes.SearchRoutes)
	s.call(ctx, "ai.RankRoutes", s.ai.RankRoutes)
	s.call(ctx, "map.MapRoutesToFrontend", s.maps.MapRoutesToFrontend)
	s.call(ctx, "session.SaveContext", func(ctx context.Context) error { return s.sessions.SaveContext(ctx, req.SessionID) })
	return Result{}, nil
}

func (s *Service) GetRoutes(ctx context.Context, req SessionRequest) (Result, error) {
	slog.Info("travel-service: get routes", "session_id", req.SessionID)
	return Result{}, nil
}

func (s *Service) call(ctx context.Context, name string, fn func(context.Context) error) {
	if err := fn(ctx); err != nil {
		slog.Error("travel-service: downstream call failed", "call", name, "error", err)
	}
}
