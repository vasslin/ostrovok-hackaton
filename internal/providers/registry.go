package providers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

var envName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
var headerName = regexp.MustCompile("^[!#$%&'*+.^_`|~0-9A-Za-z-]+$")
var queryName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.-]*$`)

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
	if len(reg) == 0 {
		return nil, fmt.Errorf("registry %s is empty", path)
	}
	for name, provider := range reg {
		if err := validateProvider(name, provider); err != nil {
			return nil, fmt.Errorf("registry %s: %w", path, err)
		}
	}
	return reg, nil
}

func validateProvider(name string, p Provider) error {
	if name == "" {
		return fmt.Errorf("provider name is empty")
	}
	if strings.TrimSpace(p.BaseURL) != p.BaseURL {
		return fmt.Errorf("%s: base_url has surrounding whitespace", name)
	}
	base, err := url.Parse(p.BaseURL)
	if err != nil || base == nil || (base.Scheme != "https" && base.Scheme != "http") || base.Hostname() == "" || base.User != nil || base.Opaque != "" || base.RawQuery != "" || base.Fragment != "" {
		return fmt.Errorf("%s: base_url must be an HTTP(S) URL without credentials, query or fragment", name)
	}
	if p.TimeoutMS <= 0 || p.RPS <= 0 || p.CacheTTLS < 0 {
		return fmt.Errorf("%s: timeout_ms and rps must be positive; cache_ttl_s cannot be negative", name)
	}
	if len(p.Endpoints) == 0 {
		return fmt.Errorf("%s: endpoints are empty", name)
	}

	a := p.Auth
	switch a.Type {
	case "none":
		if a.Name != "" || a.Env != "" || a.EnvUser != "" || a.EnvPass != "" {
			return fmt.Errorf("%s: none auth cannot include credentials", name)
		}
	case "header":
		if !headerName.MatchString(a.Name) || !envName.MatchString(a.Env) || a.EnvUser != "" || a.EnvPass != "" {
			return fmt.Errorf("%s: invalid header auth", name)
		}
	case "query":
		if !queryName.MatchString(a.Name) || !envName.MatchString(a.Env) || a.EnvUser != "" || a.EnvPass != "" {
			return fmt.Errorf("%s: invalid query auth", name)
		}
	case "basic":
		if a.Name != "" || a.Env != "" || !envName.MatchString(a.EnvUser) || !envName.MatchString(a.EnvPass) {
			return fmt.Errorf("%s: invalid basic auth", name)
		}
	default:
		return fmt.Errorf("%s: unknown auth type %q", name, a.Type)
	}
	if base.Scheme != "https" && a.Type != "none" {
		return fmt.Errorf("%s: credentials require HTTPS", name)
	}

	for endpointName, endpoint := range p.Endpoints {
		if endpointName == "" {
			return fmt.Errorf("%s: endpoint name is empty", name)
		}
		if endpoint.Method != http.MethodGet && endpoint.Method != http.MethodPost {
			return fmt.Errorf("%s.%s: unsupported method %q", name, endpointName, endpoint.Method)
		}
		if !strings.HasPrefix(endpoint.Path, "/") || strings.Contains(endpoint.Path, "//") || strings.ContainsAny(endpoint.Path, "?#%\\\r\n") {
			return fmt.Errorf("%s.%s: invalid path %q", name, endpointName, endpoint.Path)
		}
		for _, segment := range strings.Split(endpoint.Path, "/") {
			if segment == "." || segment == ".." {
				return fmt.Errorf("%s.%s: invalid path %q", name, endpointName, endpoint.Path)
			}
		}
		if (endpoint.Method == http.MethodGet && endpoint.ParamsIn != "query") || (endpoint.Method == http.MethodPost && endpoint.ParamsIn != "json") {
			return fmt.Errorf("%s.%s: incompatible method and params_in", name, endpointName)
		}
	}
	return nil
}
