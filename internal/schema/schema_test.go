package schema_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/ostrovok-hackathon-2026/godnota/internal/schema"
)

// Fails when a struct changed but `go generate ./...` wasn't re-run.
func TestSchemasAreUpToDate(t *testing.T) {
	files, err := schema.Generate()
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		onDisk, err := os.ReadFile(filepath.Join("../../configs", f.Path))
		if err != nil {
			t.Errorf("%s: %v (run go generate ./...)", f.Path, err)
			continue
		}
		if !bytes.Equal(onDisk, f.JSON) {
			t.Errorf("%s is out of date: run go generate ./...", f.Path)
		}
	}
}
