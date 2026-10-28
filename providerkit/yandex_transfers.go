package providerkit

import (
	"strconv"
	"strings"
	"time"
)

type yandexPlace struct {
	Code string `json:"code"`
}
type yandexThread struct {
	UID           string `json:"uid"`
	TransportType string `json:"transport_type"`
}
type yandexDetail struct {
	IsTransfer   bool         `json:"is_transfer"`
	Thread       yandexThread `json:"thread"`
	TransferFrom yandexPlace  `json:"transfer_from"`
	TransferTo   yandexPlace  `json:"transfer_to"`
	Departure    string       `json:"departure"`
	Arrival      string       `json:"arrival"`
}
type yandexSegment struct {
	Departure     string         `json:"departure"`
	Arrival       string         `json:"arrival"`
	HasTransfers  bool           `json:"has_transfers"`
	From          yandexPlace    `json:"from"`
	To            yandexPlace    `json:"to"`
	DepartureFrom yandexPlace    `json:"departure_from"`
	ArrivalTo     yandexPlace    `json:"arrival_to"`
	Thread        yandexThread   `json:"thread"`
	Details       []yandexDetail `json:"details"`
}

func yandexStation(code string) bool {
	if len(code) < 2 || code[0] != 's' {
		return false
	}
	_, err := strconv.ParseUint(code[1:], 10, 64)
	return err == nil
}
func normalizeYandexTransfer(item yandexSegment, mode, date string, now time.Time) (TransportPart, bool) {
	if len(item.Details) < 3 || len(item.Details) > 7 || len(item.Details)%2 == 0 || !strings.HasPrefix(item.Departure, date) || !yandexStation(item.DepartureFrom.Code) || !yandexStation(item.ArrivalTo.Code) {
		return TransportPart{}, false
	}
	legs := make([]TransferLeg, 0, (len(item.Details)+1)/2)
	ids := []string{"yandex_rasp", item.Departure, item.Arrival}
	origin := item.DepartureFrom.Code
	for i := 0; i < len(item.Details); i += 2 {
		detail := item.Details[i]
		if detail.IsTransfer || detail.Thread.UID == "" || detail.Thread.TransportType != mode {
			return TransportPart{}, false
		}
		dep, e1 := time.Parse(time.RFC3339, detail.Departure)
		arr, e2 := time.Parse(time.RFC3339, detail.Arrival)
		if e1 != nil || e2 != nil || !arr.After(dep) {
			return TransportPart{}, false
		}
		destination := item.ArrivalTo.Code
		if i+1 < len(item.Details) {
			connection := item.Details[i+1]
			if !connection.IsTransfer || !yandexStation(connection.TransferFrom.Code) || !yandexStation(connection.TransferTo.Code) {
				return TransportPart{}, false
			}
			destination = connection.TransferFrom.Code
		}
		legs = append(legs, TransferLeg{OriginStation: origin, DestinationStation: destination, DepartureAt: detail.Departure, ArrivalAt: detail.Arrival, Mode: mode, ServiceID: detail.Thread.UID})
		ids = append(ids, detail.Thread.UID, detail.Departure, detail.Arrival)
		if i+1 < len(item.Details) {
			origin = item.Details[i+1].TransferTo.Code
		}
	}
	first, e1 := time.Parse(time.RFC3339, item.Departure)
	last, e2 := time.Parse(time.RFC3339, item.Arrival)
	legFirst, _ := time.Parse(time.RFC3339, legs[0].DepartureAt)
	legLast, _ := time.Parse(time.RFC3339, legs[len(legs)-1].ArrivalAt)
	if e1 != nil || e2 != nil || !first.Equal(legFirst) || !last.Equal(legLast) {
		return TransportPart{}, false
	}
	return TransportPart{ID: id(ids...), Provider: "yandex_rasp", Source: "live", FetchedAt: now.UTC().Format(time.RFC3339), Mode: mode, Origin: item.DepartureFrom.Code, Destination: item.ArrivalTo.Code, DepartureAt: item.Departure, ArrivalAt: item.Arrival, Transfers: len(legs) - 1, Legs: legs, Warnings: []string{"Schedule only: price and availability unknown", "Данные: Яндекс.Расписания - https://rasp.yandex.ru/"}}, true
}
