package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/ostrovok-hackathon-2026/godnota/providerkit"
)

func TestHotelCapabilitiesExposeEnvironmentWithoutSecrets(t *testing.T) {
	t.Setenv("RATEHAWK_MODE", "")
	t.Setenv("HOTELBEDS_MODE", "")
	searcher, _ := providerkit.NewSearcher(nil, "live")
	c := New(&testService{}).WithSearcher(searcher)
	got, err := c.GetCapabilities(context.Background(), nil)
	if err != nil || got.AsMap()["hotels"].(map[string]any)["status"] != "not_configured" {
		t.Fatalf("%v %v", got, err)
	}
	t.Setenv("RATEHAWK_MODE", "sandbox")
	t.Setenv("RATEHAWK_KEY_ID", "private-id")
	t.Setenv("RATEHAWK_API_KEY", "private-token")
	got, err = c.GetCapabilities(context.Background(), nil)
	if err != nil || got.AsMap()["hotels"].(map[string]any)["environment"] != "sandbox" || got.AsMap()["hotels"].(map[string]any)["status"] != "available" {
		t.Fatalf("%v %v", got, err)
	}
	if strings.Contains(got.String(), "private-") {
		t.Fatal("secret leaked")
	}
	searcher.Mode = "demo"
	got, _ = c.GetCapabilities(context.Background(), nil)
	if got.AsMap()["hotels"].(map[string]any)["environment"] != "synthetic" {
		t.Fatal("demo presented as live")
	}
}
