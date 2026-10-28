package trip

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
)

// MaxStayNights matches the hotel API limit (checkout at most 30 days after checkin).
const MaxStayNights = 30

// MaxDaysAhead is how far ahead a departure may be.
const MaxDaysAhead = 365

var (
	validateOnce sync.Once
	validate     *validator.Validate
)

// Validator returns a validator that reports fields by their JSON names.
func Validator() *validator.Validate {
	validateOnce.Do(func() {
		validate = validator.New(validator.WithRequiredStructEnabled())
		validate.RegisterTagNameFunc(func(f reflect.StructField) string {
			name := strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})
	})
	return validate
}

// ValidateAt checks the struct tags plus cross-field rules, relative to now.
func (t TripRequest) ValidateAt(now time.Time) error {
	var errs []error
	if err := Validator().Struct(t); err != nil {
		errs = append(errs, err)
	}

	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	dep, depErr := time.Parse(time.DateOnly, t.Dates.DepartureDate)
	if t.Dates.DepartureDate != "" && depErr == nil {
		if dep.Before(today) {
			errs = append(errs, fmt.Errorf("/dates/departure_date: %s is in the past", t.Dates.DepartureDate))
		}
		if dep.After(today.AddDate(0, 0, MaxDaysAhead)) {
			errs = append(errs, fmt.Errorf("/dates/departure_date: more than %d days ahead", MaxDaysAhead))
		}
		if t.Dates.DepartureMonth != "" && t.Dates.DepartureMonth != dep.Format("2006-01") {
			errs = append(errs, errors.New("/dates/departure_month: disagrees with departure_date"))
		}
	}
	if t.Dates.DepartureMonth != "" {
		if month, err := time.Parse("2006-01", t.Dates.DepartureMonth); err == nil {
			lastDay := month.AddDate(0, 1, -1)
			if lastDay.Before(today) {
				errs = append(errs, errors.New("/dates/departure_month: is in the past"))
			}
			if month.After(today.AddDate(0, 0, MaxDaysAhead)) {
				errs = append(errs, fmt.Errorf("/dates/departure_month: more than %d days ahead", MaxDaysAhead))
			}
		}
	}
	if t.Dates.ReturnDate != "" && t.Dates.DepartureDate == "" {
		errs = append(errs, errors.New("/dates/return_date: requires departure_date"))
	}
	if t.Dates.ReturnDate != "" && depErr == nil {
		if ret, err := time.Parse(time.DateOnly, t.Dates.ReturnDate); err == nil {
			if !ret.After(dep) {
				errs = append(errs, errors.New("/dates/return_date: must be after departure_date"))
			}
			if ret.Sub(dep) > MaxStayNights*24*time.Hour {
				errs = append(errs, fmt.Errorf("/dates/return_date: stay is longer than %d nights", MaxStayNights))
			}
			if t.Dates.Nights > 0 && int(ret.Sub(dep).Hours()/24) != t.Dates.Nights {
				errs = append(errs, errors.New("/dates/nights: disagrees with return_date"))
			}
		}
	}
	return errors.Join(errs...)
}

// MissingRequired lists JSON paths of required fields that are empty, for
// Completeness.MissingRequired: the backend uses the count to decide whether
// the agent asks in chat or the frontend shows a short form.
func (t TripRequest) MissingRequired() []string {
	var verrs validator.ValidationErrors
	if !errors.As(Validator().Struct(t), &verrs) {
		return []string{}
	}
	missing := []string{}
	for _, fe := range verrs {
		if strings.HasPrefix(fe.Tag(), "required") {
			missing = append(missing, jsonPath(fe.Namespace()))
		}
	}
	return missing
}

// jsonPath turns "TripRequest.travelers.pets[0].kind" into "/travelers/pets/0/kind".
func jsonPath(namespace string) string {
	parts := strings.Split(namespace, ".")[1:] // drop the root type name
	path := ""
	for _, p := range parts {
		name, index, hasIndex := strings.Cut(p, "[")
		path += "/" + name
		if hasIndex {
			path += "/" + strings.TrimSuffix(index, "]")
		}
	}
	return path
}
