// Package schema generates JSON Schema files from the Go config structs, so the
// structs stay the single source of truth for required fields and format limits.
package schema

import (
	"encoding/json"
	"path/filepath"

	"github.com/invopop/jsonschema"

	"github.com/ostrovok-hackathon-2026/godnota/internal/providers"
	"github.com/ostrovok-hackathon-2026/godnota/internal/trip"
)

// File is one generated schema: its path relative to configs/ and its JSON.
type File struct {
	Path string
	JSON []byte
}

// Generate builds every schema: configs/trip_request.schema.json and
// configs/providers/<provider>.<endpoint>.schema.json.
func Generate() ([]File, error) {
	files := []File{}
	add := func(path, title string, v any) error {
		b, err := render(title, v)
		if err != nil {
			return err
		}
		files = append(files, File{Path: path, JSON: b})
		return nil
	}

	if err := add("trip_request.schema.json", "TripRequest: unified trip config filled by the AI agent", &trip.TripRequest{}); err != nil {
		return nil, err
	}
	for _, s := range providers.Catalog {
		if err := add(filepath.Join("providers", s.Name()+".schema.json"), s.Name()+" request parameters", s.New()); err != nil {
			return nil, err
		}
	}
	return files, nil
}

func render(title string, v any) ([]byte, error) {
	r := jsonschema.Reflector{
		RequiredFromJSONSchemaTags: true, // required = jsonschema:"required", not "no omitempty"
		DoNotReference:             true, // inline nested objects: easier to read and to give an LLM
		ExpandedStruct:             true,
	}
	s := r.Reflect(v)
	s.Title = title
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}
