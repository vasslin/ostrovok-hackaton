package providers

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/invopop/jsonschema"
)

// hotels_mock: our own service with the ETG / RateHawk B2B API v3 contract.
// All endpoints are POST with a JSON body. Limits below are copied from ETG docs.

// HotelsMulticomplete is POST /api/b2b/v3/search/multicomplete/: region and hotel
// autocomplete (at most 5 of each).
type HotelsMulticomplete struct {
	Query    string `json:"query" validate:"required,min=1,max=100" jsonschema:"required,minLength=1,maxLength=100"`
	Language string `json:"language,omitempty" validate:"omitempty,oneof=ru en" jsonschema:"enum=ru,enum=en"`
}

// HotelsGuests is one room.
type HotelsGuests struct {
	Adults   int        `json:"adults" validate:"min=1,max=6" jsonschema:"required,minimum=1,maximum=6"`
	Children []ChildAge `json:"children" validate:"max=4,dive,min=0,max=17" jsonschema:"required,maxItems=4" jsonschema_description:"Ages of children in this room."`
}

// ChildAge is a child's age in full years (0-17).
type ChildAge int

func (ChildAge) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{Type: "integer", Minimum: json.Number("0"), Maximum: json.Number("17")}
}

type HotelsFilter struct {
	StarRating  []int    `json:"star_rating,omitempty" validate:"omitempty,dive,min=0,max=5" jsonschema_description:"0 = no rating."`
	Kind        []string `json:"kind,omitempty" validate:"omitempty,dive,oneof=hotel hostel apartment guest_house resort"`
	MealType    string   `json:"meal_type,omitempty" validate:"omitempty,oneof=nomeal breakfast half-board full-board all-inclusive" jsonschema:"enum=nomeal,enum=breakfast,enum=half-board,enum=full-board,enum=all-inclusive"`
	PriceFrom   int      `json:"price_from,omitempty" validate:"omitempty,gt=0" jsonschema:"exclusiveMinimum=0" jsonschema_description:"Whole-stay price in the request currency."`
	PriceTo     int      `json:"price_to,omitempty" validate:"omitempty,gt=0,gtfield=PriceFrom" jsonschema:"exclusiveMinimum=0" jsonschema_description:"Whole-stay price in the request currency; greater than price_from."`
	HasTransfer *bool    `json:"has_transfer,omitempty"`
}

type UpsellTime struct {
	Time string `json:"time" validate:"required,datetime=15:04" jsonschema:"required,pattern=^([01][0-9]|2[0-3]):[0-5][0-9]$"`
}

type HotelsUpsells struct {
	EarlyCheckin *UpsellTime `json:"early_checkin,omitempty"`
	LateCheckout *UpsellTime `json:"late_checkout,omitempty"`
}

// HotelsSearchBase holds the fields shared by every search request.
type HotelsSearchBase struct {
	Checkin     string         `json:"checkin" validate:"required,datetime=2006-01-02" jsonschema:"required,format=date" jsonschema_description:"Not later than 730 days from today."`
	Checkout    string         `json:"checkout" validate:"required,datetime=2006-01-02" jsonschema:"required,format=date" jsonschema_description:"After checkin and not later than 30 days after it."`
	Residency   string         `json:"residency" validate:"required,len=2,alpha,lowercase" jsonschema:"required,pattern=^[a-z]{2}$" jsonschema_description:"Guests' citizenship, ISO 3166-1 alpha-2 lowercase."`
	Language    string         `json:"language,omitempty" validate:"omitempty,oneof=ru en" jsonschema:"enum=ru,enum=en"`
	Currency    string         `json:"currency,omitempty" validate:"omitempty,len=3,alpha,uppercase" jsonschema:"pattern=^[A-Z]{3}$"`
	Guests      []HotelsGuests `json:"guests" validate:"required,min=1,max=9,dive" jsonschema:"required,minItems=1,maxItems=9" jsonschema_description:"One object per room."`
	HotelsLimit int            `json:"hotels_limit,omitempty" validate:"omitempty,min=1,max=500" jsonschema:"minimum=1,maximum=500"`
	Filter      *HotelsFilter  `json:"filter,omitempty"`
	Upsells     *HotelsUpsells `json:"upsells,omitempty"`
}

// Check enforces the ETG date rules that tags can't express.
func (b HotelsSearchBase) Check() error {
	in, errIn := time.Parse(time.DateOnly, b.Checkin)
	out, errOut := time.Parse(time.DateOnly, b.Checkout)
	if errIn != nil || errOut != nil {
		return nil // reported by the tags
	}
	if !out.After(in) {
		return errors.New("checkout must be after checkin")
	}
	if out.Sub(in) > 30*24*time.Hour {
		return errors.New("checkout must be at most 30 days after checkin")
	}
	return nil
}

// HotelsSearchRegion is POST /api/b2b/v3/search/serp/region/.
type HotelsSearchRegion struct {
	HotelsSearchBase
	RegionID int `json:"region_id" validate:"required,min=1" jsonschema:"required,minimum=1" jsonschema_description:"From multicomplete."`
}

// HotelsSearchGeo is POST /api/b2b/v3/search/serp/geo/.
type HotelsSearchGeo struct {
	HotelsSearchBase
	Latitude  float64 `json:"latitude" validate:"min=-90,max=90" jsonschema:"required,minimum=-90,maximum=90"`
	Longitude float64 `json:"longitude" validate:"min=-180,max=180" jsonschema:"required,minimum=-180,maximum=180"`
	Radius    int     `json:"radius" validate:"required,min=1,max=70000" jsonschema:"required,minimum=1,maximum=70000" jsonschema_description:"Meters (our cap)."`
}

// HotelsSearchHotels is POST /api/b2b/v3/search/serp/hotels/.
type HotelsSearchHotels struct {
	HotelsSearchBase
	Hids []int64 `json:"hids" validate:"required,min=1,max=300,dive,min=1,max=9999999999" jsonschema:"required,minItems=1,maxItems=300"`
}

// HotelsHotelpage is POST /api/b2b/v3/search/hp/: every rate of one hotel.
type HotelsHotelpage struct {
	HotelsSearchBase
	Hid int64 `json:"hid" validate:"required,min=1,max=9999999999" jsonschema:"required,minimum=1,maximum=9999999999" jsonschema_description:"Numeric hotel id, at most 10 digits."`
}

// HotelsPrebook is POST /api/b2b/v3/hotel/prebook/: re-checks the rate; a price
// change must be shown to the user before booking.
type HotelsPrebook struct {
	Hash                 string `json:"hash" validate:"required,min=1,max=256" jsonschema:"required,minLength=1,maxLength=256" jsonschema_description:"book_hash from the hotel page; valid 24 hours."`
	PriceIncreasePercent *int   `json:"price_increase_percent,omitempty" validate:"omitempty,min=0,max=100" jsonschema:"minimum=0,maximum=100"`
}

// HotelsBookingForm is POST /api/b2b/v3/hotel/order/booking/form/.
type HotelsBookingForm struct {
	PartnerOrderID string `json:"partner_order_id" validate:"required,uuid" jsonschema:"required,format=uuid" jsonschema_description:"Our booking id; unique per booking attempt."`
	BookHash       string `json:"book_hash" validate:"required,min=1,max=256" jsonschema:"required,minLength=1,maxLength=256"`
	Language       string `json:"language" validate:"required,oneof=ru en" jsonschema:"required,enum=ru,enum=en"`
	UserIP         string `json:"user_ip" validate:"required,ip" jsonschema:"required" jsonschema_description:"End user's IPv4 or IPv6 address."`
}

type BookingUser struct {
	Email   string `json:"email" validate:"required,email" jsonschema:"required,format=email"`
	Phone   string `json:"phone,omitempty" validate:"omitempty,max=32" jsonschema:"maxLength=32"`
	Comment string `json:"comment,omitempty" validate:"omitempty,min=1,max=256" jsonschema:"minLength=1,maxLength=256"`
}

type BookingPartner struct {
	PartnerOrderID string `json:"partner_order_id" validate:"required,uuid" jsonschema:"required,format=uuid"`
}

type BookingGuest struct {
	FirstName string `json:"first_name" validate:"required,min=1,max=64" jsonschema:"required,minLength=1,maxLength=64"`
	LastName  string `json:"last_name" validate:"required,min=1,max=64" jsonschema:"required,minLength=1,maxLength=64"`
	IsChild   bool   `json:"is_child,omitempty"`
	Age       *int   `json:"age,omitempty" validate:"required_if=IsChild true,omitempty,min=0,max=17" jsonschema:"minimum=0,maximum=17" jsonschema_description:"Required for children."`
}

type BookingRoom struct {
	Guests []BookingGuest `json:"guests" validate:"required,min=1,max=10,dive" jsonschema:"required,minItems=1,maxItems=10"`
}

type BookingPayment struct {
	Type         string `json:"type" validate:"required,oneof=now hotel deposit" jsonschema:"required,enum=now,enum=hotel,enum=deposit"`
	Amount       string `json:"amount" validate:"required,numeric" jsonschema:"required,pattern=^[0-9]+(\\.[0-9]+)?$" jsonschema_description:"Decimal string, as in ETG."`
	CurrencyCode string `json:"currency_code" validate:"required,len=3,alpha,uppercase" jsonschema:"required,pattern=^[A-Z]{3}$"`
}

// HotelsBookingFinish is POST /api/b2b/v3/hotel/order/booking/finish/: starts the
// async booking; poll booking_status afterwards.
type HotelsBookingFinish struct {
	User        BookingUser    `json:"user" jsonschema:"required"`
	Partner     BookingPartner `json:"partner" jsonschema:"required"`
	Language    string         `json:"language" validate:"required,oneof=ru en" jsonschema:"required,enum=ru,enum=en"`
	Rooms       []BookingRoom  `json:"rooms" validate:"required,min=1,max=9,dive" jsonschema:"required,minItems=1,maxItems=9"`
	PaymentType BookingPayment `json:"payment_type" jsonschema:"required"`
}

// HotelsBookingStatus is POST /api/b2b/v3/hotel/order/booking/finish/status/.
type HotelsBookingStatus struct {
	PartnerOrderID string `json:"partner_order_id" validate:"required,uuid" jsonschema:"required,format=uuid"`
}

// HotelsCancel is POST /api/b2b/v3/hotel/order/cancel/.
type HotelsCancel struct {
	PartnerOrderID string `json:"partner_order_id" validate:"required,uuid" jsonschema:"required,format=uuid"`
}
