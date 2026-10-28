package travel

import (
	"encoding/json"
	sessionv1 "github.com/godnota/pkg/generated/session/v1"
	"os"
	"strconv"
)

func limitClarifications(loaded *sessionv1.GetAgentContextResponse, result *agentResult) {
	limit := 3
	if n, err := strconv.Atoi(os.Getenv("MAX_CLARIFICATION_ROUNDS")); err == nil && n > 0 && n <= 10 {
		limit = n
	}
	rounds := 0
	for _, message := range loaded.GetChatHistory() {
		if message.GetRole() != "assistant" {
			continue
		}
		var turn agentTurn
		if json.Unmarshal([]byte(message.GetText()), &turn) == nil && turn.Kind == "planning_turn" && len(turn.Questions) > 0 {
			rounds++
		}
	}
	if rounds >= limit && len(result.Questions) > 0 {
		result.Questions = nil
		result.Status = "questions"
		result.Warnings = append(result.Warnings, "Лимит уточнений достигнут. Заполните недостающие параметры в карточке поездки; неизвестные значения не подставляются автоматически.")
	}
}
