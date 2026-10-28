package providers

// AviasalesPricesForDates is GET /aviasales/v3/prices_for_dates: cached flight
// prices Aviasales users found recently (not live availability).
type AviasalesPricesForDates struct {
	Origin      string `json:"origin" url:"origin" validate:"required,len=3,alpha,uppercase" jsonschema:"required,pattern=^[A-Z]{3}$" jsonschema_description:"IATA code of the origin city or airport, e.g. MOW."`
	Destination string `json:"destination" url:"destination" validate:"required,len=3,alpha,uppercase" jsonschema:"required,pattern=^[A-Z]{3}$" jsonschema_description:"IATA code of the destination city or airport, e.g. KZN."`
	DepartureAt string `json:"departure_at,omitempty" url:"departure_at,omitempty" validate:"omitempty,datetime=2006-01-02|datetime=2006-01" jsonschema:"pattern=^[0-9]{4}-[0-9]{2}(-[0-9]{2})?$" jsonschema_description:"YYYY-MM-DD or YYYY-MM."`
	ReturnAt    string `json:"return_at,omitempty" url:"return_at,omitempty" validate:"omitempty,datetime=2006-01-02|datetime=2006-01" jsonschema:"pattern=^[0-9]{4}-[0-9]{2}(-[0-9]{2})?$" jsonschema_description:"YYYY-MM-DD or YYYY-MM; omit for one-way."`
	OneWay      bool   `json:"one_way" url:"one_way" jsonschema_description:"true returns one ticket per date group."`
	Direct      bool   `json:"direct" url:"direct" jsonschema_description:"Non-stop flights only."`
	Currency    string `json:"currency" url:"currency" validate:"required,len=3,alpha,lowercase" jsonschema:"required,pattern=^[a-z]{3}$,default=rub"`
	Sorting     string `json:"sorting,omitempty" url:"sorting,omitempty" validate:"omitempty,oneof=price route" jsonschema:"enum=price,enum=route,default=price"`
	Unique      bool   `json:"unique,omitempty" url:"unique,omitempty"`
	Limit       int    `json:"limit,omitempty" url:"limit,omitempty" validate:"omitempty,min=1,max=100" jsonschema:"minimum=1,maximum=100,default=30" jsonschema_description:"Our cap; the official maximum is not confirmed."`
	Page        int    `json:"page,omitempty" url:"page,omitempty" validate:"omitempty,min=1" jsonschema:"minimum=1,default=1"`
}

// TravelpayoutsPlaces is GET https://autocomplete.travelpayouts.com/places2:
// city or airport name -> IATA code, coordinates, Russian word forms.
type TravelpayoutsPlaces struct {
	Term   string   `json:"term" url:"term" validate:"required,min=1,max=100" jsonschema:"required,minLength=1,maxLength=100"`
	Locale string   `json:"locale" url:"locale" validate:"required,oneof=ru en" jsonschema:"required,enum=ru,enum=en"`
	Types  []string `json:"types" url:"types[]" validate:"required,min=1,dive,oneof=city airport country" jsonschema:"required,minItems=1,uniqueItems=true"`
}

// RaspSearch is GET /v3.0/search/: threads between two points
// (plane, train, suburban, bus, water, helicopter).
type RaspSearch struct {
	From           string `json:"from" url:"from" validate:"required,min=2,max=20" jsonschema:"required,minLength=2,maxLength=20" jsonschema_description:"Yandex code: c213 (city) or s9600213 (station); other systems via system."`
	To             string `json:"to" url:"to" validate:"required,min=2,max=20" jsonschema:"required,minLength=2,maxLength=20"`
	Date           string `json:"date,omitempty" url:"date,omitempty" validate:"omitempty,datetime=2006-01-02" jsonschema:"format=date" jsonschema_description:"Without a date only direct threads for all dates are returned."`
	TransportTypes string `json:"transport_types,omitempty" url:"transport_types,omitempty" validate:"omitempty,oneof=plane train suburban bus water helicopter" jsonschema:"enum=plane,enum=train,enum=suburban,enum=bus,enum=water,enum=helicopter" jsonschema_description:"Omit for all types."`
	System         string `json:"system,omitempty" url:"system,omitempty" validate:"omitempty,oneof=yandex iata sirena express esr" jsonschema:"enum=yandex,enum=iata,enum=sirena,enum=express,enum=esr,default=yandex"`
	Transfers      bool   `json:"transfers" url:"transfers" jsonschema_description:"Add routes with transfers."`
	Limit          int    `json:"limit,omitempty" url:"limit,omitempty" validate:"omitempty,min=1,max=100" jsonschema:"minimum=1,maximum=100,default=100"`
	Offset         int    `json:"offset,omitempty" url:"offset,omitempty" validate:"omitempty,min=0" jsonschema:"minimum=0,default=0"`
	Lang           string `json:"lang,omitempty" url:"lang,omitempty" validate:"omitempty,oneof=ru_RU uk_UA" jsonschema:"enum=ru_RU,enum=uk_UA,default=ru_RU"`
	Format         string `json:"format,omitempty" url:"format,omitempty" validate:"omitempty,oneof=json xml" jsonschema:"enum=json,enum=xml,default=json"`
	ResultTimezone string `json:"result_timezone,omitempty" url:"result_timezone,omitempty" validate:"omitempty,max=64" jsonschema:"maxLength=64" jsonschema_description:"IANA time zone; default is each station's own zone."`
}

// RaspNearestSettlement is GET /v3.0/nearest_settlement/: city code for coordinates.
type RaspNearestSettlement struct {
	Lat      float64 `json:"lat" url:"lat" validate:"min=-90,max=90" jsonschema:"required,minimum=-90,maximum=90"`
	Lng      float64 `json:"lng" url:"lng" validate:"min=-180,max=180" jsonschema:"required,minimum=-180,maximum=180"`
	Distance int     `json:"distance,omitempty" url:"distance,omitempty" validate:"omitempty,min=1,max=50" jsonschema:"minimum=1,maximum=50" jsonschema_description:"Search radius in km (our cap)."`
	Lang     string  `json:"lang,omitempty" url:"lang,omitempty" validate:"omitempty,oneof=ru_RU uk_UA" jsonschema:"enum=ru_RU,enum=uk_UA"`
	Format   string  `json:"format,omitempty" url:"format,omitempty" validate:"omitempty,oneof=json xml" jsonschema:"enum=json,enum=xml"`
}

// RaspThread is GET /v3.0/thread/: stops of one route.
type RaspThread struct {
	UID    string `json:"uid" url:"uid" validate:"required,max=100" jsonschema:"required,minLength=1,maxLength=100"`
	Date   string `json:"date,omitempty" url:"date,omitempty" validate:"omitempty,datetime=2006-01-02" jsonschema:"format=date"`
	Lang   string `json:"lang,omitempty" url:"lang,omitempty" validate:"omitempty,oneof=ru_RU uk_UA" jsonschema:"enum=ru_RU,enum=uk_UA"`
	Format string `json:"format,omitempty" url:"format,omitempty" validate:"omitempty,oneof=json xml" jsonschema:"enum=json,enum=xml"`
}

// RaspCopyright is GET /v3.0/copyright/: the attribution block we must show.
type RaspCopyright struct {
	Format string `json:"format,omitempty" url:"format,omitempty" validate:"omitempty,oneof=json xml" jsonschema:"enum=json,enum=xml"`
}
