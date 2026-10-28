// Command smoke sends every example request in configs/examples to the real
// provider APIs and saves the responses to testdata/<provider>/<endpoint>.<ext>.
// Providers whose key isn't set are skipped. Secrets are redacted from the
// output and from the saved files.
//
//	go run ./cmd/smoke                      # everything that has a key
//	go run ./cmd/smoke -only aviasales,osrm # just these providers (or provider.endpoint)
package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ostrovok-hackathon-2026/godnota/internal/providers"
)

func main() {
	configs := flag.String("configs", "configs", "configs directory")
	out := flag.String("out", "testdata", "where to save responses")
	only := flag.String("only", "", "comma-separated providers or provider.endpoint names")
	envFile := flag.String("env", ".env", "file with KEY=VALUE secrets")
	flag.Parse()

	if err := loadEnv(*envFile); err != nil && !os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "env:", err)
		os.Exit(1)
	}
	reg, err := providers.LoadRegistry(filepath.Join(*configs, "providers.json"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	secrets := secretValues(reg)

	failed := 0
	for _, s := range providers.Catalog {
		if !selected(*only, s) {
			continue
		}
		p := reg[s.Provider]
		if s.Provider == "nominatim" && s.Endpoint == "search_amenity" && *only == "" {
			fmt.Printf("SKIP %-38s public Nominatim does not allow systematic POI collection\n", s.Name())
			continue
		}
		if missing := missingKey(p.Auth); missing != "" {
			fmt.Printf("SKIP %-38s no %s\n", s.Name(), missing)
			continue
		}

		status, body, contentType, err := call(reg, *configs, s)
		if err != nil {
			fmt.Printf("ERR  %-38s %s\n", s.Name(), redact(err.Error(), secrets))
			failed++
			continue
		}
		body = []byte(redact(string(body), secrets))
		path := filepath.Join(*out, s.Provider, s.Endpoint+extension(contentType, body))
		if err := save(path, body); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		mark := "OK  "
		if status >= 300 {
			mark = "FAIL"
			failed++
		}
		fmt.Printf("%s %-38s HTTP %d  %6d bytes  %s\n", mark, s.Name(), status, len(body), preview(body))

		if p.RPS > 0 { // stay under the provider's rate limit
			time.Sleep(time.Duration(float64(time.Second) / p.RPS))
		}
	}
	if failed > 0 {
		os.Exit(1)
	}
}

func call(reg providers.Registry, configs string, s providers.Spec) (int, []byte, string, error) {
	params := s.New()
	data, err := os.ReadFile(filepath.Join(configs, "examples", s.Name()+".json"))
	if err != nil {
		return 0, nil, "", err
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(params); err != nil {
		return 0, nil, "", fmt.Errorf("example: %w", err)
	}

	p := reg[s.Provider]
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.TimeoutMS)*time.Millisecond)
	defer cancel()
	req, err := reg.NewHTTPRequest(ctx, providers.Request{Provider: s.Provider, Endpoint: s.Endpoint, Params: params})
	if err != nil {
		return 0, nil, "", err
	}
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "godnota-hackathon/0.1")
	}
	client := &http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			err = urlErr.Err
		}
		return 0, nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		return resp.StatusCode, nil, resp.Header.Get("Content-Type"), fmt.Errorf("provider returned HTTP redirect")
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20+1))
	if len(body) > 2<<20 {
		return resp.StatusCode, nil, resp.Header.Get("Content-Type"), fmt.Errorf("provider response exceeds 2 MiB")
	}
	return resp.StatusCode, body, resp.Header.Get("Content-Type"), err
}

func selected(only string, s providers.Spec) bool {
	if only == "" {
		return true
	}
	for _, name := range strings.Split(only, ",") {
		name = strings.TrimSpace(name)
		if name == s.Provider || name == s.Name() {
			return true
		}
	}
	return false
}

func missingKey(a providers.Auth) string {
	for _, env := range []string{a.Env, a.EnvUser, a.EnvPass} {
		if env != "" && os.Getenv(env) == "" {
			return env
		}
	}
	return ""
}

func secretValues(reg providers.Registry) []string {
	var vals []string
	for _, p := range reg {
		for _, env := range []string{p.Auth.Env, p.Auth.EnvUser, p.Auth.EnvPass} {
			if v := os.Getenv(env); env != "" && len(v) >= 6 {
				vals = append(vals, v)
			}
		}
	}
	return vals
}

func redact(s string, secrets []string) string {
	for _, v := range secrets {
		s = strings.ReplaceAll(s, v, "***")
		s = strings.ReplaceAll(s, url.QueryEscape(v), "***")
		s = strings.ReplaceAll(s, url.PathEscape(v), "***")
	}
	return s
}

// loadEnv reads KEY=VALUE lines; variables already set in the environment win.
func loadEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok || value == "" {
			continue
		}
		if _, set := os.LookupEnv(key); !set {
			os.Setenv(strings.TrimSpace(key), strings.TrimSpace(value))
		}
	}
	return sc.Err()
}

func extension(contentType string, body []byte) string {
	switch {
	case strings.Contains(contentType, "xml"), bytes.HasPrefix(bytes.TrimSpace(body), []byte("<?xml")):
		return ".xml"
	case strings.Contains(contentType, "json"), json.Valid(body):
		return ".json"
	default:
		return ".txt"
	}
}

func save(path string, body []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if json.Valid(body) { // pretty-print so diffs are readable
		var buf bytes.Buffer
		if json.Indent(&buf, body, "", "  ") == nil {
			body = append(buf.Bytes(), '\n')
		}
	}
	return os.WriteFile(path, body, 0o644)
}

func preview(body []byte) string {
	s := strings.Join(strings.Fields(string(body)), " ")
	if len([]rune(s)) > 110 {
		s = string([]rune(s)[:110]) + "…"
	}
	return s
}
