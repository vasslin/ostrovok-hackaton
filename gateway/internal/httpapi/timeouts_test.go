package httpapi

import (
	"testing"
	"time"
)

func TestAIHasTimeForPlanningAndRecommendations(t *testing.T) {
	t.Setenv("LLM_TIMEOUT", "180s")
	for _, operation := range []string{"/v1/travel/plan", "/v1/travel/answers", "/v1/travel/idea"} {
		if OperationTimeout(operation) < 6*time.Minute {
			t.Fatalf("%s cancels real local inference too early", operation)
		}
	}
	if OperationTimeout("/v1/travel/search") < 30*time.Second || OperationTimeout("/v1/session/create") > 10*time.Second {
		t.Fatal("timeouts for non-AI requests changed")
	}
	t.Setenv("LLM_TIMEOUT", "invalid")
	if OperationTimeout("/v1/travel/plan") < 6*time.Minute {
		t.Fatal("invalid timeout has no fallback")
	}
}
