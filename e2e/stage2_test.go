//go:build e2e

// These tests build and launch separate native processes, then use the public HTTP API.
package e2e

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"github.com/ostrovok-hackathon-2026/godnota/internal/devtool"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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

func demoOrder(t *testing.T, base, token, key string, body map[string]any, expected int) map[string]any {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest("POST", base+"/api/orders", bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", key)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != expected {
		t.Fatalf("demo order HTTP %d, expected %d", resp.StatusCode, expected)
	}
	return result
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
		candidates, ok := result["candidates"].([]any)
		if !ok || len(candidates) == 0 {
			t.Fatal("stage-3 candidates missing")
		}
		t.Logf("synthetic backend route candidates: %d", len(candidates))
		first := candidates[0].(map[string]any)
		if first["bookable"] != false || first["price_status"] != "unknown" || first["ranking"] != "deterministic" {
			t.Fatal("route provenance or price status lost")
		}
		cached := call(t, api, "GET", path+"/routes", token, nil, 200)
		if len(cached["candidates"].([]any)) != len(candidates) {
			t.Fatal("cached candidates lost")
		}
		orderBody := map[string]any{"session_id": id, "route_id": first["id"]}
		order := demoOrder(t, cluster.URL, token, "e2e-order-1", orderBody, 200)
		repeated := demoOrder(t, cluster.URL, token, "e2e-order-1", orderBody, 200)
		if order["order_id"] == "" || repeated["order_id"] != order["order_id"] || order["mode"] != "demo" || order["status"] != "demo_booked" {
			t.Fatal("demo order idempotency or mode lost")
		}
		demoOrder(t, cluster.URL, token, "", orderBody, 400)
		demoOrder(t, cluster.URL, token, "other-key", map[string]any{"session_id": id, "route_id": "nonexistent"}, 404)
		demoOrder(t, cluster.URL, token, "other-key", map[string]any{"session_id": id, "route_id": first["id"], "route_snapshot": map[string]any{"id": first["id"]}}, 400)
		if len(candidates) > 1 {
			second := candidates[1].(map[string]any)
			demoOrder(t, cluster.URL, token, "e2e-order-1", map[string]any{"session_id": id, "route_id": second["id"]}, 409)
		}
		orderPath := "/api/orders/" + order["order_id"].(string)
		stored := call(t, api, "GET", orderPath, token, nil, 200)
		if stored["route_snapshot"].(map[string]any)["id"] != first["id"] {
			t.Fatal("route snapshot missing")
		}
		if call(t, api, "POST", orderPath+"/pay", token, map[string]any{}, 200)["status"] != "demo_paid" {
			t.Fatal("demo payment failed")
		}
		if call(t, api, "POST", orderPath+"/cancel", token, map[string]any{}, 200)["status"] != "demo_cancelled" {
			t.Fatal("demo cancellation failed")
		}
		stranger := call(t, api, "POST", "/api/auth/register", "", credential(), 201)["access_token"].(string)
		call(t, api, "GET", orderPath, stranger, nil, 404)
		call(t, api, "POST", orderPath+"/pay", stranger, map[string]any{}, 404)
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
		stored = call(t, api, "GET", orderPath, token, nil, 200)
		if stored["route_snapshot"].(map[string]any)["id"] != first["id"] {
			t.Fatal("order snapshot changed after config edit")
		}
		call(t, api, "POST", path+"/search", token, map[string]any{}, 409)
		now := time.Now()
		d1, d2, d3 := now.AddDate(0, 0, 7).Format(time.DateOnly), now.AddDate(0, 0, 8).Format(time.DateOnly), now.AddDate(0, 0, 9).Format(time.DateOnly)
		multi := devtool.ExampleTrip(now)
		multi["destination"] = map[string]any{"city": "Самара", "country_code": "RU"}
		multi["stops"] = []any{map[string]any{"city": "Казань", "check_in": d1, "check_out": d2}, map[string]any{"city": "Самара", "check_in": d2, "check_out": d3}}
		call(t, api, "PATCH", path+"/config", token, map[string]any{"config": map[string]any{"version": "1.0", "request": multi}}, 200)
		call(t, api, "POST", path+"/confirm", token, map[string]any{}, 200)
		multiResult := call(t, api, "POST", path+"/search", token, map[string]any{}, 200)
		multiRoutes, ok := multiResult["candidates"].([]any)
		if !ok || len(multiRoutes) == 0 {
			t.Fatal("multi-city routes missing")
		}
		m := multiRoutes[0].(map[string]any)
		if len(m["transport"].([]any)) != 3 || len(m["hotels"].([]any)) != 2 {
			t.Fatal("multi-city legs or stays missing")
		}
		t.Logf("synthetic multi-city routes: %d; demo order lifecycle passed", len(multiRoutes))
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

func TestFullStackStartsUIAndStopsEveryChild(t *testing.T) {
	root, err := devtool.RepoRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(t.TempDir(), "bin")
	if err = devtool.Build(context.Background(), root, bin, io.Discard); err != nil {
		t.Fatal(err)
	}
	cluster, err := devtool.Start(context.Background(), devtool.Options{Root: root, BinDir: bin, DataDir: t.TempDir(), Mode: "demo", RandomPorts: true, FullStack: true, Output: io.Discard})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cluster.Close)
	if err = devtool.Smoke(context.Background(), cluster.WebURL, io.Discard); err != nil {
		t.Fatal("UI proxy:", err)
	}
	response, err := http.Get(cluster.WebURL + "/config.js")
	if err != nil {
		t.Fatal(err)
	}
	data, _ := io.ReadAll(response.Body)
	response.Body.Close()
	if !strings.Contains(string(data), "apiBaseUrl: '/api'") {
		t.Fatal("UI still needs another device's localhost")
	}
	addresses := []string{strings.TrimPrefix(cluster.WebURL, "http://")}
	for _, port := range cluster.Ports {
		addresses = append(addresses, net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	}
	cluster.Close()
	for _, address := range addresses {
		conn, err := net.DialTimeout("tcp", address, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			t.Fatalf("child port remained open after shutdown: %s", address)
		}
	}
}
