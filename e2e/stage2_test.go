//go:build e2e

// These tests build and launch separate native processes, then use the public HTTP API.
package e2e

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"github.com/ostrovok-hackathon-2026/godnota/internal/devtool"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func call(t *testing.T, a *devtool.API, method, path, token string, body any, code int) map[string]any {
	t.Helper()
	result, err := a.Call(context.Background(), method, path, token, body, code)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
func credential() map[string]string {
	return map[string]string{"email": strings.ToLower(rand.Text()) + "@example.com", "password": rand.Text()}
}
func start(t *testing.T, root, bin, mode string) *devtool.Cluster {
	t.Helper()
	data := t.TempDir()
	cluster, err := devtool.Start(context.Background(), devtool.Options{Root: root, BinDir: bin, DataDir: data, Mode: mode, RandomPorts: true, Output: io.Discard, ExtraEnv: map[string]string{"TRAVELPAYOUTS_TOKEN": "", "YANDEX_RASP_API_KEY": ""}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cluster.Close()
		if t.Failed() {
			logs, _ := filepath.Glob(filepath.Join(data, "logs", "*.log"))
			for _, p := range logs {
				raw, _ := os.ReadFile(p)
				if len(raw) > 2000 {
					raw = raw[len(raw)-2000:]
				}
				t.Log(filepath.Base(p), string(raw))
			}
		}
	})
	return cluster
}
func TestStage2EndToEnd(t *testing.T) {
	root, err := devtool.RepoRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(t.TempDir(), "bin")
	build, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	if err = devtool.Build(build, root, bin, os.Stdout); err != nil {
		t.Fatal(err)
	}
	t.Run("demo authenticated lifecycle and ownership", func(t *testing.T) {
		cluster := start(t, root, bin, "demo")
		api := devtool.NewAPI(cluster.URL)
		call(t, api, "GET", "/health", "", nil, 200)
		call(t, api, "POST", "/api/sessions", "", map[string]any{}, 401)
		creds := credential()
		registered := call(t, api, "POST", "/api/auth/register", "", creds, 201)
		old := registered["access_token"].(string)
		call(t, api, "POST", "/api/auth/register", "", creds, 409)
		bad := map[string]string{"email": creds["email"], "password": "incorrect-password"}
		call(t, api, "POST", "/api/auth/login", "", bad, 401)
		login := call(t, api, "POST", "/api/auth/login", "", creds, 200)
		token := login["access_token"].(string)
		call(t, api, "GET", "/api/auth/me", old, nil, 401)
		call(t, api, "GET", "/api/auth/me", token, nil, 200)
		created := call(t, api, "POST", "/api/sessions", token, map[string]any{}, 200)
		id := created["session_id"].(string)
		path := "/api/sessions/" + id
		call(t, api, "POST", "/api/sessions", token, map[string]any{"user_id": "forged"}, 403)
		call(t, api, "POST", path+"/search", token, map[string]any{}, 409)
		trip := devtool.ExampleTrip(time.Now())
		mapped := call(t, api, "POST", "/api/providers/map", token, map[string]any{"config": trip, "resolved": devtool.ExampleResolved()}, 200)
		providers := map[string]bool{}
		for _, c := range mapped["configs"].([]any) {
			providers[c.(map[string]any)["provider"].(string)] = true
		}
		for _, name := range []string{"hotel", "aviasales", "yandex_rasp"} {
			if !providers[name] {
				t.Fatal("provider mapping missing", name)
			}
		}
		call(t, api, "POST", "/api/providers/map", token, map[string]any{"config": map[string]any{"unknown": true}}, 400)
		body := map[string]any{"config": map[string]any{"version": "1.0", "request": trip}}
		call(t, api, "PATCH", path+"/config", token, body, 200)
		got := call(t, api, "GET", path+"/config", token, nil, 200)
		if got["config"] == nil {
			t.Fatal("saved config missing")
		}
		call(t, api, "POST", path+"/confirm", token, map[string]any{}, 200)
		result := call(t, api, "POST", path+"/search", token, map[string]any{}, 200)
		parts := result["parts"].(map[string]any)
		transport := parts["transport_parts"].([]any)
		hotels := parts["hotel_parts"].([]any)
		if len(transport) != 4 || len(hotels) != 1 {
			t.Fatalf("unexpected parts: %+v", parts)
		}
		hotel := hotels[0].(map[string]any)
		if hotel["source"] != "synthetic" || hotel["price"].(map[string]any)["amount_minor"] != "1000000" {
			t.Fatal("price/provenance contract broken")
		}
		limitations := fmt.Sprint(parts["limitations"])
		if !strings.Contains(limitations, "Unimplemented") {
			t.Fatal("stage-3 stub must be reported, not silently succeed")
		}
		call(t, api, "GET", path+"/routes", token, nil, 200)
		stranger := call(t, api, "POST", "/api/auth/register", "", credential(), 201)["access_token"].(string)
		for _, endpoint := range []string{path, path + "/config", path + "/routes"} {
			call(t, api, "GET", endpoint, stranger, nil, 404)
		}
		call(t, api, "POST", path+"/search", stranger, map[string]any{}, 404)
		call(t, api, "PATCH", path+"/config", stranger, body, 404)
		var wg sync.WaitGroup
		failures := make(chan error, 8)
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := api.Call(context.Background(), "GET", path+"/config", token, nil, 200)
				if err != nil {
					failures <- err
				}
			}()
		}
		wg.Wait()
		close(failures)
		for err := range failures {
			t.Error(err)
		}
		call(t, api, "PATCH", path+"/config", token, body, 200)
		call(t, api, "GET", path+"/routes", token, nil, 404)
		call(t, api, "POST", path+"/search", token, map[string]any{}, 409)
		call(t, api, "POST", "/api/auth/logout", token, nil, 204)
		call(t, api, "GET", "/api/auth/me", token, nil, 401)
	})
	t.Run("live missing credentials preserves explicit provider errors", func(t *testing.T) {
		cluster := start(t, root, bin, "live")
		api := devtool.NewAPI(cluster.URL)
		token := call(t, api, "POST", "/api/auth/register", "", credential(), 201)["access_token"].(string)
		id := call(t, api, "POST", "/api/sessions", token, map[string]any{}, 200)["session_id"].(string)
		path := "/api/sessions/" + id
		body := map[string]any{"config": map[string]any{"request": devtool.ExampleTrip(time.Now()), "resolved": devtool.ExampleResolved()}}
		call(t, api, "PATCH", path+"/config", token, body, 200)
		call(t, api, "POST", path+"/confirm", token, map[string]any{}, 200)
		parts := call(t, api, "POST", path+"/search", token, map[string]any{}, 200)["parts"].(map[string]any)
		if len(parts["transport_parts"].([]any)) != 0 || len(parts["hotel_parts"].([]any)) != 0 || len(parts["errors"].([]any)) == 0 {
			t.Fatalf("live error replaced by invented offers: %+v", parts)
		}
	})
	t.Run("portable smoke command", func(t *testing.T) {
		cluster := start(t, root, bin, "demo")
		var output bytes.Buffer
		if err := devtool.Smoke(context.Background(), cluster.URL, &output); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(output.String(), "PASS:") {
			t.Fatal("smoke did not finish")
		}
	})
}
