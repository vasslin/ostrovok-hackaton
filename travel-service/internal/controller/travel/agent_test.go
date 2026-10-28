package travel

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	aiv1 "github.com/godnota/pkg/generated/ai/v1"
	commonv1 "github.com/godnota/pkg/generated/common/v1"
	sessionv1 "github.com/godnota/pkg/generated/session/v1"
	pb "github.com/godnota/pkg/generated/travel/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type durableSessions struct {
	sessionv1.SessionServiceClient
	state              *sessionv1.GetAgentContextResponse
	reads, saves       int
	failRead, failSave error
}

func (f *durableSessions) GetAgentContext(_ context.Context, r *sessionv1.GetAgentContextRequest, _ ...grpc.CallOption) (*sessionv1.GetAgentContextResponse, error) {
	f.reads++
	if f.failRead != nil {
		return nil, f.failRead
	}
	if r.GetUserId() != f.state.GetUserId() {
		return nil, status.Error(codes.PermissionDenied, "owner mismatch")
	}
	return proto.Clone(f.state).(*sessionv1.GetAgentContextResponse), nil
}
func (f *durableSessions) SaveContext(_ context.Context, r *sessionv1.SaveContextRequest, _ ...grpc.CallOption) (*sessionv1.SaveContextResponse, error) {
	if f.failSave != nil {
		return nil, f.failSave
	}
	f.saves++
	if len(r.GetMessages()) > 2 {
		panic("full history was appended again")
	}
	f.state.ChatHistory = append(f.state.ChatHistory, r.GetMessages()...)
	if r.Cfg != nil {
		f.state.Cfg = proto.Clone(r.Cfg).(*commonv1.TripConfig)
	}
	if r.Summary != nil {
		f.state.ChatSummary = r.GetSummary()
	}
	return &sessionv1.SaveContextResponse{UserId: r.GetUserId(), ChatId: r.GetChatId()}, nil
}

type agentFake struct {
	aiv1.AIServiceClient
	generate  func(context.Context, *aiv1.GenerateOrUpdateConfigRequest) (*aiv1.GenerateOrUpdateConfigResponse, error)
	update    func(context.Context, *aiv1.UpdateConfigRequest) (*aiv1.UpdateConfigResponse, error)
	refine    func(context.Context, *aiv1.RefineConfigRequest) (*aiv1.RefineConfigResponse, error)
	recommend func(context.Context, *aiv1.GenerateRecommendationsRequest) (*aiv1.GenerateRecommendationsResponse, error)
}

func (f agentFake) GenerateOrUpdateConfig(ctx context.Context, r *aiv1.GenerateOrUpdateConfigRequest, _ ...grpc.CallOption) (*aiv1.GenerateOrUpdateConfigResponse, error) {
	return f.generate(ctx, r)
}
func (f agentFake) UpdateConfig(ctx context.Context, r *aiv1.UpdateConfigRequest, _ ...grpc.CallOption) (*aiv1.UpdateConfigResponse, error) {
	return f.update(ctx, r)
}
func (f agentFake) RefineConfig(ctx context.Context, r *aiv1.RefineConfigRequest, _ ...grpc.CallOption) (*aiv1.RefineConfigResponse, error) {
	return f.refine(ctx, r)
}
func (f agentFake) GenerateRecommendations(ctx context.Context, r *aiv1.GenerateRecommendationsRequest, _ ...grpc.CallOption) (*aiv1.GenerateRecommendationsResponse, error) {
	return f.recommend(ctx, r)
}
func durableState() *durableSessions {
	return &durableSessions{state: &sessionv1.GetAgentContextResponse{UserId: "owner", ChatId: "chat", ChatSummary: "old summary", Cfg: &commonv1.TripConfig{Version: "1.0"}}}
}
func TestAgentLoadsAndAppendsAcrossControllerRestart(t *testing.T) {
	sessions := durableState()
	question := &commonv1.Question{Id: "adults", Field: "/travelers/adults", Text: "Сколько взрослых?", Required: true}
	ai := agentFake{generate: func(ctx context.Context, r *aiv1.GenerateOrUpdateConfigRequest) (*aiv1.GenerateOrUpdateConfigResponse, error) {
		if sessions.reads != 1 || r.GetConfig().GetVersion() != "1.0" || r.GetText() != "Хочу в Сочи" || r.GetContext().GetSummary() != "old summary" {
			t.Fatal("loaded state not forwarded")
		}
		md, _ := metadata.FromOutgoingContext(ctx)
		if got := md.Get("authorization"); len(got) != 1 || got[0] != "Bearer secret" {
			t.Fatal("AI authorization missing")
		}
		return &aiv1.GenerateOrUpdateConfigResponse{Status: "questions", Config: r.GetConfig(), Questions: []*commonv1.Question{question}, Summary: "new summary"}, nil
	}, update: func(_ context.Context, r *aiv1.UpdateConfigRequest) (*aiv1.UpdateConfigResponse, error) {
		if sessions.reads != 2 || len(r.GetQuestions()) != 1 || r.GetQuestions()[0].GetId() != "adults" || r.GetAnswers()[0].GetText() != "Двое" || len(r.GetContext().GetMessages()) != 2 || r.GetContext().GetSummary() != "new summary" {
			t.Fatal("saved questions/history not recovered")
		}
		return &aiv1.UpdateConfigResponse{Status: "generate", Config: r.GetConfig(), Summary: "updated after answers"}, nil
	}}
	controller := New(nil).WithPlanning(sessions, nil).WithAgent(ai, "secret")
	out, err := controller.PlanTrip(context.Background(), &pb.PlanTripRequest{SessionId: "chat", UserId: "owner", Text: "Хочу в Сочи"})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Questions) != 1 || sessions.saves != 1 || len(sessions.state.ChatHistory) != 2 || sessions.state.ChatSummary != "new summary" {
		t.Fatal("turn was not persisted")
	}
	restarted := New(nil).WithPlanning(sessions, nil).WithAgent(ai, "secret")
	answered, err := restarted.SubmitAnswers(context.Background(), &pb.SubmitAnswersRequest{SessionId: "chat", UserId: "owner", Answers: []*commonv1.Answer{{QuestionId: "adults", Text: "Двое"}}})
	if err != nil {
		t.Fatal(err)
	}
	if !answered.GetReadyForSearch() || sessions.saves != 2 || len(sessions.state.ChatHistory) != 4 || sessions.state.ChatSummary != "updated after answers" {
		t.Fatal("answer turn lost or history duplicated")
	}
	if _, err = restarted.GetConfig(context.Background(), &pb.GetConfigRequest{SessionId: "chat", UserId: "owner"}); err != nil {
		t.Fatal(err)
	}
	if sessions.saves != 3 || sessions.reads != 3 || len(sessions.state.ChatHistory) != 5 || sessions.state.ChatSummary != "updated after answers" {
		t.Fatal("read response did not update history or cleared context")
	}
}
func TestIdeasPersistAndSelectionUsesSavedID(t *testing.T) {
	sessions := durableState()
	sessions.state.PreviousChatsSummary = []*sessionv1.PreviousChatSummary{{ChatId: "old", Summary: "Любит музеи"}}
	idea := &aiv1.TravelIdea{Id: "kazan", City: "Казань", Reason: "Идея", Config: &commonv1.TripConfig{Version: "1.0"}}
	ai := agentFake{generate: func(_ context.Context, r *aiv1.GenerateOrUpdateConfigRequest) (*aiv1.GenerateOrUpdateConfigResponse, error) {
		return &aiv1.GenerateOrUpdateConfigResponse{Status: "advice", Config: r.GetConfig(), Summary: "wants ideas"}, nil
	}, recommend: func(_ context.Context, r *aiv1.GenerateRecommendationsRequest) (*aiv1.GenerateRecommendationsResponse, error) {
		if len(r.GetLastSessions()) != 1 || r.GetLastSessions()[0].GetContext().GetSummary() != "Любит музеи" {
			t.Fatal("previous summaries dropped")
		}
		return &aiv1.GenerateRecommendationsResponse{Advice: []*aiv1.TravelIdea{idea}}, nil
	}, refine: func(_ context.Context, r *aiv1.RefineConfigRequest) (*aiv1.RefineConfigResponse, error) {
		if !proto.Equal(r.GetSelectedIdea(), idea) || r.GetConfig() == nil {
			t.Fatal("wrong selection")
		}
		return &aiv1.RefineConfigResponse{Status: "questions", Config: r.GetConfig(), Summary: "selected Kazan"}, nil
	}}
	c := New(nil).WithPlanning(sessions, nil).WithAgent(ai, "")
	out, err := c.PlanTrip(context.Background(), &pb.PlanTripRequest{UserId: "owner", SessionId: "chat", Text: "Предложи поездку"})
	if err != nil || len(out.GetAdvice()) != 1 {
		t.Fatal(err)
	}
	c = New(nil).WithPlanning(sessions, nil).WithAgent(ai, "")
	if _, err = c.SelectIdea(context.Background(), &pb.SelectIdeaRequest{UserId: "owner", SessionId: "chat", IdeaId: "invented"}); status.Code(err) != codes.InvalidArgument {
		t.Fatal("invented idea accepted", err)
	}
	if _, err = c.SelectIdea(context.Background(), &pb.SelectIdeaRequest{UserId: "owner", SessionId: "chat", IdeaId: "kazan"}); err != nil {
		t.Fatal(err)
	}
	if len(sessions.state.ChatHistory) != 4 {
		t.Fatal("selection not persisted")
	}
}
func TestStorageAndAIFailuresPropagate(t *testing.T) {
	for _, mode := range []string{"read", "owner", "ai", "save"} {
		t.Run(mode, func(t *testing.T) {
			sessions := durableState()
			calls := 0
			down := status.Error(codes.Unavailable, "down")
			ai := agentFake{generate: func(_ context.Context, r *aiv1.GenerateOrUpdateConfigRequest) (*aiv1.GenerateOrUpdateConfigResponse, error) {
				calls++
				if mode == "ai" {
					return nil, down
				}
				return &aiv1.GenerateOrUpdateConfigResponse{Status: "questions", Config: r.GetConfig(), Summary: "next"}, nil
			}}
			user := "owner"
			if mode == "owner" {
				user = "other"
			}
			if mode == "read" {
				sessions.failRead = down
			}
			if mode == "save" {
				sessions.failSave = down
			}
			c := New(nil).WithPlanning(sessions, nil).WithAgent(ai, "")
			out, err := c.PlanTrip(context.Background(), &pb.PlanTripRequest{SessionId: "chat", UserId: user, Text: "Поездка"})
			if out != nil || err == nil || sessions.saves != 0 {
				t.Fatal("failure became success")
			}
			if (mode == "read" || mode == "owner") && calls != 0 {
				t.Fatal("AI called without loaded context")
			}
			if mode != "owner" && !errors.Is(err, down) {
				t.Fatal("downstream error lost", err)
			}
		})
	}
}
func TestHistoryMessageEnvelope(t *testing.T) {
	turn := agentTurn{Kind: "planning_turn", Questions: []*commonv1.Question{{Id: "q"}}, Ideas: []*aiv1.TravelIdea{{Id: "i"}}}
	data, _ := json.Marshal(turn)
	loaded := &sessionv1.GetAgentContextResponse{ChatHistory: []*sessionv1.DialogMessage{{Role: "assistant", Text: string(data)}, {Role: "assistant", Text: `{"kind":"travel_response"}`}}}
	if lastTurn(loaded).Questions[0].GetId() != "q" {
		t.Fatal("nonplanning read erased pending questions")
	}
}

func TestBackendResolvedDoesNotEnterModelHistory(t *testing.T) {
	loaded := &sessionv1.GetAgentContextResponse{ChatHistory: []*sessionv1.DialogMessage{{Role: "assistant", Text: `{"kind":"travel_response","response":{"config":{"request":{"origin":{"city":"Москва"}},"resolved":{"private_code":"secret-code"}}}}`}}}
	got := conversation(loaded).GetMessages()[0].GetContent()
	if strings.Contains(got, "resolved") || strings.Contains(got, "secret-code") || !strings.Contains(got, "Москва") {
		t.Fatal("backend resolver leaked through assistant history")
	}
}
