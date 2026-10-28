package service

type SearchPartsResponse struct {
	TransportParts []*TransportPart `json:"transport_parts"`
	HotelParts     []*HotelPart     `json:"hotel_parts"`
	Errors         []*ProviderError `json:"errors"`
	Limitations    []string         `json:"limitations"`
}
type Money struct {
	AmountMinor Minor  `json:"amount_minor"`
	Currency    string `json:"currency"`
	Scope       string `json:"scope"`
}
type TransportPart struct {
	ID                string   `json:"id"`
	Provider          string   `json:"provider"`
	Source            string   `json:"source"`
	FetchedAt         string   `json:"fetched_at"`
	Mode              string   `json:"mode"`
	Origin            string   `json:"origin"`
	Destination       string   `json:"destination"`
	DepartureAt       string   `json:"departure_at"`
	ArrivalAt         string   `json:"arrival_at"`
	Price             *Money   `json:"price"`
	BookingURL        string   `json:"booking_url"`
	Warnings          []string `json:"warnings"`
	ReturnDepartureAt string   `json:"return_departure_at"`
	ReturnArrivalAt   string   `json:"return_arrival_at"`
	OriginCity        string   `json:"origin_city"`
	DestinationCity   string   `json:"destination_city"`
	Transfers         int      `json:"transfers"`
	ReturnTransfers   int      `json:"return_transfers"`
}
type HotelPart struct {
	ID               string   `json:"id"`
	Provider         string   `json:"provider"`
	Source           string   `json:"source"`
	FetchedAt        string   `json:"fetched_at"`
	City             string   `json:"city"`
	Name             string   `json:"name"`
	CheckIn          string   `json:"check_in"`
	CheckOut         string   `json:"check_out"`
	Stars            int      `json:"stars"`
	Price            *Money   `json:"price"`
	FreeCancellation bool     `json:"free_cancellation"`
	Warnings         []string `json:"warnings"`
}
type ProviderError struct {
	Provider  string `json:"provider"`
	Kind      string `json:"kind"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
}
