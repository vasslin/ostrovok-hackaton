# Provider integration

The provider branch was merged into `main` at `0fcf00f` and stage-two integration at `91c8ca8`. The root module is in `go.work`; `.env` is ignored and no longer tracked, while `.env.example` lists the configuration names.

## Service contract

`api/provider/v1/provider.proto` exposes the stage-two `SearchParts` method and the separate `SearchTransport`, `SearchHotels` and `SearchActivities` methods. The latter three accept `SearchInput` with:

- `trip_request`: an object in the unified `TripRequest` shape, or the chosen adapter's input shape.
- `input_format`: `trip_request` (default), `ai_config`, or `frontend_criteria`.
- `resolved`: independently verified origin/destination provider codes and coordinates. Hotel search additionally requires confirmed `hotel_checkin` and `hotel_checkout`; Hotelbeds requires an explicit `hotelbeds_source_market`.

The result contains offers, provider issues, missing fields, limitations and discovery requests. A missing optional `price_minor` means unknown, never free. Protobuf JSON represents int64 values as strings; a gateway must convert them safely before sending the frontend's numeric minor-unit fields. Quote scope, currency, source, retrieval time and synthetic/sandbox status must travel with the offer.

The AI branch's flat config does not specify budget scope or structured preferences; the frontend criteria do not identify pet species, hotel room allocation or source market. The adapters return unresolved fields instead of guessing. The existing frontend expects a complete `Trip` and cannot safely render partial provider offers or unknown totals without a gateway/UI change.

## Running the provider service

From the repository root, after supplying any credentials in a local environment:

```sh
go run ./provider-service/cmd/app
go run ./provider-service/cmd/search -kind activity -resolved configs/examples/resolved_trip.kazan_weekend.json
```

The service reads `PROVIDERS_CONFIG` (or `PROVIDER_CONFIG_FILE`, default `configs/providers.json`) and `POI_INDEX_PATH` (default `configs/poi/kazan.json`). Docker Compose passes the ignored local `.env` to the provider service. No key values belong in request configs, logs, fixtures or reports.

Transport uses Aviasales and Yandex schedules; Wikipedia and the local OSM index provide attractions. External responses are bounded, redirects are not followed, requests are rate-limited per provider, and successful GET responses are cached in process. City discovery requests are returned for a caller to resolve; the service does not select an ambiguous city automatically.

Hotelbeds is enabled only with `HOTELBEDS_MODE=evaluation`, `HOTELBEDS_API_KEY` and `HOTELBEDS_SECRET`. RateHawk requires `RATEHAWK_MODE=sandbox` or `production`, `RATEHAWK_KEY_ID` and `RATEHAWK_API_KEY`. Hotel searches currently support one explicitly specified room. Neither path books a hotel. Sandbox/evaluation offers are labeled and must not be presented as live production inventory. Hotelbeds evaluation returned offers in an Aveiro test; wider destination coverage and RateHawk access remain unverified. In live mode, stage-two `SearchParts` includes configured hotel providers when the input includes confirmed hotel dates, destination coordinates and the explicit source market. Missing data and provider failures are returned as hotel errors; unverified filters and cancellation conditions are warnings. Demo mode remains synthetic. See the [run report](../evidence/2026-10-28/RUN_REPORT.ru.md).

The supplied OSM index is a partial Kazan snapshot: 17 museums captured on 2026-10-25 and one named theatre and cinema checked on 2026-10-28. It is not a complete venue directory. Import a licensed Overpass extract using `python3 scripts/import_osm_pois.py INPUT OUTPUT --coverage DESCRIPTION`; public Nominatim must not be used to download POI lists. OSM locations do not establish opening hours, ticket prices or feasibility during a layover.

## Remaining team integration

1. AI and frontend owners should convert their input into the documented contract, resolve missing fields with the traveler, and retain unsupported preferences. Their branches are still separate.
2. The gateway and travel service call `SearchParts` and `BuildRoutes`; assembly supports dated stops and bounded same-station connections. Connect `candidates` to frontend and resolve discovery candidates. AI ranking and cross-station transfer calculations remain unimplemented. Multi-city live searches require a verified `resolved.stops` entry per stop; end-to-end live coverage remains unverified. See [route assembly](../route/README.md).
3. Enforce hard traveler constraints, time-zone-aware connections, price scope, actual hotel details and prebook before recommending a complete trip. Cached Aviasales prices are not confirmed seats; Yandex schedules often lack fares.
4. Run the manual scenario matrix against the integrated app and record evidence. Provider-level probes and the separate synthetic frontend demo do not by themselves establish an end-to-end pass.
