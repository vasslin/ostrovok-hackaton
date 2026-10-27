// Package godnota is the team's hackathon backend.
//
// The JSON Schemas in configs/ are generated from the Go structs in
// internal/trip and internal/providers. After changing a struct, run:
//
//	go generate ./...
package godnota

//go:generate go run ./cmd/genschema -out configs
