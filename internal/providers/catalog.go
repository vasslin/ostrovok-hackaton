package providers

// Spec ties a provider endpoint (as named in configs/providers.json) to the Go
// struct holding its request parameters.
type Spec struct {
	Provider string
	Endpoint string
	New      func() any // returns a pointer to a zero request struct
}

// Name is "<provider>.<endpoint>", used for schema and example file names.
func (s Spec) Name() string { return s.Provider + "." + s.Endpoint }

func spec[T any](provider, endpoint string) Spec {
	return Spec{Provider: provider, Endpoint: endpoint, New: func() any { return new(T) }}
}

// Catalog lists every endpoint we call. Keep it in sync with configs/providers.json
// (a test checks this).
var Catalog = []Spec{
	spec[AviasalesPricesForDates]("aviasales", "prices_for_dates"),
	spec[TravelpayoutsPlaces]("travelpayouts_places", "places2"),
	spec[RaspSearch]("yandex_rasp", "search"),
	spec[RaspNearestSettlement]("yandex_rasp", "nearest_settlement"),
	spec[RaspThread]("yandex_rasp", "thread"),
	spec[RaspCopyright]("yandex_rasp", "copyright"),
	spec[WikipediaGeosearch]("wikipedia", "geosearch"),
	spec[NominatimSearch]("nominatim", "search"),
	spec[NominatimAmenitySearch]("nominatim", "search_amenity"),
	spec[OSRMRoute]("osrm", "route"),
	spec[OpenMeteoForecast]("open_meteo", "forecast"),
	spec[OpenMeteoArchive]("open_meteo_archive", "archive"),
	spec[OpenMeteoMarine]("open_meteo_marine", "marine"),
	spec[CBRDaily]("cbr", "daily"),
}

// Lookup finds the spec for a provider endpoint.
func Lookup(provider, endpoint string) (Spec, bool) {
	for _, s := range Catalog {
		if s.Provider == provider && s.Endpoint == endpoint {
			return s, true
		}
	}
	return Spec{}, false
}
