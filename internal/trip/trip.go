// Package trip defines TripRequest: the single config the AI agent fills from the
// user's message. It holds human values only (city names, dates, counts, wishes);
// provider codes (IATA, Yandex codes, hotel ids) are resolved later in Go.
package trip

import (
	"encoding/json"

	"github.com/invopop/jsonschema"
)

const SchemaVersion = "1.0"

type TripRequest struct {
	SchemaVersion string        `json:"schema_version" validate:"required,eq=1.0" jsonschema:"required,enum=1.0"`
	Locale        string        `json:"locale" validate:"required,oneof=ru en" jsonschema:"required,enum=ru,enum=en,default=ru"`
	Currency      string        `json:"currency" validate:"required,len=3,uppercase" jsonschema:"required,pattern=^[A-Z]{3}$,default=RUB" jsonschema_description:"ISO 4217 code for all money fields."`
	Origin        Place         `json:"origin" jsonschema:"required"`
	Destination   *Destination  `json:"destination,omitempty" jsonschema_description:"Omit (or null) when the traveler doesn't know where to go: inspiration mode."`
	Dates         Dates         `json:"dates" jsonschema:"required"`
	Stops         []Stop        `json:"stops,omitempty" validate:"omitempty,max=6,dive" jsonschema:"maxItems=6" jsonschema_description:"Ordered stays between origin and the return home. Empty means one destination."`
	Travelers     Travelers     `json:"travelers" jsonschema:"required"`
	Transport     *Transport    `json:"transport,omitempty"`
	Hotel         *Hotel        `json:"hotel,omitempty"`
	Leisure       *Leisure      `json:"leisure,omitempty"`
	Budget        *Budget       `json:"budget,omitempty"`
	AIContext     *AIContext    `json:"ai_context,omitempty"`
	Completeness  *Completeness `json:"completeness,omitempty" jsonschema_description:"Filled by the backend after validation, never by the AI."`
}

type Place struct {
	City        string `json:"city" validate:"required,min=1,max=100" jsonschema:"required,minLength=1,maxLength=100" jsonschema_description:"City name as the traveler said it, e.g. Москва."`
	CountryCode string `json:"country_code,omitempty" validate:"omitempty,len=2,uppercase" jsonschema:"pattern=^[A-Z]{2}$" jsonschema_description:"ISO 3166-1 alpha-2."`
}

type Stop struct {
	City     string `json:"city" validate:"required,min=1,max=100" jsonschema:"required,minLength=1,maxLength=100"`
	CheckIn  string `json:"check_in" validate:"required,datetime=2006-01-02" jsonschema:"required,format=date"`
	CheckOut string `json:"check_out" validate:"required,datetime=2006-01-02" jsonschema:"required,format=date"`
}

type Destination struct {
	City        string   `json:"city,omitempty" validate:"omitempty,min=1,max=100" jsonschema:"minLength=1,maxLength=100"`
	CountryCode string   `json:"country_code,omitempty" validate:"omitempty,len=2,uppercase" jsonschema:"pattern=^[A-Z]{2}$"`
	Candidates  []string `json:"candidates,omitempty" validate:"omitempty,max=5,dive,min=1,max=100" jsonschema:"maxItems=5" jsonschema_description:"Cities the AI suggests when the destination is vague."`
}

type Dates struct {
	DepartureDate  string `json:"departure_date,omitempty" validate:"required_without=DepartureMonth,omitempty,datetime=2006-01-02" jsonschema:"format=date"`
	DepartureMonth string `json:"departure_month,omitempty" validate:"required_without=DepartureDate,omitempty,datetime=2006-01" jsonschema:"pattern=^[0-9]{4}-(0[1-9]|1[0-2])$" jsonschema_description:"Use when only the month is known (YYYY-MM)."`
	ReturnDate     string `json:"return_date,omitempty" validate:"omitempty,datetime=2006-01-02" jsonschema:"format=date" jsonschema_description:"After departure_date and at most 30 days later (hotel API limit)."`
	Nights         int    `json:"nights,omitempty" validate:"omitempty,min=1,max=30" jsonschema:"minimum=1,maximum=30"`
	FlexibleDays   int    `json:"flexible_days,omitempty" validate:"omitempty,min=0,max=7" jsonschema:"minimum=0,maximum=7"`
}

// JSONSchemaExtend adds the "departure_date or departure_month" rule.
func (Dates) JSONSchemaExtend(s *jsonschema.Schema) {
	s.AnyOf = []*jsonschema.Schema{
		{Required: []string{"departure_date"}},
		{Required: []string{"departure_month"}},
	}
}

type Travelers struct {
	Adults       int    `json:"adults" validate:"required,min=1,max=9" jsonschema:"required,minimum=1,maximum=9"`
	ChildrenAges []Age  `json:"children_ages,omitempty" validate:"omitempty,max=8,dive,min=0,max=17" jsonschema:"maxItems=8"`
	Pets         []Pet  `json:"pets,omitempty" validate:"omitempty,max=3,dive" jsonschema:"maxItems=3"`
	Residency    string `json:"residency,omitempty" validate:"omitempty,len=2,lowercase" jsonschema:"pattern=^[a-z]{2}$,default=ru" jsonschema_description:"Citizenship, ISO 3166-1 alpha-2 lowercase (hotel API format)."`
}

// Age is a child's age in full years (0-17).
type Age int

func (Age) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{Type: "integer", Minimum: json.Number("0"), Maximum: json.Number("17")}
}

type Pet struct {
	Kind     string  `json:"kind" validate:"required,oneof=dog cat other" jsonschema:"required,enum=dog,enum=cat,enum=other"`
	WeightKg float64 `json:"weight_kg,omitempty" validate:"omitempty,gt=0,max=100" jsonschema:"exclusiveMinimum=0,maximum=100"`
}

type Transport struct {
	Modes              []string `json:"modes,omitempty" validate:"omitempty,min=1,dive,oneof=plane train suburban bus water helicopter" jsonschema:"minItems=1,uniqueItems=true"`
	RoundTrip          *bool    `json:"round_trip,omitempty" jsonschema:"default=true"`
	DirectOnly         bool     `json:"direct_only,omitempty" jsonschema:"default=false"`
	MaxTransfers       *int     `json:"max_transfers,omitempty" validate:"omitempty,min=0,max=3" jsonschema:"minimum=0,maximum=3"`
	MinTransferMinutes int      `json:"min_transfer_minutes,omitempty" validate:"omitempty,min=0,max=1440" jsonschema:"minimum=0,maximum=1440"`
	MaxWaitHours       int      `json:"max_wait_hours,omitempty" validate:"omitempty,min=0,max=48" jsonschema:"minimum=0,maximum=48"`
	MaxDurationHours   float64  `json:"max_duration_hours,omitempty" validate:"omitempty,min=0.5,max=48" jsonschema:"minimum=0.5,maximum=48"`
	DepartureTimeFrom  string   `json:"departure_time_from,omitempty" validate:"omitempty,datetime=15:04" jsonschema:"pattern=^([01][0-9]|2[0-3]):[0-5][0-9]$" jsonschema_description:"HH:MM, local time."`
	DepartureTimeTo    string   `json:"departure_time_to,omitempty" validate:"omitempty,datetime=15:04" jsonschema:"pattern=^([01][0-9]|2[0-3]):[0-5][0-9]$"`
}

// JSONSchemaExtend sets the allowed values for each item of modes.
func (Transport) JSONSchemaExtend(s *jsonschema.Schema) {
	if p, ok := s.Properties.Get("modes"); ok {
		p.Items = enumOf("plane", "train", "suburban", "bus", "water", "helicopter")
	}
}

type Hotel struct {
	Needed           *bool    `json:"needed,omitempty" jsonschema:"default=true"`
	Rooms            int      `json:"rooms,omitempty" validate:"omitempty,min=1,max=9" jsonschema:"minimum=1,maximum=9"`
	StarRatingMin    *int     `json:"star_rating_min,omitempty" validate:"omitempty,min=0,max=5" jsonschema:"minimum=0,maximum=5"`
	Kinds            []string `json:"kinds,omitempty" validate:"omitempty,dive,oneof=hotel hostel apartment guest_house resort" jsonschema:"uniqueItems=true"`
	Meal             string   `json:"meal,omitempty" validate:"omitempty,oneof=nomeal breakfast half_board full_board all_inclusive" jsonschema:"enum=nomeal,enum=breakfast,enum=half_board,enum=full_board,enum=all_inclusive"`
	FreeCancellation bool     `json:"free_cancellation,omitempty"`
	PricePerNightMax int      `json:"price_per_night_max,omitempty" validate:"omitempty,gt=0" jsonschema:"exclusiveMinimum=0" jsonschema_description:"In TripRequest.currency."`
	Amenities        []string `json:"amenities,omitempty" validate:"omitempty,dive,oneof=pool kids_club pet_friendly parking wifi spa beach accessible" jsonschema:"uniqueItems=true"`
	Location         []string `json:"location,omitempty" validate:"omitempty,dive,oneof=center beach quiet walkable near_station near_airport" jsonschema:"uniqueItems=true"`
	NearPlace        string   `json:"near_place,omitempty" validate:"omitempty,max=100" jsonschema:"maxLength=100" jsonschema_description:"A landmark to stay close to, e.g. Казанский Кремль."`
	EarlyCheckin     string   `json:"early_checkin,omitempty" validate:"omitempty,datetime=15:04" jsonschema:"pattern=^([01][0-9]|2[0-3]):[0-5][0-9]$"`
	LateCheckout     string   `json:"late_checkout,omitempty" validate:"omitempty,datetime=15:04" jsonschema:"pattern=^([01][0-9]|2[0-3]):[0-5][0-9]$"`
}

func (Hotel) JSONSchemaExtend(s *jsonschema.Schema) {
	setItems(s, "kinds", "hotel", "hostel", "apartment", "guest_house", "resort")
	setItems(s, "amenities", "pool", "kids_club", "pet_friendly", "parking", "wifi", "spa", "beach", "accessible")
	setItems(s, "location", "center", "beach", "quiet", "walkable", "near_station", "near_airport")
}

type Leisure struct {
	Interests       []string `json:"interests,omitempty" validate:"omitempty,max=8,dive,oneof=museums cinemas theatres concerts festivals parks nature architecture history food kids sport nightlife" jsonschema:"maxItems=8,uniqueItems=true"`
	DuringLayovers  *bool    `json:"during_layovers,omitempty" jsonschema:"default=true"`
	MinLayoverHours float64  `json:"min_layover_hours,omitempty" validate:"omitempty,min=2,max=24" jsonschema:"minimum=2,maximum=24,default=4"`
	MaxItemsPerStop int      `json:"max_items_per_stop,omitempty" validate:"omitempty,min=1,max=20" jsonschema:"minimum=1,maximum=20,default=5"`
}

func (Leisure) JSONSchemaExtend(s *jsonschema.Schema) {
	setItems(s, "interests", "museums", "cinemas", "theatres", "concerts", "festivals", "parks", "nature",
		"architecture", "history", "food", "kids", "sport", "nightlife")
}

type Budget struct {
	Total  int    `json:"total,omitempty" validate:"omitempty,gt=0" jsonschema:"exclusiveMinimum=0" jsonschema_description:"In TripRequest.currency."`
	Scope  string `json:"scope,omitempty" validate:"omitempty,oneof=total per_person per_night" jsonschema:"enum=total,enum=per_person,enum=per_night,default=total"`
	IsHard bool   `json:"is_hard,omitempty" jsonschema:"default=false"`
}

// AIContext holds what the AI infers beyond the explicit fields: trip purpose,
// wishes like a walkable area, and the assumptions it made.
type AIContext struct {
	TripPurpose  string       `json:"trip_purpose,omitempty" validate:"omitempty,oneof=leisure business family romantic friends other" jsonschema:"enum=leisure,enum=business,enum=family,enum=romantic,enum=friends,enum=other"`
	WalkableArea bool         `json:"walkable_area,omitempty" jsonschema_description:"«прогулочный район»"`
	EventsWanted []string     `json:"events_wanted,omitempty" validate:"omitempty,max=5,dive,min=1,max=50" jsonschema:"maxItems=5"`
	MustHave     []string     `json:"must_have,omitempty" validate:"omitempty,max=10,dive,min=1,max=80" jsonschema:"maxItems=10"`
	MustAvoid    []string     `json:"must_avoid,omitempty" validate:"omitempty,max=10,dive,min=1,max=80" jsonschema:"maxItems=10"`
	Notes        string       `json:"notes,omitempty" validate:"omitempty,max=500" jsonschema:"maxLength=500"`
	Assumptions  []Assumption `json:"assumptions,omitempty" validate:"omitempty,dive"`
}

type Assumption struct {
	Field  string `json:"field" validate:"required,startswith=/" jsonschema:"required,pattern=^/" jsonschema_description:"JSON path of the assumed field, e.g. /origin/city."`
	Value  string `json:"value" validate:"required,max=200" jsonschema:"required,maxLength=200"`
	Reason string `json:"reason" validate:"required,max=200" jsonschema:"required,maxLength=200"`
}

type Completeness struct {
	MissingRequired []string `json:"missing_required" jsonschema:"required" jsonschema_description:"JSON paths of required fields that are still empty."`
}

func enumOf(values ...string) *jsonschema.Schema {
	enum := make([]any, len(values))
	for i, v := range values {
		enum[i] = v
	}
	return &jsonschema.Schema{Type: "string", Enum: enum}
}

func setItems(s *jsonschema.Schema, prop string, values ...string) {
	if p, ok := s.Properties.Get(prop); ok {
		p.Items = enumOf(values...)
	}
}
