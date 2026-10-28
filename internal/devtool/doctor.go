package devtool

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func nativeModel(config map[string]string) {
	for _, key := range []string{"LLM_PROVIDER", "LLM_BASE_URL", "LLM_MODEL", "LLM_TIMEOUT", "LLM_API_KEY", "SERVICE_API_KEY"} {
		if value, ok := os.LookupEnv(key); ok {
			config[key] = value
		}
	}
	if config["LLM_PROVIDER"] == "" {
		config["LLM_PROVIDER"] = "ollama"
	}
	if config["LLM_PROVIDER"] != "ollama" {
		return
	}
	if config["LLM_MODEL"] == "" {
		config["LLM_MODEL"] = "travel-qwen:4b"
	}
	if config["LLM_BASE_URL"] == "" {
		config["LLM_BASE_URL"] = "http://localhost:11434/v1"
	}
	parsed, err := url.Parse(config["LLM_BASE_URL"])

	if err == nil && parsed.Hostname() == "host.docker.internal" {
		host := strings.Replace(parsed.Host, "host.docker.internal", "127.0.0.1", 1)
		parsed.Host = host
		config["LLM_BASE_URL"] = parsed.String()
	}
}

func Doctor(ctx context.Context, envFile string, out io.Writer) error {
	config, err := ReadEnv(envFile)
	if err != nil {
		return err
	}
	nativeModel(config)
	if config["LLM_PROVIDER"] != "ollama" {
		return fmt.Errorf("doctor checks free Ollama mode; set LLM_PROVIDER=ollama")
	}
	base := strings.TrimSuffix(config["LLM_BASE_URL"], "/")
	req, err := http.NewRequestWithContext(ctx, "GET", base+"/models", nil)
	if err != nil {
		return fmt.Errorf("invalid LLM_BASE_URL")
	}
	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Ollama is unreachable: start ollama serve; native launches use localhost:11434, Docker uses host.docker.internal:11434")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama model list returned HTTP %d", response.StatusCode)
	}
	var listed struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err = json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&listed); err != nil {
		return fmt.Errorf("Ollama returned invalid model list")
	}
	for _, model := range listed.Data {
		if model.ID == config["LLM_MODEL"] {
			fmt.Fprintf(out, "PASS: Ollama is reachable; model %s is installed; no LLM API key needed\n", model.ID)
			return nil
		}
	}
	return fmt.Errorf("model %s is missing: ollama pull qwen3:4b; ollama create travel-qwen:4b -f ai-service/ollama/Modelfile", config["LLM_MODEL"])
}
