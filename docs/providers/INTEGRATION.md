# Integration notes for `providers-api`

This page describes the checked state of the local `providers-api` working tree on 2026-10-27. No merge or commit has been made.

## Git and workspace

The historical heads `origin/main` and `origin/providers-api` conflict on `.env.example` and `.gitignore`. The local files contain the intended union: service addresses, blank provider/AI credentials, and ignore rules for local secrets and Go artifacts. Whoever performs the merge must select those combined contents. `origin/main` also tracks `.env`; an ignore rule alone does not stop tracking it. Remove it from the Git index before placing secrets in that filename.

`origin/main` is a multi-module workspace and its `go.work` does not include the root module. Add `.` to the `use` block after merging. Then services can import `github.com/ostrovok-hackathon-2026/godnota/providerkit`. The implementation in `internal/providers` cannot be imported directly from `github.com/godnota/provider-service`; use the public wrapper.

## Data flow

1. Convert the AI response into `TripRequest` and validate it against `configs/trip_request.schema.json` and `TripRequest.ValidateAt`. The current AI branch uses a flat `travel.v1.Config`; fields for pets, hotel filters, modes, country and month-only dates are absent there. Define an explicit adapter and preserve unknown values as unresolved.
2. Resolve each city to IATA, Yandex code and coordinates. Match country, place type and city name; do not assume the first autocomplete hit is correct. Obtain the hotel region from the future hotel source or use resolved coordinates. Preserve the user's choice when multiple cities share a name.
3. Determine hotel check-in and checkout from actual transport arrival/departure dates. A night train can shift check-in to the next calendar day. Pass those dates through `ResolvedTrip`.
4. Call `providerkit.BuildTripPlan(input, resolved, now)`. Execute `Discovery` requests first, fill missing resolution values, then rebuild. Only execute `Search` requests whose required fields are resolved. `Missing` is actionable; `Limitations` must stay with the resulting offers.
5. `providerkit.LoadRegistry` and `Registry.NewHTTPRequest` validate and encode typed provider requests. The caller still needs bounded responses, rate limiting, retry rules, caching and provider-specific response normalizers. Never follow a redirect carrying a provider credential. `cmd/smoke` demonstrates bounded reads and blocked redirects, but is a probe tool rather than a service client.
6. Return normalized offers with provider, retrieval time, currency, amount scale, price scope, live/cached/synthetic status and missing data. `origin/main` provider/map/route protobuf requests and responses are empty; define these fields before wiring gRPC controllers. The current `provider-service` returns empty results.
7. Adapt normalized routes to `origin/frontend`'s `Trip` shape. That UI expects stop/segment IDs, WGS84 coordinates, ISO datetimes with offsets, money in minor units, authoritative totals, a revision and idempotent demo booking. Its current demo adapter does not match `origin/main` gateway routes.

## Source boundaries

- Aviasales Data API provides cached fares, not confirmed seats or final checkout prices.
- Yandex Rasp provides schedules, often without ticket prices. Missing price must stay unknown.
- Hotel preferences remain in `TripRequest`, but no hotel provider is configured. ETG / RateHawk requires a contract and a separate HTTPS adapter; see `docs/providers/hotels.md`.
- Wikimedia provides articles about nearby places, not opening hours or event inventory. Each image needs its own reuse information.
- Public Nominatim is for individual geocoding, not automatic lists of museums or cinemas. A curated offline OSM index or a separately approved POI source is needed for those interests.
- Layover activities require a real stop, free time, travel time to the place, opening hours and a margin before departure; city-centre proximity alone is insufficient.

The report template and scenario matrix in `docs/` should be completed against the final integrated build, with fixture names and recorded evidence. Do not mark the full journey as passed based on this provider module alone.
