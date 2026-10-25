// Command genschema writes the JSON Schema files under configs/.
// Run it with: go generate ./...
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ostrovok-hackathon-2026/godnota/internal/schema"
)

func main() {
	out := flag.String("out", "configs", "configs directory")
	flag.Parse()

	files, err := schema.Generate()
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		path := filepath.Join(*out, f.Path)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(path, f.JSON, 0o644); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Printf("wrote %d schemas to %s\n", len(files), *out)
}
