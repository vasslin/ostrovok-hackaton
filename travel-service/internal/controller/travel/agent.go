package travel

import (
	"context"
	"encoding/json"
	"strings"
	"time"
	"unicode/utf8"

	aiv1 "github.com/godnota/pkg/generated/ai/v1"
	commonv1 "github.com/godnota/pkg/generated/common/v1"
	sessionv1 "github.com/godnota/pkg/generated/session/v1"
	pb "github.com/godnota/pkg/generated/travel/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type agentTurn struct {
	State     string               `json:"state,omitempty"`
	Kind      string               `json:"kind"`
	Response  json.RawMessage      `json:"response"`
	Questions []*commonv1.Question `json:"questions"`
	Ideas     []*aiv1.TravelIdea   `json:"ideas"`
}
type agentResult struct {
	Status        string
	Config        *commonv1.TripConfig
	Questions     []*commonv1.Question
	Advice        []*aiv1.TravelIdea
	Summary       string
	MissingFields []string
	Warnings      []string
}

func (c *PlanningController) WithAgent(ai aiv1.AIServiceClient, key string) *PlanningController {
	c.ai = ai
	c.aiKey = key
	return c
}
func (c *PlanningController) agentContext(ctx context.Context) context.Context {
	if c.aiKey != "" {
		md, _ := metadata.FromOutgoingContext(ctx)
		md = md.Copy()
		md.Set("authorization", "Bearer "+c.aiKey)
		return metadata.NewOutgoingContext(ctx, md)
	}
	return ctx
}
func tail(value string, limit int) string {
	runes := []rune(value)
	if len(runes) > limit {
		return string(runes[len(runes)-limit:])
	}
	return value
}
func conversation(agent *sessionv1.GetAgentContextResponse) *aiv1.ConversationContext {
	out := &aiv1.ConversationContext{Summary: tail(agent.GetChatSummary(), 3000)}
	history := agent.GetChatHistory()
	total := 0
	for i := len(history) - 1; i >= 0 && len(out.Messages) < 20; i-- {
		m := history[i]
		if m.GetRole() != "user" && m.GetRole() != "assistant" {
			continue
		}
		raw := m.GetText()
		if m.GetRole() == "assistant" {
			raw = modelHistory(raw)
		}
		text := tail(raw, 6000)
		size := utf8.RuneCountInString(text)
		if total+size > 12000 {
			break
		}
		out.Messages = append([]*aiv1.Message{{Role: m.GetRole(), Content: text}}, out.Messages...)
		total += size
	}
	return out
}
func lastTurn(agent *sessionv1.GetAgentContextResponse) agentTurn {
	history := agent.GetChatHistory()
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].GetRole() != "assistant" {
			continue
		}
		var turn agentTurn
		if json.Unmarshal([]byte(history[i].GetText()), &turn) == nil && turn.Kind == "planning_turn" {
			return turn
		}
	}
	return agentTurn{}
}
func (c *PlanningController) record(ctx context.Context, userID, id, userText string, cfg *commonv1.TripConfig, summary *string, response proto.Message, questions []*commonv1.Question, ideas []*aiv1.TravelIdea, planning bool) error {
	data, err := (protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: true}).Marshal(response)
	if err != nil {
		return status.Error(codes.Internal, "response encoding failed")
	}
	kind := "travel_response"
	if planning {
		kind = "planning_turn"
	}
	state := ""
	switch response.(type) {
	case *pb.UpdateConfigResponse:
		state = "config_updated"
	case *pb.ConfirmConfigResponse:
		state = "confirmed"
	}
	stored, err := json.Marshal(agentTurn{State: state, Kind: kind, Response: data, Questions: questions, Ideas: ideas})
	if err != nil {
		return status.Error(codes.Internal, "context encoding failed")
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	messages := []*sessionv1.DialogMessage{}
	if strings.TrimSpace(userText) != "" {
		messages = append(messages, &sessionv1.DialogMessage{Role: "user", Text: userText, CreatedAt: now})
	}
	messages = append(messages, &sessionv1.DialogMessage{Role: "assistant", Text: string(stored), CreatedAt: now})
	_, err = c.sessions.SaveContext(ctx, &sessionv1.SaveContextRequest{UserId: userID, ChatId: id, Messages: messages, Cfg: cfg, Summary: summary})
	return err
}
func (c *PlanningController) finishAgent(ctx context.Context, loaded *sessionv1.GetAgentContextResponse, result agentResult, text string, response proto.Message) error {
	if result.Config == nil || (result.Status != "generate" && result.Status != "questions" && result.Status != "advice") {
		return status.Error(codes.DataLoss, "AI returned invalid planning response")
	}
	if err := c.record(ctx, loaded.GetUserId(), loaded.GetChatId(), text, result.Config, &result.Summary, response, result.Questions, result.Advice, true); err != nil {
		return err
	}
	c.mu.Lock()
	delete(c.cache, loaded.GetChatId())
	c.confirmed[loaded.GetChatId()] = false
	c.mu.Unlock()
	return nil
}
func (c *PlanningController) recommend(ctx context.Context, loaded *sessionv1.GetAgentContextResponse, result *agentResult, text string) error {
	if result.Status != "advice" {
		return nil
	}
	previous := []*aiv1.PreviousSession{}
	for _, chat := range loaded.GetPreviousChatsSummary() {
		if len(previous) == 5 {
			break
		}
		previous = append(previous, &aiv1.PreviousSession{Context: &aiv1.ConversationContext{Summary: tail(chat.GetSummary(), 3000)}})
	}
	ideas, err := c.ai.GenerateRecommendations(c.agentContext(ctx), &aiv1.GenerateRecommendationsRequest{Text: text, Config: modelConfig(result.Config), Context: conversation(loaded), LastSessions: previous, CurrentDate: time.Now().UTC().Format(time.DateOnly)})
	if err != nil {
		return err
	}
	if ideas == nil {
		return status.Error(codes.DataLoss, "AI returned no recommendations")
	}
	result.Advice = ideas.GetAdvice()
	return nil
}
func (c *PlanningController) PlanTrip(ctx context.Context, r *pb.PlanTripRequest) (*pb.PlanTripResponse, error) {
	unlock := c.lock(r.GetSessionId())
	defer unlock()
	loaded, err := c.load(ctx, r.GetUserId(), r.GetSessionId())
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(r.GetText()) == "" {
		return nil, status.Error(codes.InvalidArgument, "text required")
	}
	if c.ai == nil {
		return nil, status.Error(codes.Unavailable, "AI client not configured")
	}
	got, err := c.ai.GenerateOrUpdateConfig(c.agentContext(ctx), &aiv1.GenerateOrUpdateConfigRequest{SessionId: r.GetSessionId(), Text: r.GetText(), Config: modelConfig(loaded.GetCfg()), Context: conversation(loaded), CurrentDate: time.Now().UTC().Format(time.DateOnly)})
	if err != nil {
		return nil, err
	}
	if got == nil {
		return nil, status.Error(codes.DataLoss, "AI returned no response")
	}
	result := agentResult{Status: got.GetStatus(), Config: got.GetConfig(), Questions: got.GetQuestions(), Advice: got.GetAdvice(), Summary: got.GetSummary(), MissingFields: got.GetMissingFields(), Warnings: got.GetWarnings()}
	if err = c.recommend(ctx, loaded, &result, r.GetText()); err != nil {
		return nil, err
	}
	limitClarifications(loaded, &result)
	result.Config = retainResolved(loaded.GetCfg(), result.Config)
	out := &pb.PlanTripResponse{SessionId: r.GetSessionId(), Intent: result.Status, Config: result.Config, Questions: result.Questions, ReadyForSearch: result.Status == "generate", Advice: result.Advice, Summary: result.Summary, MissingFields: result.MissingFields, Warnings: result.Warnings}
	if err = c.finishAgent(ctx, loaded, result, r.GetText(), out); err != nil {
		return nil, err
	}
	return out, nil
}
func (c *PlanningController) SubmitAnswers(ctx context.Context, r *pb.SubmitAnswersRequest) (*pb.SubmitAnswersResponse, error) {
	unlock := c.lock(r.GetSessionId())
	defer unlock()
	loaded, err := c.load(ctx, r.GetUserId(), r.GetSessionId())
	if err != nil {
		return nil, err
	}
	if c.ai == nil {
		return nil, status.Error(codes.Unavailable, "AI client not configured")
	}
	turn := lastTurn(loaded)
	if len(r.GetAnswers()) == 0 || len(turn.Questions) == 0 {
		return nil, status.Error(codes.InvalidArgument, "answers require saved questions")
	}
	got, err := c.ai.UpdateConfig(c.agentContext(ctx), &aiv1.UpdateConfigRequest{SessionId: r.GetSessionId(), Config: modelConfig(loaded.GetCfg()), Context: conversation(loaded), Questions: turn.Questions, Answers: r.GetAnswers(), CurrentDate: time.Now().UTC().Format(time.DateOnly)})
	if err != nil {
		return nil, err
	}
	if got == nil {
		return nil, status.Error(codes.DataLoss, "AI returned no response")
	}
	result := agentResult{Status: got.GetStatus(), Config: got.GetConfig(), Questions: got.GetQuestions(), Advice: got.GetAdvice(), Summary: got.GetSummary(), MissingFields: got.GetMissingFields(), Warnings: got.GetWarnings()}
	textData, _ := json.Marshal(r.GetAnswers())
	text := string(textData)
	if err = c.recommend(ctx, loaded, &result, text); err != nil {
		return nil, err
	}
	limitClarifications(loaded, &result)
	result.Config = retainResolved(loaded.GetCfg(), result.Config)
	out := &pb.SubmitAnswersResponse{SessionId: r.GetSessionId(), Config: result.Config, Questions: result.Questions, ReadyForSearch: result.Status == "generate", Advice: result.Advice, Summary: result.Summary, MissingFields: result.MissingFields, Warnings: result.Warnings}
	if err = c.finishAgent(ctx, loaded, result, text, out); err != nil {
		return nil, err
	}
	return out, nil
}
func (c *PlanningController) SelectIdea(ctx context.Context, r *pb.SelectIdeaRequest) (*pb.SelectIdeaResponse, error) {
	unlock := c.lock(r.GetSessionId())
	defer unlock()
	loaded, err := c.load(ctx, r.GetUserId(), r.GetSessionId())
	if err != nil {
		return nil, err
	}
	if c.ai == nil {
		return nil, status.Error(codes.Unavailable, "AI client not configured")
	}
	var selected *aiv1.TravelIdea
	for _, idea := range lastTurn(loaded).Ideas {
		if idea.GetId() == r.GetIdeaId() && idea.GetId() != "" {
			selected = idea
			break
		}
	}
	if selected == nil {
		return nil, status.Error(codes.InvalidArgument, "idea_id must match a saved idea")
	}
	got, err := c.ai.RefineConfig(c.agentContext(ctx), &aiv1.RefineConfigRequest{SessionId: r.GetSessionId(), Text: r.GetComment(), SelectedIdea: selected, Config: modelConfig(loaded.GetCfg()), Context: conversation(loaded), CurrentDate: time.Now().UTC().Format(time.DateOnly)})
	if err != nil {
		return nil, err
	}
	if got == nil {
		return nil, status.Error(codes.DataLoss, "AI returned no response")
	}
	result := agentResult{Status: got.GetStatus(), Config: got.GetConfig(), Questions: got.GetQuestions(), Advice: got.GetAdvice(), Summary: got.GetSummary(), MissingFields: got.GetMissingFields(), Warnings: got.GetWarnings()}
	limitClarifications(loaded, &result)
	result.Config = retainResolved(loaded.GetCfg(), result.Config)
	out := &pb.SelectIdeaResponse{SessionId: r.GetSessionId(), Config: result.Config, Questions: result.Questions, Advice: result.Advice, Summary: result.Summary, MissingFields: result.MissingFields, Warnings: result.Warnings, ReadyForSearch: result.Status == "generate"}
	if err = c.finishAgent(ctx, loaded, result, "Выбрана идея "+r.GetIdeaId()+". "+r.GetComment(), out); err != nil {
		return nil, err
	}
	return out, nil
}

func modelHistory(text string) string {
	var value any
	if json.Unmarshal([]byte(text), &value) != nil {
		return text
	}
	var clean func(any)
	clean = func(v any) {
		switch item := v.(type) {
		case map[string]any:
			delete(item, "resolved")
			for _, child := range item {
				clean(child)
			}
		case []any:
			for _, child := range item {
				clean(child)
			}
		}
	}
	clean(value)
	data, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(data)
}
