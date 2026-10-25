package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"

	"github.com/google/go-querystring/query"

	"github.com/ostrovok-hackathon-2026/godnota/internal/trip"
)

// Request is one provider call: which endpoint, and its parameters (one of the
// structs in Catalog).
type Request struct {
	Provider string
	Endpoint string
	Params   any
}

// RequestSet maps a provider name to its calls for one search. A slice, because
// a provider may need several calls (outbound + return leg, search → hotel page).
type RequestSet map[string][]Request

// Validate runs the struct tags and, if the params have a Check() method, the
// cross-field rules.
func Validate(params any) error {
	if err := trip.Validator().Struct(params); err != nil {
		return err
	}
	if c, ok := params.(interface{ Check() error }); ok {
		return c.Check()
	}
	return nil
}

// NewHTTPRequest validates r and turns it into an *http.Request using the
// registry: base URL, method, path placeholders, query or JSON encoding, auth.
func (reg Registry) NewHTTPRequest(ctx context.Context, r Request) (*http.Request, error) {
	p, ok := reg[r.Provider]
	if !ok {
		return nil, fmt.Errorf("unknown provider %q", r.Provider)
	}
	ep, ok := p.Endpoints[r.Endpoint]
	if !ok {
		return nil, fmt.Errorf("unknown endpoint %s.%s", r.Provider, r.Endpoint)
	}
	if err := Validate(r.Params); err != nil {
		return nil, fmt.Errorf("%s.%s: %w", r.Provider, r.Endpoint, err)
	}

	path, err := fillPath(ep.Path, r.Params)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(strings.TrimSuffix(p.BaseURL, "/") + path)
	if err != nil {
		return nil, err
	}

	var body io.Reader
	switch ep.ParamsIn {
	case "query":
		q, err := query.Values(r.Params)
		if err != nil {
			return nil, err
		}
		u.RawQuery = q.Encode()
	case "json":
		b, err := json.Marshal(r.Params)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(b)
	default:
		return nil, fmt.Errorf("%s.%s: unknown params_in %q", r.Provider, r.Endpoint, ep.ParamsIn)
	}

	req, err := http.NewRequestWithContext(ctx, ep.Method, u.String(), body)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	for k, v := range p.Headers {
		req.Header.Set(k, v)
	}
	if err := applyAuth(req, p.Auth); err != nil {
		return nil, fmt.Errorf("%s: %w", r.Provider, err)
	}
	return req, nil
}

func applyAuth(req *http.Request, a Auth) error {
	switch a.Type {
	case "", "none":
		return nil
	case "header":
		v, err := secret(a.Env)
		if err != nil {
			return err
		}
		req.Header.Set(a.Name, v)
	case "query":
		v, err := secret(a.Env)
		if err != nil {
			return err
		}
		q := req.URL.Query()
		q.Set(a.Name, v)
		req.URL.RawQuery = q.Encode()
	case "basic":
		user, err := secret(a.EnvUser)
		if err != nil {
			return err
		}
		pass, err := secret(a.EnvPass)
		if err != nil {
			return err
		}
		req.SetBasicAuth(user, pass)
	default:
		return fmt.Errorf("unknown auth type %q", a.Type)
	}
	return nil
}

func secret(env string) (string, error) {
	v := os.Getenv(env)
	if v == "" {
		return "", fmt.Errorf("env %s is not set", env)
	}
	return v, nil
}

// fillPath replaces {name} in the path with the field tagged path:"name".
func fillPath(path string, params any) (string, error) {
	v := reflect.Indirect(reflect.ValueOf(params))
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		name := t.Field(i).Tag.Get("path")
		if name == "" {
			continue
		}
		// keep , and ; readable: OSRM coordinates are "lon,lat;lon,lat"
		value := strings.NewReplacer("%2C", ",", "%3B", ";").Replace(url.PathEscape(fmt.Sprint(v.Field(i).Interface())))
		path = strings.ReplaceAll(path, "{"+name+"}", value)
	}
	if strings.ContainsAny(path, "{}") {
		return "", fmt.Errorf("path %q has unfilled placeholders", path)
	}
	return path, nil
}
