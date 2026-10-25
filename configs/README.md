# configs/

Configs for the travel providers. The full plan is in [docs/providers/PLAN.md](../docs/providers/PLAN.md).

| Path | What | Edited by |
|---|---|---|
| `providers.json` | Registry: base URL, auth (env var names only), rate limit, timeout, cache TTL, endpoints (method, path, `query` or `json`) for every provider | hand |
| `trip_request.schema.json` | Schema of `TripRequest`, the one config the AI agent fills (human values only, no provider codes) | **generated** |
| `providers/<provider>.<endpoint>.schema.json` | Schema of the request parameters for one provider endpoint | **generated** |
| `examples/trip_request.*.json` | 5 filled trip configs (Kazan weekend, vague sea trip, SPb→Moscow train, Kizhi by boat, family with a dog) | hand |
| `examples/<provider>.<endpoint>.json` | One request config per endpoint, mostly built from the Kazan trip | hand |

## Where the rules live

Required fields and format limits are defined **once**, in Go struct tags:

- `internal/trip/trip.go` — `TripRequest`
- `internal/providers/{tickets,hotels,places}.go` — one struct per endpoint

```go
Origin string `json:"origin" url:"origin" validate:"required,len=3,alpha,uppercase" jsonschema:"required,pattern=^[A-Z]{3}$"`
```

- `validate` — checked in Go at runtime (`providers.Validate`, `TripRequest.ValidateAt`)
- `jsonschema` — copied into the generated `*.schema.json` (give these to the LLM and the frontend)
- `url` / `json` — how the field is sent (query string or JSON body); `path:"name"` fills `{name}` in the endpoint path

Rules tags can't express (return after departure, at most 30 nights, not in the past) live in `TripRequest.ValidateAt` and `HotelsSearchBase.Check`.

## Commands

Needs Go 1.26+ (`brew install go`; an older Go downloads 1.26 by itself).

```bash
go generate ./...
```

```bash
go test ./...
```

Send every example to the real APIs (keys from `.env`, see `.env.example`) and save the responses to `testdata/`:

```bash
go run ./cmd/smoke
```

Run `go generate ./...` after changing any struct; `go test` fails if a schema is out of date, if an example breaks its schema or tags, or if `providers.json` and the Go catalog (`internal/providers/catalog.go`) disagree.

## Adding an endpoint

1. Add a request struct with tags in `internal/providers/`.
2. Add it to `Catalog` in `internal/providers/catalog.go`.
3. Add the endpoint to `providers.json`.
4. Add `examples/<provider>.<endpoint>.json`.
5. Run `go generate ./...` and `go test ./...`.

## Using it in code

```go
reg, _ := providers.LoadRegistry("configs/providers.json")
req, err := reg.NewHTTPRequest(ctx, providers.Request{
    Provider: "aviasales", Endpoint: "prices_for_dates",
    Params: &providers.AviasalesPricesForDates{Origin: "MOW", Destination: "KZN", DepartureAt: "2026-12-12", Currency: "rub"},
})
// req is a ready *http.Request: validated params, URL, query/body, auth from env vars
```

Secrets come from env vars named in `providers.json`: `TRAVELPAYOUTS_TOKEN`, `YANDEX_RASP_API_KEY`, `HOTELS_KEY_ID`, `HOTELS_KEY`.
