package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ostrovok-hackathon-2026/godnota/providerkit"
)

func main() {
	tripPath := flag.String("trip", "configs/examples/trip_request.kazan_weekend.json", "TripRequest JSON path")
	resolvedPath := flag.String("resolved", "", "resolved provider codes and stay dates JSON path")
	nowText := flag.String("now", "", "today in YYYY-MM-DD for reproducible planning")
	flag.Parse()

	var input providerkit.TripRequest
	if err := readJSON(*tripPath, &input); err != nil {
		fail(err)
	}
	var resolved providerkit.ResolvedTrip
	if *resolvedPath != "" {
		if err := readJSON(*resolvedPath, &resolved); err != nil {
			fail(err)
		}
	}
	now := time.Now()
	if *nowText != "" {
		var err error
		now, err = time.Parse(time.DateOnly, *nowText)
		if err != nil {
			fail(err)
		}
	}
	plan, err := providerkit.BuildTripPlan(input, resolved, now)
	if err != nil {
		fail(err)
	}
	output, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		fail(err)
	}
	fmt.Println(string(output))
}

func readJSON(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return fmt.Errorf("%s: trailing JSON value", path)
	}
	return nil
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
