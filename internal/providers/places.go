package providers

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// WikipediaGeosearch is GET https://ru.wikipedia.org/w/api.php with
// generator=geosearch: articles near a point, with coordinates, a short
// description, a photo and a link in one call. Skip pages without a
// description: those are usually areas (e.g. «Казанская агломерация»), not places.
type WikipediaGeosearch struct {
	Action        string `json:"action" url:"action" validate:"required,eq=query" jsonschema:"required,enum=query"`
	Generator     string `json:"generator" url:"generator" validate:"required,eq=geosearch" jsonschema:"required,enum=geosearch"`
	Coord         string `json:"ggscoord" url:"ggscoord" validate:"required,max=40,contains=0x7C" jsonschema:"required,pattern=^-?[0-9.]+\\|-?[0-9.]+$" jsonschema_description:"lat|lon, e.g. 55.7987|49.1059."`
	Radius        int    `json:"ggsradius" url:"ggsradius" validate:"min=10,max=10000" jsonschema:"required,minimum=10,maximum=10000" jsonschema_description:"Meters (MediaWiki allows 10-10000)."`
	Limit         int    `json:"ggslimit,omitempty" url:"ggslimit,omitempty" validate:"omitempty,min=1,max=50" jsonschema:"minimum=1,maximum=50,default=20" jsonschema_description:"Our cap of 50 keeps a photo for every page (pageimages returns at most 50)."`
	Prop          string `json:"prop,omitempty" url:"prop,omitempty" validate:"omitempty,max=100" jsonschema:"maxLength=100,default=coordinates|pageimages|description|info"`
	InProp        string `json:"inprop,omitempty" url:"inprop,omitempty" validate:"omitempty,eq=url" jsonschema:"enum=url" jsonschema_description:"url adds fullurl (link to the article)."`
	PIProp        string `json:"piprop,omitempty" url:"piprop,omitempty" validate:"omitempty,oneof=thumbnail original" jsonschema:"enum=thumbnail,enum=original"`
	PIThumbSize   int    `json:"pithumbsize,omitempty" url:"pithumbsize,omitempty" validate:"omitempty,min=50,max=1000" jsonschema:"minimum=50,maximum=1000" jsonschema_description:"Thumbnail width in px."`
	Format        string `json:"format" url:"format" validate:"required,eq=json" jsonschema:"required,enum=json"`
	FormatVersion int    `json:"formatversion,omitempty" url:"formatversion,omitempty" validate:"omitempty,oneof=1 2" jsonschema:"enum=1,enum=2,default=2" jsonschema_description:"2 returns pages as a list."`
}

func (r WikipediaGeosearch) Check() error {
	parts := strings.Split(r.Coord, "|")
	if len(parts) != 2 {
		return fmt.Errorf("ggscoord must be lat|lon")
	}
	lat, errLat := strconv.ParseFloat(parts[0], 64)
	lon, errLon := strconv.ParseFloat(parts[1], 64)
	if errLat != nil || errLon != nil || !validCoordinates(lat, lon) {
		return fmt.Errorf("ggscoord is outside valid latitude/longitude range")
	}
	return nil
}

// NominatimSearch is GET /search: geocoding. Nominatim allows at most one
// request per second and requires an identifying User-Agent.
type NominatimSearch struct {
	Q              string `json:"q" url:"q" validate:"required,min=1,max=255" jsonschema:"required,minLength=1,maxLength=255"`
	Format         string `json:"format" url:"format" validate:"required,oneof=json jsonv2 geojson" jsonschema:"required,enum=json,enum=jsonv2,enum=geojson"`
	Limit          int    `json:"limit,omitempty" url:"limit,omitempty" validate:"omitempty,min=1,max=40" jsonschema:"minimum=1,maximum=40"`
	AcceptLanguage string `json:"accept-language,omitempty" url:"accept-language,omitempty" validate:"omitempty,max=20" jsonschema:"maxLength=20"`
	CountryCodes   string `json:"countrycodes,omitempty" url:"countrycodes,omitempty" validate:"omitempty,max=60" jsonschema:"maxLength=60" jsonschema_description:"Comma-separated ISO alpha-2 codes, lowercase."`
}

// NominatimAmenitySearch is GET /search with structured parameters: places of
// one type in a city (museums, theatres, cinemas that Wikipedia often lacks).
// Results can include stops named after a place (type bus_stop): keep only the
// types you asked for. Coverage is partial (few cinemas), at most 40 results.
type NominatimAmenitySearch struct {
	Amenity        string `json:"amenity" url:"amenity" validate:"required,min=1,max=50" jsonschema:"required,minLength=1,maxLength=50" jsonschema_description:"Place type, e.g. museum, theatre, cinema, zoo, park, restaurant."`
	City           string `json:"city" url:"city" validate:"required,min=1,max=100" jsonschema:"required,minLength=1,maxLength=100"`
	Country        string `json:"country,omitempty" url:"country,omitempty" validate:"omitempty,max=100" jsonschema:"maxLength=100"`
	Format         string `json:"format" url:"format" validate:"required,oneof=json jsonv2 geojson" jsonschema:"required,enum=json,enum=jsonv2,enum=geojson"`
	Limit          int    `json:"limit,omitempty" url:"limit,omitempty" validate:"omitempty,min=1,max=40" jsonschema:"minimum=1,maximum=40"`
	AcceptLanguage string `json:"accept-language,omitempty" url:"accept-language,omitempty" validate:"omitempty,max=20" jsonschema:"maxLength=20"`
}

// OSRMRoute is GET /route/v1/driving/{coordinates}: drive time and distance.
// The public demo server is for light use only, so cache the results.
type OSRMRoute struct {
	Coordinates string `json:"coordinates" url:"-" path:"coordinates" validate:"required,max=200" jsonschema:"required,maxLength=200" jsonschema_description:"lon,lat;lon,lat - longitude first."`
	Overview    string `json:"overview,omitempty" url:"overview,omitempty" validate:"omitempty,oneof=false simplified full" jsonschema:"enum=false,enum=simplified,enum=full"`
}

func (r OSRMRoute) Check() error {
	points := strings.Split(r.Coordinates, ";")
	if len(points) < 2 || len(points) > 10 {
		return fmt.Errorf("coordinates must contain 2 to 10 lon,lat points")
	}
	for _, point := range points {
		parts := strings.Split(point, ",")
		if len(parts) != 2 {
			return fmt.Errorf("coordinates must contain lon,lat points")
		}
		lon, errLon := strconv.ParseFloat(parts[0], 64)
		lat, errLat := strconv.ParseFloat(parts[1], 64)
		if errLon != nil || errLat != nil || !validCoordinates(lat, lon) {
			return fmt.Errorf("coordinates contain invalid latitude/longitude")
		}
	}
	return nil
}

func validCoordinates(lat, lon float64) bool {
	return !math.IsNaN(lat) && !math.IsInf(lat, 0) && !math.IsNaN(lon) && !math.IsInf(lon, 0) && lat >= -90 && lat <= 90 && lon >= -180 && lon <= 180
}

// OpenMeteoForecast is GET https://api.open-meteo.com/v1/forecast (≤16 days ahead).
type OpenMeteoForecast struct {
	Latitude     float64 `json:"latitude" url:"latitude" validate:"min=-90,max=90" jsonschema:"required,minimum=-90,maximum=90"`
	Longitude    float64 `json:"longitude" url:"longitude" validate:"min=-180,max=180" jsonschema:"required,minimum=-180,maximum=180"`
	Daily        string  `json:"daily,omitempty" url:"daily,omitempty" validate:"omitempty,max=300" jsonschema:"maxLength=300" jsonschema_description:"Comma-separated, e.g. temperature_2m_max,temperature_2m_min,precipitation_sum."`
	Hourly       string  `json:"hourly,omitempty" url:"hourly,omitempty" validate:"omitempty,max=300" jsonschema:"maxLength=300"`
	Timezone     string  `json:"timezone,omitempty" url:"timezone,omitempty" validate:"omitempty,max=64" jsonschema:"maxLength=64,default=auto"`
	ForecastDays int     `json:"forecast_days,omitempty" url:"forecast_days,omitempty" validate:"omitempty,min=1,max=16" jsonschema:"minimum=1,maximum=16"`
	StartDate    string  `json:"start_date,omitempty" url:"start_date,omitempty" validate:"omitempty,datetime=2006-01-02" jsonschema:"format=date"`
	EndDate      string  `json:"end_date,omitempty" url:"end_date,omitempty" validate:"omitempty,datetime=2006-01-02" jsonschema:"format=date"`
}

// OpenMeteoArchive is GET https://archive-api.open-meteo.com/v1/archive (1940 onwards).
type OpenMeteoArchive struct {
	Latitude  float64 `json:"latitude" url:"latitude" validate:"min=-90,max=90" jsonschema:"required,minimum=-90,maximum=90"`
	Longitude float64 `json:"longitude" url:"longitude" validate:"min=-180,max=180" jsonschema:"required,minimum=-180,maximum=180"`
	StartDate string  `json:"start_date" url:"start_date" validate:"required,datetime=2006-01-02" jsonschema:"required,format=date"`
	EndDate   string  `json:"end_date" url:"end_date" validate:"required,datetime=2006-01-02" jsonschema:"required,format=date"`
	Daily     string  `json:"daily,omitempty" url:"daily,omitempty" validate:"omitempty,max=300" jsonschema:"maxLength=300"`
	Hourly    string  `json:"hourly,omitempty" url:"hourly,omitempty" validate:"omitempty,max=300" jsonschema:"maxLength=300"`
	Timezone  string  `json:"timezone,omitempty" url:"timezone,omitempty" validate:"omitempty,max=64" jsonschema:"maxLength=64,default=auto"`
}

// OpenMeteoMarine is GET https://marine-api.open-meteo.com/v1/marine
// (sea_surface_temperature; forecast 0-8 days, history from 2022).
type OpenMeteoMarine struct {
	Latitude     float64 `json:"latitude" url:"latitude" validate:"min=-90,max=90" jsonschema:"required,minimum=-90,maximum=90"`
	Longitude    float64 `json:"longitude" url:"longitude" validate:"min=-180,max=180" jsonschema:"required,minimum=-180,maximum=180"`
	Hourly       string  `json:"hourly,omitempty" url:"hourly,omitempty" validate:"omitempty,max=300" jsonschema:"maxLength=300"`
	Daily        string  `json:"daily,omitempty" url:"daily,omitempty" validate:"omitempty,max=300" jsonschema:"maxLength=300"`
	ForecastDays int     `json:"forecast_days,omitempty" url:"forecast_days,omitempty" validate:"omitempty,min=0,max=8" jsonschema:"minimum=0,maximum=8"`
	StartDate    string  `json:"start_date,omitempty" url:"start_date,omitempty" validate:"omitempty,datetime=2006-01-02" jsonschema:"format=date"`
	EndDate      string  `json:"end_date,omitempty" url:"end_date,omitempty" validate:"omitempty,datetime=2006-01-02" jsonschema:"format=date"`
	Timezone     string  `json:"timezone,omitempty" url:"timezone,omitempty" validate:"omitempty,max=64" jsonschema:"maxLength=64"`
}

// CBRDaily is GET https://www.cbr.ru/scripts/XML_daily.asp: daily exchange
// rates as windows-1251 XML.
type CBRDaily struct {
	DateReq string `json:"date_req,omitempty" url:"date_req,omitempty" validate:"omitempty,datetime=02/01/2006" jsonschema:"pattern=^[0-9]{2}/[0-9]{2}/[0-9]{4}$" jsonschema_description:"DD/MM/YYYY; omit for the last registered date."`
}
