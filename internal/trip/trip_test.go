package trip_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/ostrovok-hackathon-2026/godnota/internal/trip"
)

var today = time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC)

func TestExamplesAreValid(t *testing.T) {
	c := jsonschema.NewCompiler()
	c.AssertFormat()
	sch, err := c.Compile("../../configs/trip_request.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	files, _ := filepath.Glob("../../configs/examples/trip_request.*.json")
	if len(files) == 0 {
		t.Fatal("no trip_request examples found")
	}
	for _, path := range files {
		t.Run(filepath.Base(path), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var tr trip.TripRequest
			dec := json.NewDecoder(bytes.NewReader(data))
			dec.DisallowUnknownFields()
			if err := dec.Decode(&tr); err != nil {
				t.Fatal(err)
			}
			if err := tr.ValidateAt(today); err != nil {
				t.Errorf("Go validation: %v", err)
			}
			if m := tr.MissingRequired(); len(m) != 0 {
				t.Errorf("missing required: %v", m)
			}
			inst, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
			if err != nil {
				t.Fatal(err)
			}
			if err := sch.Validate(inst); err != nil {
				t.Errorf("schema: %v", err)
			}
		})
	}
}

func TestMissingRequired(t *testing.T) {
	tr := trip.TripRequest{SchemaVersion: "1.0", Locale: "ru", Currency: "RUB"}
	got := tr.MissingRequired()
	want := []string{"/origin/city", "/dates/departure_date", "/dates/departure_month", "/travelers/adults"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("MissingRequired() = %v, want %v", got, want)
	}

	tr.Origin.City = "Москва"
	tr.Dates.DepartureMonth = "2026-11"
	tr.Travelers.Adults = 2
	if got := tr.MissingRequired(); len(got) != 0 {
		t.Errorf("after filling: %v", got)
	}
}

func TestCrossFieldRules(t *testing.T) {
	valid := func() trip.TripRequest {
		return trip.TripRequest{
			SchemaVersion: "1.0", Locale: "ru", Currency: "RUB",
			Origin:    trip.Place{City: "Москва"},
			Dates:     trip.Dates{DepartureDate: "2026-12-12", ReturnDate: "2026-12-14"},
			Travelers: trip.Travelers{Adults: 2},
		}
	}
	if err := valid().ValidateAt(today); err != nil {
		t.Fatalf("valid trip rejected: %v", err)
	}

	cases := map[string]func(*trip.TripRequest){
		"return before departure": func(r *trip.TripRequest) { r.Dates.ReturnDate = "2026-12-10" },
		"stay over 30 nights":     func(r *trip.TripRequest) { r.Dates.ReturnDate = "2027-01-20" },
		"departure in the past":   func(r *trip.TripRequest) { r.Dates.DepartureDate = "2026-10-01"; r.Dates.ReturnDate = "" },
		"departure > 365 days":    func(r *trip.TripRequest) { r.Dates.DepartureDate = "2027-12-01"; r.Dates.ReturnDate = "" },
		"bad month":               func(r *trip.TripRequest) { r.Dates = trip.Dates{DepartureMonth: "2026-13"} },
		"child aged 18":           func(r *trip.TripRequest) { r.Travelers.ChildrenAges = []trip.Age{18} },
		"10 adults":               func(r *trip.TripRequest) { r.Travelers.Adults = 10 },
		"unknown transport":       func(r *trip.TripRequest) { r.Transport = &trip.Transport{Modes: []string{"car"}} },
		"uppercase residency":     func(r *trip.TripRequest) { r.Travelers.Residency = "RU" },
		"bad time":                func(r *trip.TripRequest) { r.Transport = &trip.Transport{DepartureTimeFrom: "25:00"} },
		"assumption without path": func(r *trip.TripRequest) {
			r.AIContext = &trip.AIContext{Assumptions: []trip.Assumption{{Field: "origin", Value: "x", Reason: "y"}}}
		},
	}
	for name, mutate := range cases {
		r := valid()
		mutate(&r)
		if err := r.ValidateAt(today); err == nil {
			t.Errorf("%s: expected an error", name)
		} else if testing.Verbose() {
			t.Logf("%s: %s", name, strings.ReplaceAll(err.Error(), "\n", "; "))
		}
	}
}
