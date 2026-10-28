package travel

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	commonv1 "github.com/godnota/pkg/generated/common/v1"
	sessionv1 "github.com/godnota/pkg/generated/session/v1"
	travelv1 "github.com/godnota/pkg/generated/travel/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestLiveSessionFlow(t *testing.T) {
	address := os.Getenv("TRAVEL_INTEGRATION_ADDR")
	if address == "" {
		t.Skip("TRAVEL_INTEGRATION_ADDR is not set")
	}
	sessionAddress := os.Getenv("SESSION_INTEGRATION_ADDR")
	if sessionAddress == "" {
		sessionAddress = "localhost:50052"
	}
	sessionConn, err := grpc.NewClient(sessionAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer sessionConn.Close()
	travelConn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer travelConn.Close()
	sessions := sessionv1.NewSessionServiceClient(sessionConn)
	travel := travelv1.NewTravelServiceClient(travelConn)
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
	defer cancel()
	user := fmt.Sprintf("integration-%d", time.Now().UnixNano())
	created, err := sessions.CreateSession(ctx, &sessionv1.CreateSessionRequest{UserId: user})
	if err != nil {
		t.Fatal(err)
	}
	id := created.GetSessionId()
	initial, err := travel.PlanTrip(ctx, &travelv1.PlanTripRequest{UserId: user, SessionId: id, Text: "Хочу из Москвы в Сочи. Даты и состав пока уточняем."})
	if err != nil {
		t.Fatal(err)
	}
	if len(initial.GetQuestions()) == 0 || initial.GetReadyForSearch() {
		t.Fatal("partial trip marked complete")
	}
	stored, err := sessions.GetAgentContext(ctx, &sessionv1.GetAgentContextRequest{UserId: user, ChatId: id})
	if err != nil {
		t.Fatal(err)
	}
	if stored.GetCfg() == nil || stored.GetChatSummary() == "" || len(stored.GetChatHistory()) != 2 {
		t.Fatal("first turn not persisted")
	}
	start := time.Now().UTC().AddDate(0, 0, 10).Format(time.DateOnly)
	end := time.Now().UTC().AddDate(0, 0, 17).Format(time.DateOnly)
	text := fmt.Sprintf("Поедем с %s по %s, двое взрослых, бюджет 80000 рублей на всю поездку.", start, end)
	answers := []*commonv1.Answer{}
	for _, q := range initial.GetQuestions() {
		answers = append(answers, &commonv1.Answer{QuestionId: q.GetId(), Text: text})
	}
	completed, err := travel.SubmitAnswers(ctx, &travelv1.SubmitAnswersRequest{UserId: user, SessionId: id, Answers: answers})
	if err != nil {
		t.Fatal(err)
	}
	if !completed.GetReadyForSearch() {
		t.Fatalf("trip did not become complete: %v", completed.GetMissingFields())
	}
	stored, err = sessions.GetAgentContext(ctx, &sessionv1.GetAgentContextRequest{UserId: user, ChatId: id})
	if err != nil {
		t.Fatal(err)
	}
	if len(stored.GetChatHistory()) != 4 || stored.GetCfg().GetRequest() == nil || stored.GetChatSummary() == "" {
		t.Fatal("answer turn not persisted")
	}
	got, err := travel.GetConfig(ctx, &travelv1.GetConfigRequest{UserId: user, SessionId: id})
	if err != nil || got.GetConfig().GetRequest() == nil {
		t.Fatal("read config", err)
	}
	stored, err = sessions.GetAgentContext(ctx, &sessionv1.GetAgentContextRequest{UserId: user, ChatId: id})
	if err != nil || len(stored.GetChatHistory()) != 5 {
		t.Fatal("read response was not appended", err)
	}
	t.Log("real travel -> session -> Qwen flow persisted both turns and GetConfig reply")
}
