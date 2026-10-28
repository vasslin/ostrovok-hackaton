package providerkit

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/ostrovok-hackathon-2026/godnota/internal/providers"
)

const transferJSON = `{"segments":[{"departure_from":{"code":"s10"},"arrival_to":{"code":"s30"},"departure":"2026-11-28T11:00:00+03:00","arrival":"2026-11-28T18:00:00+04:00","has_transfers":true,"details":[{"thread":{"uid":"a","transport_type":"plane"},"departure":"2026-11-28T11:00:00+03:00","arrival":"2026-11-28T12:00:00+03:00"},{"is_transfer":true,"transfer_from":{"code":"s20"},"transfer_to":{"code":"s20"}},{"thread":{"uid":"b","transport_type":"plane"},"departure":"2026-11-28T14:00:00+03:00","arrival":"2026-11-28T18:00:00+04:00"}]}]}`

func TestYandexTransferEnvelope(t *testing.T) {
	q := &providers.RaspSearch{From: "c43", To: "c51", Date: "2026-11-28", TransportTypes: "plane", Transfers: true}
	job := Request{Provider: "yandex_rasp", Params: q}
	result, err := normalize(job, []byte(transferJSON), time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC))
	if err != nil || len(result) != 1 || result[0].Transfers != 1 || len(result[0].Legs) != 2 || result[0].Legs[0].DestinationStation != "s20" || result[0].Legs[1].OriginStation != "s20" || result[0].Price != nil {
		t.Fatalf("invalid transfer %+v %v", result, err)
	}
	q.Transfers = false
	result, err = normalize(job, []byte(transferJSON), time.Now())
	if err != nil || len(result) != 0 {
		t.Fatal("direct-only returned transfer", err)
	}
	q.Transfers = true
	for _, broken := range []string{strings.Replace(transferJSON, `"is_transfer":true`, `"is_transfer":false`, 1), strings.Replace(transferJSON, `"code":"s20"`, `"code":"c20"`, 1), strings.Replace(transferJSON, `"uid":"b"`, `"uid":""`, 1), strings.Replace(transferJSON, `"transport_type":"plane"`, `"transport_type":"train"`, 1)} {
		result, err = normalize(job, []byte(broken), time.Now())
		if err != nil || len(result) != 0 {
			t.Fatal("malformed transfer accepted", err)
		}
	}
	var segment yandexSegment
	if err := json.Unmarshal([]byte(`{"has_transfers":true,"details":[]}`), &segment); err != nil {
		t.Fatal(err)
	}
	if _, ok := normalizeYandexTransfer(segment, "plane", "2026-11-28", time.Now()); ok {
		t.Fatal("missing legs accepted")
	}
}
