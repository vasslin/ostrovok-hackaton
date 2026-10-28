package provider

import (
	"context"
	"os"

	"google.golang.org/protobuf/types/known/structpb"
)

func (c *Controller) GetCapabilities(context.Context, *structpb.Struct) (*structpb.Struct, error) {
	mode, status, name, environment := "live", "not_configured", "", ""
	if c.searcher != nil {
		mode = c.searcher.Mode
	}
	if mode == "demo" {
		status, name, environment = "available", "demo", "synthetic"
	} else if c.searcher != nil && c.searcher.Hotels != nil {
		status, name, environment = "available", "adapter", "live"
	} else if c.service != nil {
		if os.Getenv("RATEHAWK_MODE") == "production" || os.Getenv("RATEHAWK_MODE") == "sandbox" {
			name, environment = "ratehawk", os.Getenv("RATEHAWK_MODE")
			if os.Getenv("RATEHAWK_KEY_ID") != "" && os.Getenv("RATEHAWK_API_KEY") != "" {
				status = "available"
			}
		} else if os.Getenv("HOTELBEDS_MODE") == "evaluation" {
			name, environment = "hotelbeds", "evaluation"
			if os.Getenv("HOTELBEDS_API_KEY") != "" && os.Getenv("HOTELBEDS_SECRET") != "" {
				status = "available"
			}
		}
	}
	return structpb.NewStruct(map[string]any{"mode": mode, "hotels": map[string]any{"status": status, "provider": name, "environment": environment}, "transport": map[string]any{"aviasales": mode == "demo" || os.Getenv("TRAVELPAYOUTS_TOKEN") != "", "yandex_rasp": mode == "demo" || os.Getenv("YANDEX_RASP_API_KEY") != ""}})
}
