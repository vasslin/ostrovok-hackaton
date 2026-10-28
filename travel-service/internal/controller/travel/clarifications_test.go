package travel

import (
	"encoding/json"
	commonv1 "github.com/godnota/pkg/generated/common/v1"
	sessionv1 "github.com/godnota/pkg/generated/session/v1"
	"testing"
)

func TestClarificationLimitDoesNotInventReadyConfig(t *testing.T) {
	t.Setenv("MAX_CLARIFICATION_ROUNDS", "2")
	raw, _ := json.Marshal(agentTurn{Kind: "planning_turn", Questions: []*commonv1.Question{{Id: "q", Text: "Дата?"}}})
	loaded := &sessionv1.GetAgentContextResponse{ChatHistory: []*sessionv1.DialogMessage{{Role: "assistant", Text: string(raw)}, {Role: "assistant", Text: string(raw)}}}
	result := agentResult{Status: "questions", Questions: []*commonv1.Question{{Id: "next"}}}
	limitClarifications(loaded, &result)
	if result.Status != "questions" || len(result.Questions) != 0 || len(result.Warnings) != 1 || result.Config != nil {
		t.Fatal("limit forced fabricated or ready config")
	}
}
