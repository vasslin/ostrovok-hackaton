package httpapi

import (
	"os"
	"time"
)

func OperationTimeout(operation string) time.Duration {
	switch operation {
	case "/v1/travel/plan", "/v1/travel/answers", "/v1/travel/idea":
		timeout, err := time.ParseDuration(os.Getenv("LLM_TIMEOUT"))
		if err != nil || timeout < time.Second || timeout > 10*time.Minute {
			timeout = 180 * time.Second
		}
		return 2*timeout + 20*time.Second
	case "/v1/travel/search":
		return 30 * time.Second
	default:
		return backendTimeout
	}
}
