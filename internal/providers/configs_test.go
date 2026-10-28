package providers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/ostrovok-hackathon-2026/godnota/internal/providers"
)

const configsDir = "../../configs"

func loadRegistry(t *testing.T) providers.Registry {
	t.Helper()
	reg, err := providers.LoadRegistry(filepath.Join(configsDir, "providers.json"))
	if err != nil {
		t.Fatal(err)
	}
	return reg
}

// setFakeSecrets fills every env var the registry mentions, so requests can be built.
func setFakeSecrets(t *testing.T, reg providers.Registry) {
	for _, p := range reg {
		for _, env := range []string{p.Auth.Env, p.Auth.EnvUser, p.Auth.EnvPass} {
			if env != "" {
				t.Setenv(env, "test-"+strings.ToLower(env))
			}
		}
	}
}

func decodeStrict(t *testing.T, path string, v any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
}

func validateAgainstSchema(t *testing.T, schemaPath, examplePath string) {
	t.Helper()
	c := jsonschema.NewCompiler()
	c.AssertFormat()
	sch, err := c.Compile(schemaPath)
	if err != nil {
		t.Fatalf("compile %s: %v", schemaPath, err)
	}
	f, err := os.Open(examplePath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	inst, err := jsonschema.UnmarshalJSON(f)
	if err != nil {
		t.Fatal(err)
	}
	if err := sch.Validate(inst); err != nil {
		t.Errorf("%s does not match %s:\n%v", examplePath, schemaPath, err)
	}
}

func TestHotelbedsExampleSchema(t *testing.T) {
	validateAgainstSchema(t, "../../configs/hotelbeds/availability.schema.json", "../../configs/hotelbeds/availability.example.json")
}

func TestRegistryMatchesCatalog(t *testing.T) {
	reg := loadRegistry(t)
	var inRegistry, inCatalog []string
	for name, p := range reg {
		for ep := range p.Endpoints {
			inRegistry = append(inRegistry, name+"."+ep)
		}
	}
	for _, s := range providers.Catalog {
		inCatalog = append(inCatalog, s.Name())
	}
	sort.Strings(inRegistry)
	sort.Strings(inCatalog)
	if strings.Join(inRegistry, " ") != strings.Join(inCatalog, " ") {
		t.Fatalf("providers.json and Catalog differ:\nregistry: %v\ncatalog:  %v", inRegistry, inCatalog)
	}
}

// Every endpoint has an example that passes the Go tags, the generated schema,
// and turns into an HTTP request.
func TestEndpointExamples(t *testing.T) {
	reg := loadRegistry(t)
	setFakeSecrets(t, reg)
	for _, s := range providers.Catalog {
		t.Run(s.Name(), func(t *testing.T) {
			example := filepath.Join(configsDir, "examples", s.Name()+".json")
			params := s.New()
			decodeStrict(t, example, params)
			validateAgainstSchema(t, filepath.Join(configsDir, "providers", s.Name()+".schema.json"), example)
			if err := providers.Validate(params); err != nil {
				t.Fatalf("Go validation: %v", err)
			}

			req, err := reg.NewHTTPRequest(context.Background(), providers.Request{Provider: s.Provider, Endpoint: s.Endpoint, Params: params})
			if err != nil {
				t.Fatalf("NewHTTPRequest: %v", err)
			}
			ep := reg[s.Provider].Endpoints[s.Endpoint]
			if req.Method != ep.Method {
				t.Errorf("method = %s, want %s", req.Method, ep.Method)
			}
			if !strings.HasPrefix(req.URL.String(), reg[s.Provider].BaseURL) {
				t.Errorf("URL %s does not start with %s", req.URL, reg[s.Provider].BaseURL)
			}
		})
	}
}

func TestBuiltRequests(t *testing.T) {
	reg := loadRegistry(t)
	setFakeSecrets(t, reg)
	build := func(provider, endpoint string) (string, string) {
		t.Helper()
		s, _ := providers.Lookup(provider, endpoint)
		params := s.New()
		decodeStrict(t, filepath.Join(configsDir, "examples", s.Name()+".json"), params)
		req, err := reg.NewHTTPRequest(context.Background(), providers.Request{Provider: provider, Endpoint: endpoint, Params: params})
		if err != nil {
			t.Fatal(err)
		}
		body := ""
		if req.Body != nil {
			b, _ := io.ReadAll(req.Body)
			body = string(b)
		}
		return req.URL.String(), body
	}

	// GET with query params and header auth
	u, _ := build("aviasales", "prices_for_dates")
	want := "https://api.travelpayouts.com/aviasales/v3/prices_for_dates?currency=rub&departure_at=2026-12-12&destination=KZN&direct=true&limit=30&one_way=false&origin=MOW&page=1&return_at=2026-12-14&sorting=price"
	if u != want {
		t.Errorf("aviasales URL\n got %s\nwant %s", u, want)
	}

	// query auth lands in the URL
	if u, _ := build("yandex_rasp", "search"); !strings.Contains(u, "apikey=test-yandex_rasp_api_key") || !strings.Contains(u, "from=c213") {
		t.Errorf("rasp URL missing apikey or from: %s", u)
	}

	// path placeholders
	if u, _ := build("osrm", "route"); u != "https://router.project-osrm.org/route/v1/driving/49.2787,55.6062;49.1062,55.7963?overview=false" {
		t.Errorf("osrm URL = %s", u)
	}
	if u, _ := build("wikipedia", "geosearch"); !strings.HasPrefix(u, "https://ru.wikipedia.org/w/api.php?") || !strings.Contains(u, "ggscoord=55.7987%7C49.1059") {
		t.Errorf("wikipedia URL = %s", u)
	}

}

func TestInvalidParamsAreRejected(t *testing.T) {
	cases := map[string]any{
		"lowercase IATA":              &providers.AviasalesPricesForDates{Origin: "mow", Destination: "KZN", Currency: "rub"},
		"bad departure_at":            &providers.AviasalesPricesForDates{Origin: "MOW", Destination: "KZN", Currency: "rub", DepartureAt: "12.12.2026"},
		"uppercase currency":          &providers.AviasalesPricesForDates{Origin: "MOW", Destination: "KZN", Currency: "RUB"},
		"rasp unknown transport":      &providers.RaspSearch{From: "c213", To: "c43", TransportTypes: "car"},
		"rasp limit over 100":         &providers.RaspSearch{From: "c213", To: "c43", Limit: 101},
		"nominatim limit over 40":     &providers.NominatimSearch{Q: "Казань", Format: "jsonv2", Limit: 41},
		"marine forecast 9 days":      &providers.OpenMeteoMarine{Latitude: 27, Longitude: 33, ForecastDays: 9},
		"wikipedia radius > 10 km":    &providers.WikipediaGeosearch{Action: "query", Generator: "geosearch", Coord: "55.7987|49.1059", Radius: 20000, Format: "json"},
		"wikipedia coord without |":   &providers.WikipediaGeosearch{Action: "query", Generator: "geosearch", Coord: "55.7987,49.1059", Radius: 3000, Format: "json"},
		"amenity search without city": &providers.NominatimAmenitySearch{Amenity: "museum", Format: "jsonv2"},
	}
	for name, params := range cases {
		if err := providers.Validate(params); err == nil {
			t.Errorf("%s: expected a validation error", name)
		}
	}
}

func TestRequestParamsMustMatchEndpoint(t *testing.T) {
	reg := providers.Registry{
		"aviasales": {
			BaseURL: "https://example.com",
			Auth:    providers.Auth{Type: "none"},
			RPS:     1, TimeoutMS: 1000,
			Endpoints: map[string]providers.Endpoint{
				"prices_for_dates": {Method: "GET", Path: "/prices_for_dates", ParamsIn: "query"},
			},
		},
	}
	var nilPrices *providers.AviasalesPricesForDates
	cases := map[string]any{
		"nil":        nil,
		"typed nil":  nilPrices,
		"non-struct": "origin=MOW",
		"wrong type": &providers.NominatimSearch{Q: "Казань", Format: "jsonv2"},
	}
	for name, params := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := reg.NewHTTPRequest(context.Background(), providers.Request{Provider: "aviasales", Endpoint: "prices_for_dates", Params: params})
			if err == nil {
				t.Fatal("expected a request error")
			}
		})
	}
	_, err := reg.NewHTTPRequest(context.Background(), providers.Request{
		Provider: "aviasales",
		Endpoint: "prices_for_dates",
		Params: &providers.AviasalesPricesForDates{
			Origin: "MOW", Destination: "KZN", Currency: "rub",
		},
	})
	if err != nil {
		t.Fatalf("valid request: %v", err)
	}
}

func TestRegistryRejectsInvalidTransportAndEndpoints(t *testing.T) {
	valid := func() providers.Provider {
		return providers.Provider{
			BaseURL: "https://example.com",
			Auth:    providers.Auth{Type: "none"},
			RPS:     1, TimeoutMS: 1000,
			Endpoints: map[string]providers.Endpoint{
				"search": {Method: "GET", Path: "/search", ParamsIn: "query"},
			},
		}
	}
	cases := map[string]func(*providers.Provider){
		"http credentials": func(p *providers.Provider) {
			p.BaseURL = "http://example.com"
			p.Auth = providers.Auth{Type: "header", Name: "X-Access-Token", Env: "TEST_TOKEN"}
		},
		"url credentials": func(p *providers.Provider) { p.BaseURL = "https://user:pass@example.com" },
		"url query":       func(p *providers.Provider) { p.BaseURL = "https://example.com?token=foo" },
		"bad method": func(p *providers.Provider) {
			p.Endpoints["search"] = providers.Endpoint{Method: "TRACE", Path: "/search", ParamsIn: "query"}
		},
		"bad path": func(p *providers.Provider) {
			p.Endpoints["search"] = providers.Endpoint{Method: "GET", Path: "/../search", ParamsIn: "query"}
		},
		"bad auth": func(p *providers.Provider) {
			p.Auth = providers.Auth{Type: "header", Name: "X Token", Env: "TEST_TOKEN"}
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			p := valid()
			mutate(&p)
			data, err := json.Marshal(providers.Registry{"sample": p})
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), "providers.json")
			if err := os.WriteFile(path, data, 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := providers.LoadRegistry(path); err == nil {
				t.Fatal("expected registry validation error")
			}
		})
	}
	p := valid()
	p.BaseURL = "http://example.com"
	data, err := json.Marshal(providers.Registry{"sample": p})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "providers.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := providers.LoadRegistry(path); err != nil {
		t.Fatalf("no-auth HTTP mock must be allowed: %v", err)
	}
}
