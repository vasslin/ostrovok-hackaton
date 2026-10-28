package devtool

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestModelConfigurationWorksOutsideDocker(t *testing.T) {
	t.Setenv("LLM_PROVIDER", "ollama")
	t.Setenv("LLM_BASE_URL", "http://host.docker.internal:11434/v1")
	t.Setenv("LLM_MODEL", "travel-qwen:4b")
	config := map[string]string{}
	nativeModel(config)
	if config["LLM_BASE_URL"] != "http://127.0.0.1:11434/v1" {
		t.Fatal("native service still needs Docker host DNS")
	}
	t.Setenv("LLM_BASE_URL", "http://192.168.1.20:11434/v1")
	nativeModel(config)
	if config["LLM_BASE_URL"] != "http://192.168.1.20:11434/v1" {
		t.Fatal("remote model host overwritten")
	}
}
func TestDoctorDetectsMissingModelWithoutSendingKeys(t *testing.T) {
	installed := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Error("local key leaked")
		}
		w.Header().Set("Content-Type", "application/json")
		if installed {
			_, _ = w.Write([]byte(`{"data":[{"id":"travel-qwen:4b"}]}`))
		} else {
			_, _ = w.Write([]byte(`{"data":[]}`))
		}
	}))
	defer server.Close()
	t.Setenv("LLM_PROVIDER", "ollama")
	t.Setenv("LLM_MODEL", "travel-qwen:4b")
	t.Setenv("LLM_BASE_URL", server.URL+"/v1")
	t.Setenv("LLM_API_KEY", "must-not-be-sent")
	var out bytes.Buffer
	if err := Doctor(context.Background(), "", &out); err != nil {
		t.Fatal(err)
	}
	installed = false
	if err := Doctor(context.Background(), "", &out); err == nil {
		t.Fatal("missing model accepted")
	}
}
