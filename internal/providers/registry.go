package providers

import (
	"encoding/json"
	"fmt"
	"os"
)

// Registry is configs/providers.json: static settings per provider. Secrets are
// never stored here, only the names of the env vars that hold them.
type Registry map[string]Provider

type Provider struct {
	Description string              `json:"description"`
	BaseURL     string              `json:"base_url"`
	Auth        Auth                `json:"auth"`
	Headers     map[string]string   `json:"headers,omitempty"`
	RPS         float64             `json:"rps"`
	TimeoutMS   int                 `json:"timeout_ms"`
	CacheTTLS   int                 `json:"cache_ttl_s"`
	Attribution string              `json:"attribution,omitempty"`
	Endpoints   map[string]Endpoint `json:"endpoints"`
}

// Auth says where the credential goes.
//   - none
//   - header: Name is the header, Env holds the value
//   - query:  Name is the query parameter, Env holds the value
//   - basic:  EnvUser and EnvPass hold the HTTP Basic credentials
type Auth struct {
	Type    string `json:"type"`
	Name    string `json:"name,omitempty"`
	Env     string `json:"env,omitempty"`
	EnvUser string `json:"env_user,omitempty"`
	EnvPass string `json:"env_pass,omitempty"`
}

// Endpoint path may contain {placeholders}, filled from request fields tagged path:"name".
type Endpoint struct {
	Method   string `json:"method"` // GET or POST
	Path     string `json:"path"`
	ParamsIn string `json:"params_in"` // query or json
}

func LoadRegistry(path string) (Registry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var reg Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return reg, nil
}
