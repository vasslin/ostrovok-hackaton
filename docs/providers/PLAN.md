# Provider APIs: implementation plan

**Task 2 (Даня Антонов):** find travel provider APIs (tickets, hotels, leisure during the trip and during layovers) and deliver:

1. **Config JSON files**: one unified config that the AI agent fills; one config per provider API (in code: a map keyed by provider name, passed into the HTTP request); required fields and format limits marked.
2. **Provider API descriptions**: paths, methods, parameters and responses, detailed enough to code the calls.

Every provider below was checked on **2026-10-25** with a real request (results in §1). "To verify" marks the few details I couldn't confirm from official docs.

---

## 0. Deliverables and "done"

| # | Deliverable | Files | Done when |
|---|---|---|---|
| 1a | Unified trip config (the AI fills it) | `configs/trip_request.schema.json`, `configs/examples/trip_request.*.json` | Соня's ai-service produces it from a chat message and it validates |
| 1b | Provider registry + one request config per endpoint | `configs/providers.json`, `configs/providers/<provider>.<endpoint>.schema.json`, `configs/examples/<provider>.<endpoint>.json` | each example validates, and a real call with it succeeds |
| 1c | Required fields and format rules | Go structs with tags → generated `*.schema.json` | `go test` checks every example against its schema and struct |
| 2 | API descriptions | `docs/providers/<provider>.md` + real sample responses in `testdata/<provider>/` | a teammate can write the client from the doc alone |

---

## 1. Provider catalog (checked 2026-10-25)

### Tickets

| Config key | Provider | What it gives | Data | Access | Probe result |
|---|---|---|---|---|---|
| `aviasales` | Travelpayouts / Aviasales Data API v3 | flight prices users found recently (cache, not live seats), airline, transfers, deep link | real, cached | free partner token | 401 without token: endpoint alive ✓ |
| `travelpayouts_places` | Travelpayouts autocomplete + static `cities.json` | city or airport → IATA code, coordinates, time zone, Russian word forms («в Казань») | real | none | 200 ✓ (Казань → `KZN`) |
| `yandex_rasp` | Yandex Rasp API v3.0 (Яндекс Расписания) | schedules for **plane, train, suburban, bus, water (ships, ferries), helicopter**; prices for some segments | real | free key | 400 «Не указан ключ» on `api.rasp.yandex-net.ru`: alive ✓ |
| `transport_mock` | our generator | prices where Rasp and Aviasales have none (most trains, buses, ferries) | synthetic | — | to build |

### Hotels

| Config key | Provider | What it gives | Data | Access | Probe result |
|---|---|---|---|---|---|
| `hotels_mock` | **our service with the ETG / RateHawk B2B v3 contract** | region search, hotel page, prebook, booking, status, cancel | real hotel names and places (OSM / OpenTripMap) + synthetic rates | internal | to build |
| — | ETG / RateHawk B2B API | real inventory | — | contract only | docs public, no access |

There is no free hotel API with real prices. We build `hotels_mock` on ETG's documented contract, so swapping in a real supplier later is an adapter, not a rewrite.

### Leisure (museums, cinemas, theatres, events; also during layovers)

| Config key | Provider | What it gives | Data | Access | Probe result |
|---|---|---|---|---|---|
| **`wikipedia`** | **Wikipedia API, `generator=geosearch` (primary)** | notable places near a point, with a short description, photo, coordinates and link, in one call | real (CC BY-SA) | none | **200 ✓**: 20 places around the Kazan Kremlin, all with photos |
| **`nominatim.search_amenity`** | Nominatim structured search (`amenity` + `city`) | places of one type in a city: museums, theatres, cinemas, parks, restaurants | real (OSM) | none | **200 ✓**: 17 museums in Kazan; few cinemas (partial coverage) |
| `osm_overpass` | OSM Overpass API | any OSM object (hotels, museums, cinemas) | real | none | public server: «at capacity… for automated use run a local instance» ⚠ **offline seeding only** |
| `events_mock` | our generator | festivals and showtimes for the demo | synthetic | — | to build |

**Checked and not working (2026-10-25), not used:** Hotellook (404, shut down), KudaGo (anti-bot page), OpenTripMap (the registration email never arrives), Timepad (token only).

### Connected services

| Config key | Provider | What it gives | Access | Probe result |
|---|---|---|---|---|
| `nominatim` | OSM Nominatim | geocoding: city or address → coordinates | none; ≤1 req/s; User-Agent required | 200 ✓ |
| `open_meteo` | Open-Meteo | forecast, history (ERA5, 1940+), **sea temperature** | none; free non-commercial: <10 000 calls/day, 5 000/h, 600/min; CC BY 4.0 | 200 ✓ |
| `osrm` | OSRM demo server | drive time and distance for transfers | none; light use only | 200 ✓ (Kazan airport → Kremlin: 29 km, ~54 min) |
| `cbr` | Bank of Russia daily rates | currency conversion to RUB | none | not probed (standard XML) |

---

## 2. How the pieces fit

```
user text ─► ai-service ─fills─► TripRequest  (unified config: human values only)
                                    │ validate against schema → missing_required?
                                    │   many missing → agent asks in chat; few → user fills a form
                                    ▼
travel-service ─► provider-service
                    1. Resolver   city → IATA, Rasp code, coordinates, time zone, hotel region   (cached)
                    2. Builders   TripRequest + resolved codes → map[provider][]Request          (pure Go)
                    3. Client     generic HTTP: auth, encoding, rate limit, retry, cache (providers.json)
                    4. Normalizers  provider JSON → TransportOption / HotelOffer / LeisureItem / Weather
                    ▼
               normalized results ─► travel-service (assemble trip, map graph) ─► ai-service (rank, explain)
```

Rules that keep it reliable:

- **The AI fills human values only** (city names, dates, counts, wishes). It never writes provider codes (`KZN`, `c213`, hotel ids). The Resolver produces those, so the model can't hallucinate a code.
- **Provider requests are built in Go**, deterministically, from `TripRequest` + resolved codes. Same input gives the same calls, which makes them easy to test.
- **The board's «> незаполненных полей — через агента, < — сами»**: the backend fills `completeness.missing_required` (JSON paths). If the count is at or above a threshold (config value), the agent asks in chat; below it, the frontend shows just those fields as a form.

---

## 3. Deliverable 1a: unified config `TripRequest`

One JSON object per search. The AI fills it; the backend validates it and adds `completeness`. `*` = required.

| Path | Type | Format / limits | Example |
|---|---|---|---|
| `schema_version`* | string | const `"1.0"` | `"1.0"` |
| `locale`* | string | `ru` \| `en` | `"ru"` |
| `currency`* | string | ISO 4217, `^[A-Z]{3}$` | `"RUB"` |
| `origin.city`* | string | 1–100 chars | `"Москва"` |
| `origin.country_code` | string | ISO 3166-1 alpha-2, `^[A-Z]{2}$` | `"RU"` |
| `destination.city` | string \| null | 1–100 chars; `null` = "don't know where", inspiration mode | `"Казань"` |
| `destination.candidates` | string[] | ≤5 items, when the AI suggests options | `["Хургада","Анталья"]` |
| `dates.departure_date` | string | `YYYY-MM-DD`, ≥ today, ≤ today + 365 d | `"2026-12-12"` |
| `dates.departure_month` | string | `YYYY-MM`, used when only a month is known | `"2026-11"` |
| `dates.return_date` | string | `YYYY-MM-DD`, after departure, ≤ departure + 30 d (hotel API limit) | `"2026-12-14"` |
| `dates.nights` | int | 1–30 | `2` |
| `dates.flexible_days` | int | 0–7 | `0` |
| `travelers.adults`* | int | 1–9 | `2` |
| `travelers.children_ages` | int[] | each 0–17, ≤8 items | `[6]` |
| `travelers.pets` | object[] | `{kind: dog\|cat\|other, weight_kg: 0.1–100}`, ≤3 | `[]` |
| `travelers.residency` | string | ISO alpha-2 **lowercase** (ETG format) | `"ru"` |
| `transport.modes` | string[] | `plane` \| `train` \| `suburban` \| `bus` \| `water` \| `helicopter` (Rasp enum) | `["train","plane"]` |
| `transport.round_trip` | bool | default `true` | `true` |
| `transport.direct_only` | bool | default `false` | `false` |
| `transport.max_transfers` | int | 0–3 | `1` |
| `transport.max_duration_hours` | number | 0.5–48 | `12` |
| `transport.departure_time_from` / `_to` | string | `HH:MM` local, `^([01]\d\|2[0-3]):[0-5]\d$` | `"08:00"` |
| `hotel.needed` | bool | default `true` | `true` |
| `hotel.rooms` | int | 1–9 | `1` |
| `hotel.star_rating_min` | int | 0–5 | `3` |
| `hotel.kinds` | string[] | `hotel` \| `hostel` \| `apartment` \| `guest_house` \| `resort` | `["hotel"]` |
| `hotel.meal` | string | `nomeal` \| `breakfast` \| `half_board` \| `full_board` \| `all_inclusive` | `"breakfast"` |
| `hotel.free_cancellation` | bool | | `true` |
| `hotel.price_per_night_max` | int | > 0, in `currency` | `8000` |
| `hotel.amenities` | string[] | `pool` \| `kids_club` \| `pet_friendly` \| `parking` \| `wifi` \| `spa` \| `beach` \| `accessible` | `["wifi"]` |
| `hotel.location` | string[] | `center` \| `beach` \| `quiet` \| `walkable` \| `near_station` \| `near_airport` | `["center","walkable"]` |
| `hotel.near_place` | string | ≤100 chars | `"Казанский Кремль"` |
| `leisure.interests` | string[] | `museums` \| `cinemas` \| `theatres` \| `concerts` \| `festivals` \| `parks` \| `nature` \| `architecture` \| `history` \| `food` \| `kids` \| `sport` \| `nightlife`; ≤8 | `["museums","theatres","food"]` |
| `leisure.during_layovers` | bool | default `true` | `true` |
| `leisure.min_layover_hours` | number | 2–24, default 4 | `4` |
| `leisure.max_items_per_stop` | int | 1–20, default 5 | `5` |
| `budget.total` | int | > 0, in `currency` | `40000` |
| `budget.scope` | string | `total` \| `per_person` \| `per_night` | `"total"` |
| `budget.is_hard` | bool | default `false` | `false` |
| `ai_context.trip_purpose` | string | `leisure` \| `business` \| `family` \| `romantic` \| `friends` \| `other` | `"leisure"` |
| `ai_context.walkable_area` | bool | «прогулочный район» | `true` |
| `ai_context.events_wanted` | string[] | ≤5 items, ≤50 chars each | `["фестиваль"]` |
| `ai_context.must_have` / `must_avoid` | string[] | ≤10 items, ≤80 chars each | `["тихий номер"]` |
| `ai_context.notes` | string | ≤500 chars | |
| `ai_context.assumptions` | object[] | `{field: JSON path, value, reason ≤200}`; shown to the user as editable chips | `[{"field":"/origin/city","value":"Москва","reason":"из профиля"}]` |
| `completeness.missing_required` | string[] | JSON paths; **filled by the backend, not the AI** | `["/dates/departure_date"]` |

Also: at least one of `dates.departure_date` / `dates.departure_month` is required (`anyOf` in the schema).

Example, exact request («Москва → Казань на выходные 12–14 декабря, вдвоём, поездом или самолётом, отель у Кремля с завтраком до 8 000 за ночь, музеи и театры»):

```json
{
  "schema_version": "1.0", "locale": "ru", "currency": "RUB",
  "origin": {"city": "Москва", "country_code": "RU"},
  "destination": {"city": "Казань", "country_code": "RU"},
  "dates": {"departure_date": "2026-12-12", "return_date": "2026-12-14", "nights": 2, "flexible_days": 0},
  "travelers": {"adults": 2, "children_ages": [], "pets": [], "residency": "ru"},
  "transport": {"modes": ["train", "plane"], "round_trip": true, "direct_only": true},
  "hotel": {"needed": true, "rooms": 1, "star_rating_min": 3, "meal": "breakfast", "free_cancellation": true,
            "price_per_night_max": 8000, "location": ["center", "walkable"], "near_place": "Казанский Кремль"},
  "leisure": {"interests": ["museums", "theatres", "food"], "during_layovers": true, "max_items_per_stop": 5},
  "budget": {"total": 40000, "scope": "total", "is_hard": false},
  "ai_context": {"trip_purpose": "leisure", "walkable_area": true, "events_wanted": [], "assumptions": []},
  "completeness": {"missing_required": []}
}
```

Make 4 more examples from Василина's scenarios: the vague one («к тёплому морю в ноябре, недорого, с ребёнком»: `destination: null`, `departure_month`), Санкт-Петербург → Москва by train, a water route (e.g. Петрозаводск → Кижи, `modes: ["water"]`), and a family with a dog.

---

## 4. Deliverable 1b: per-provider configs

Two kinds of files.

**(1) `configs/providers.json`: static registry**, one entry per provider (base URL, auth, limits, endpoints). Secrets stay in env vars:

```json
{
  "aviasales": {
    "base_url": "https://api.travelpayouts.com",
    "auth": {"type": "header", "name": "X-Access-Token", "env": "TRAVELPAYOUTS_TOKEN"},
    "rps": 5, "timeout_ms": 8000, "cache_ttl_s": 3600,
    "endpoints": {
      "prices_for_dates": {"method": "GET", "path": "/aviasales/v3/prices_for_dates", "params_in": "query"}
    }
  },
  "yandex_rasp": {
    "base_url": "https://api.rasp.yandex-net.ru",
    "auth": {"type": "query", "name": "apikey", "env": "YANDEX_RASP_API_KEY"},
    "rps": 2, "timeout_ms": 8000, "cache_ttl_s": 21600,
    "attribution": "Данные: Яндекс Расписания (баннер, текст и ссылка — /v3.0/copyright/)",
    "endpoints": {
      "search":             {"method": "GET", "path": "/v3.0/search/",             "params_in": "query"},
      "nearest_settlement": {"method": "GET", "path": "/v3.0/nearest_settlement/", "params_in": "query"},
      "thread":             {"method": "GET", "path": "/v3.0/thread/",             "params_in": "query"},
      "copyright":          {"method": "GET", "path": "/v3.0/copyright/",          "params_in": "query"}
    }
  },
  "hotels_mock": {
    "base_url": "http://provider-service:8080/hotels",
    "auth": {"type": "basic", "env": "HOTELS_KEY_ID:HOTELS_KEY"},
    "rps": 10, "timeout_ms": 5000, "cache_ttl_s": 0,
    "endpoints": {
      "multicomplete": {"method": "POST", "path": "/api/b2b/v3/search/multicomplete/", "params_in": "json"},
      "search_region": {"method": "POST", "path": "/api/b2b/v3/search/serp/region/",   "params_in": "json"},
      "search_geo":    {"method": "POST", "path": "/api/b2b/v3/search/serp/geo/",      "params_in": "json"},
      "hotelpage":     {"method": "POST", "path": "/api/b2b/v3/search/hp/",            "params_in": "json"},
      "prebook":       {"method": "POST", "path": "/api/b2b/v3/hotel/prebook/",        "params_in": "json"},
      "booking_form":  {"method": "POST", "path": "/api/b2b/v3/hotel/order/booking/form/",          "params_in": "json"},
      "booking_finish":{"method": "POST", "path": "/api/b2b/v3/hotel/order/booking/finish/",        "params_in": "json"},
      "booking_status":{"method": "POST", "path": "/api/b2b/v3/hotel/order/booking/finish/status/", "params_in": "json"},
      "cancel":        {"method": "POST", "path": "/api/b2b/v3/hotel/order/cancel/",               "params_in": "json"}
    }
  }
}
```

(`rps` and `cache_ttl_s` are our own conservative choices, not official limits. Nominatim and OSRM must be `rps: 1`.)

**(2) Request config per endpoint**: the exact parameters that go into that HTTP call, with field names exactly as the provider expects, so the client can encode them directly. Built from the Kazan example above:

```json
// configs/examples/aviasales.prices_for_dates.json
{"origin": "MOW", "destination": "KZN", "departure_at": "2026-12-12", "return_at": "2026-12-14",
 "one_way": false, "direct": true, "currency": "rub", "sorting": "price", "limit": 30, "page": 1}

// configs/examples/yandex_rasp.search.json  (outbound; return leg = from/to swapped, date = return_date)
{"from": "c213", "to": "c43", "date": "2026-12-12", "transport_types": "train",
 "transfers": false, "limit": 100, "offset": 0, "lang": "ru_RU", "format": "json"}

// configs/examples/hotels_mock.search_geo.json  (near the Kremlin; price_to = 8000 × 2 nights, filters apply to the whole stay)
{"checkin": "2026-12-12", "checkout": "2026-12-14", "residency": "ru", "language": "ru", "currency": "RUB",
 "latitude": 55.7987, "longitude": 49.1059, "radius": 2000, "hotels_limit": 20,
 "guests": [{"adults": 2, "children": []}],
 "filter": {"star_rating": [3, 4, 5], "meal_type": "breakfast", "price_to": 16000}}

// configs/examples/wikipedia.geosearch.json
{"action": "query", "generator": "geosearch", "ggscoord": "55.7987|49.1059", "ggsradius": 3000, "ggslimit": 20,
 "prop": "coordinates|pageimages|description|info", "inprop": "url", "piprop": "thumbnail", "pithumbsize": 400,
 "format": "json", "formatversion": 2}
```

**In Go**, matching the board's "map [имя провайдера] → config":

```go
type Request struct {
    Provider string // "aviasales"
    Endpoint string // "prices_for_dates"
    Params   any    // e.g. AviasalesPricesForDates{...}: a struct with json/url/validate tags
}

// key = provider name; a slice, because one provider may need several calls
// (outbound + return leg, search → hotel page → prebook)
type RequestSet map[string][]Request

func BuildRequests(trip TripRequest, codes Resolved) (RequestSet, error)
func (c *Client) Do(ctx context.Context, r Request, out any) error // reads providers.json: URL, method, auth, encoding
```

---

## 5. Deliverable 1c: marking required fields and formats

A JSON file can't mark its own fields as required (no comments, no annotations). The standard way is a **JSON Schema** next to it (`"required": [...]`, `pattern`, `minLength`, `minimum`, `enum`, `format`). In Go the same rules live in **struct tags**.

**Recommendation: Go structs are the single source of truth, and the JSON Schemas are generated from them.**

```go
// GET /aviasales/v3/prices_for_dates
type AviasalesPricesForDates struct {
    Origin      string `json:"origin"       url:"origin"       validate:"required,len=3,uppercase" jsonschema:"required,pattern=^[A-Z]{3}$,description=IATA code of origin city or airport"`
    Destination string `json:"destination"  url:"destination"  validate:"required,len=3,uppercase" jsonschema:"required,pattern=^[A-Z]{3}$"`
    DepartureAt string `json:"departure_at,omitempty" url:"departure_at,omitempty" validate:"omitempty,datetime=2006-01-02|datetime=2006-01" jsonschema:"pattern=^[0-9]{4}-[0-9]{2}(-[0-9]{2})?$"`
    ReturnAt    string `json:"return_at,omitempty"    url:"return_at,omitempty"    validate:"omitempty,datetime=2006-01-02|datetime=2006-01" jsonschema:"pattern=^[0-9]{4}-[0-9]{2}(-[0-9]{2})?$"`
    OneWay      bool   `json:"one_way"      url:"one_way"`
    Direct      bool   `json:"direct"       url:"direct"`
    Currency    string `json:"currency"     url:"currency"     validate:"required,len=3,lowercase" jsonschema:"required,default=rub"`
    Sorting     string `json:"sorting"      url:"sorting"      validate:"oneof=price route"        jsonschema:"enum=price,enum=route"`
    Limit       int    `json:"limit"        url:"limit"        validate:"min=1,max=100"            jsonschema:"minimum=1,maximum=100,default=30"` // our cap
    Page        int    `json:"page"         url:"page"         validate:"min=1"                    jsonschema:"minimum=1,default=1"`
}
```

| Tag | Library | Used for |
|---|---|---|
| `validate:"required,len=3,…"` | `github.com/go-playground/validator/v10` | runtime check before every request and on the AI's output |
| `url:"origin,omitempty"` | `github.com/google/go-querystring/query` | encoding GET query strings |
| `json:"origin"` | `encoding/json` | POST bodies, example files |
| `jsonschema:"required,pattern=…"` | `github.com/invopop/jsonschema` (with `RequiredFromJSONSchemaTags: true`) | `go generate` writes `configs/**/*.schema.json` |

Why this way:

- Rules live in one place.
- Go gets runtime validation.
- The generated schema goes to three teammates. Соня passes it to the LLM as the structured-output / function schema, so the model fills exactly these fields. Василина uses it in the API docs. Даня Казанович renders the missing fields as a form from it.

Optionally validate raw incoming JSON (e.g. the AI's output) with `github.com/santhosh-tekuri/jsonschema/v6`. Its errors carry JSON paths, which go straight into `completeness.missing_required`.

Alternatives, if the team prefers: hand-write the JSON Schemas and keep the Go structs in sync with a test, or go schema-first with codegen (`github.com/omissis/go-jsonschema`). Both work, but they mean more tooling or more drift.

### Format cheat sheet (what the providers actually require)

| Thing | Rule | Who |
|---|---|---|
| Dates in requests | `YYYY-MM-DD` everywhere; Aviasales also accepts `YYYY-MM` | all |
| Date-times in responses | ISO 8601 with offset. Rasp: station local time unless `result_timezone`; **Rasp `thread` stops: `YYYY-MM-DD HH:MM:SS` without offset**. ETG cancellation times: UTC. **Keep offsets; never drop them** | Rasp, Aviasales, ETG |
| Durations | Rasp `duration` and OSRM: seconds; Aviasales `duration_to` / `duration_back`: **minutes** | Rasp, OSRM, Aviasales |
| IATA codes | 3 uppercase letters `^[A-Z]{3}$`, city (`MOW`) or airport (`SVO`) | Aviasales, places |
| Yandex codes | city `c` + digits (`c213` Moscow), station `s` + digits (`s9600213`); other systems via `system=iata\|sirena\|express\|esr` | Rasp |
| Country | ISO 3166-1 alpha-2: ETG `residency` **lowercase** (`ru`), places `country_code` uppercase (`RU`) | ETG, places |
| Currency | ISO 4217: Aviasales param **lowercase** (`rub`), ETG uppercase (`RUB`) | |
| Language | Rasp `ru_RU`; ETG `ru`; places `locale=ru`; Wikipedia host `ru.wikipedia.org` | |
| Coordinates | decimal degrees, lat −90..90, lon −180..180; **OSRM order is `lon,lat`** | all geo |
| Hotel guests | ≤9 rooms (1 object = 1 room); adults 1–6 per room; children: ages 0–17, ≤4 per room; ≤10 guests per room | ETG |
| Hotel stay | `checkin` ≤ today + 730 days; `checkout` ≤ `checkin` + 30 days | ETG |
| Rate ids | prebook `hash` 1–256 chars, valid 24 h after the hotel page; `search_hash` valid 38 min | ETG |
| Booking id | `partner_order_id`: UUID, unique per booking attempt | ETG |
| Text limits | ETG `user.comment` 1–256 chars; Rasp `stops` ≤1000; Rasp `thread.uid` ≤100 | |
| Paging | Rasp `limit` ≤100 (default 100); Nominatim `limit` ≤40 | |
| Money | Rasp `price: {whole, cents}` integers; ETG amounts are strings | |
| Weather range | Open-Meteo forecast ≤16 days ahead; marine forecast 0–8 days; history from 1940 (sea temperature from 2022) | |

---

## 6. Deliverable 2: API descriptions

> **Done:** one file per provider, in Russian, with working `curl` commands and real responses — see [README.md](README.md). This section stays as the compact English reference.

**Template** for each `docs/providers/<provider>.md`:

1. Purpose, status (date checked), data type (real / cached / synthetic)
2. Access: how to get the key, env var name, where to send it
3. Limits and terms: rate limits, attribution, caching rules
4. Endpoints: method, URL, parameter table (name, type, required, format, default), working `curl` example, response fields we use, errors
5. Mapping from `TripRequest` (from §7)
6. Real sample responses saved to `testdata/<provider>/<endpoint>.json`: used in unit tests and as an offline fallback during the defense

**Reference content** (verified details to put into those files):

### `aviasales`: Travelpayouts Aviasales Data API

- **Access:** free Travelpayouts partner account → API token in the profile. Send as `X-Access-Token` header (or `token` query param).
- **Data:** cheapest prices Aviasales users found recently (freshness ~48 h, kept up to 7 days). Not live availability. Use `link` to deep-link to aviasales.ru.
- **`GET https://api.travelpayouts.com/aviasales/v3/prices_for_dates`**
  - Params:
    - `origin`*, `destination`*: IATA city or airport
    - `departure_at`, `return_at`: `YYYY-MM` or `YYYY-MM-DD`; omit `return_at` for one-way
    - `one_way` (bool): `true` returns one ticket per date group
    - `direct` (bool)
    - `currency`: default `rub`
    - `sorting`: `price` \| `route`
    - `unique` (bool), `limit`, `page`
  - Response (**confirmed by a real call on 2026-10-25**, saved in `testdata/aviasales/prices_for_dates.json`):
    - `{success, currency, data: [...]}`
    - Each item: `origin`, `destination`, `origin_airport`, `destination_airport`, `departure_at`, `return_at`, `price`, `airline`, `flight_number`, `transfers`, `return_transfers`, `duration`, `duration_to`, `duration_back`, `gate`, `link`.
    - `departure_at` / `return_at`: local time with offset, e.g. `2026-12-12T20:35:00+03:00`.
    - `duration_to` / `duration_back`: **minutes** (MOW → KZN: 105).
    - `price`: integer in `currency`; with `return_at` set it is the round-trip price.
    - `link`: relative path, prefix it with `https://www.aviasales.ru`.
  - Errors: 401 without token (verified).
- **Lookups (no key, verified):**
  - `GET https://autocomplete.travelpayouts.com/places2?term=Казань&locale=ru&types[]=city` (or `types[]=airport`) → `[{code: "KZN", name, country_code, coordinates: {lat, lon}, cases: {vi: "в Казань", ro: "Казани", …}}]`
  - `GET https://api.travelpayouts.com/data/ru/cities.json` (3.5 MB, download once): `code`, `name`, `coordinates`, `time_zone`, `has_flightable_airport`
- To verify in official docs: the rate limit and the maximum `limit`.

### `yandex_rasp`: Yandex Rasp API v3.0

- **Access:**
  - Free key from the Yandex Developer Cabinet (developer.tech.yandex.ru), product «API Яндекс.Расписаний».
  - Send as `apikey` query param or `Authorization` header.
  - **Domain: `api.rasp.yandex-net.ru`.** The old `api.rasp.yandex.net` may not work.
  - The daily limit is shown in the cabinet (check it on day 1).
  - Attribution is required: banner, text and link (`/v3.0/copyright/`).
- **`GET /v3.0/search/`**: threads between two points.
  - Params:
    - `from`*, `to`*: Yandex codes (`c213` city, `s9600213` station), or another system with `system=iata|sirena|express|esr`
    - `date`: `YYYY-MM-DD`. Without it you get only direct threads, for all dates.
    - `transport_types`: `plane` \| `train` \| `suburban` \| `bus` \| `water` \| `helicopter`; omit for all types
    - `transfers` (bool, default `false`)
    - `limit` (default and max 100), `offset` (default 0)
    - `lang` (`ru_RU` \| `uk_UA`), `format` (`json` \| `xml`)
    - `result_timezone` (tz name), `show_systems` (`yandex` \| `esr`), `add_days_mask` (bool)
  - Response:
    - `segments[]`:
      - `departure` / `arrival`: ISO 8601 with offset
      - `duration`: seconds
      - `from` / `to`: `{code, title, station_type, transport_type}`
      - `thread`: `{uid, number, title, vehicle, transport_type, carrier: {code, title, codes: {iata, icao, sirena}}, transport_subtype}`
      - `stops` (string ≤1000), `has_transfers`, platforms and terminals
      - **`tickets_info`**: `{et_marker, places: [{name, currency, price: {whole, cents}}]}`, often empty
    - `pagination: {total, limit, offset}`
    - `interval_segments`: services with no fixed times
- **Other endpoints** (all verified alive):
  - `GET /v3.0/nearest_settlement/?lat=&lng=&distance=` → city code for coordinates. This is how the Resolver gets `c43` for Казань.
  - `GET /v3.0/thread/?uid=` → stops of a route
  - `GET /v3.0/schedule/?station=&date=` → station timetable
  - `GET /v3.0/stations_list/` → full directory (large; download once)
- **Verified with a real key on 2026-10-25** (responses in `testdata/yandex_rasp/`):
  - `search` Москва (`c213`) → Казань (`c43`), trains, 2026-12-12: 10 trains, **none with prices** (`tickets_info.places` empty). Times come as `2026-12-12T20:40:00+03:00`, `duration` in seconds.
  - `nearest_settlement` for Kazan coordinates → `c43` (confirmed).
  - `thread` for `002J_0_2` (train 002Й) → 4 stops. **Stop times have a different format: `2026-12-12 20:40:00`, no `T` and no offset.** Parse them in the station's time zone.
  - `copyright` → the attribution text, link and logo to show.
  - A new key can answer `429 «Ключ не активен»` for a while after it's issued.
- **Gotchas:**
  - No month search: for `departure_month`, query 2–3 sample dates.
  - Prices are often missing, so fall back to `transport_mock` for the price but keep the real times.

### `hotels_mock`: our service on the ETG / RateHawk B2B v3 contract

- **Why:** no free hotel price API exists (Hotellook returns 404; ETG needs a contract). We implement ETG's documented contract ourselves.
- **Conventions (as in ETG):**
  - HTTP Basic auth (`KEY_ID:KEY`).
  - JSON bodies.
  - Envelope `{data, debug, error, status}`.
  - Rate-limit headers `X-RateLimit-SecondsNumber`, `-RequestsNumber`, `-Remaining`, `-Reset`.
- `POST /api/b2b/v3/search/multicomplete/` `{query*, language}` → ≤5 regions + ≤5 hotels (region lookup for the Resolver)
- `POST /api/b2b/v3/search/serp/region/`
  - Request:
    - `checkin`* (≤ today + 730 d), `checkout`* (≤ checkin + 30 d)
    - `region_id`* (int)
    - `residency` (alpha-2 lowercase), `language`, `currency`, `hotels_limit`
    - `guests`: `[{adults* 1–6, children: [ages 0–17, ≤4]}]`, one object per room, ≤9 rooms
    - `filter`: `{star_rating[], kind[], meal_type, price_from, price_to, has_transfer}`
    - `upsells`: `{early_checkin: {time}, late_checkout: {time}}`
  - Response `hotels[]`:
    - `hid`: int, ≤10 digits
    - `rates[]`:
      - `search_hash` (valid 38 min), `daily_prices[]`, `room_name`, `allotment`, `amenities_data[]`
      - `meal_data`: `{value, has_breakfast, no_child_meal}`
      - `payment_options.payment_types[]`: `{amount, show_amount, currency_code, show_currency_code, type: now|hotel|deposit, tax_data: {taxes[]}, cancellation_penalties: {free_cancellation_before (UTC), policies: [{start_at, end_at, amount_charge, amount_show}]}}`
- `POST /api/b2b/v3/search/serp/geo/`: same plus `latitude`, `longitude`, `radius` (define our own max). `POST /api/b2b/v3/search/serp/hotels/`: same plus `hids` (ETG: ≤300).
- `POST /api/b2b/v3/search/hp/` `{hid*, checkin, checkout, guests, residency, language, currency}` → every rate for one hotel, each with a `book_hash` / `match_hash`
- `POST /api/b2b/v3/hotel/prebook/` `{hash* (1–256 chars, valid 24 h), price_increase_percent (0–100)}` → same or alternative rate. **Any price change must be shown to the user.**
- `POST /api/b2b/v3/hotel/order/booking/form/` `{partner_order_id* (UUID), book_hash*, language*, user_ip*}` → `order_id`, payment types
- `POST /api/b2b/v3/hotel/order/booking/finish/` → starts an async booking. Body:
  - `user`: `{email*, phone, comment 1–256}`
  - `partner`: `{partner_order_id*}`
  - `language`
  - `rooms`*: `[{guests: [{first_name, last_name, is_child, age}]}]`
  - `payment_type`*: `{type: now|hotel|deposit, amount, currency_code}`
- `POST /api/b2b/v3/hotel/order/booking/finish/status/` `{partner_order_id*}` → poll until a final status
- `POST /api/b2b/v3/hotel/order/cancel/` `{partner_order_id*}`
- **Mock behavior to build:**
  - Hotels seeded from real OSM data for ~30 cities.
  - Prices deterministic (seed = `hid` + dates).
  - Prebook changes the price 5–10% of the time.
  - Some dates sold out.
  - Status goes `processing` → `ok`.

### `wikipedia`: primary leisure source (also used for layovers)

- **Access:** no key. Wikimedia asks for an identifying User-Agent (set in `providers.json`). Attribution: «Википедия, CC BY-SA 4.0».
- **`GET https://ru.wikipedia.org/w/api.php`**: one call returns places with everything a card needs.
  - Params:
    - `action=query`*, `generator=geosearch`*, `format=json`*, `formatversion=2` (pages come as a list)
    - `ggscoord`*: `lat|lon`, e.g. `55.7987|49.1059`
    - `ggsradius`*: meters, 10–10000
    - `ggslimit`: up to 50 (our cap, so every page keeps its photo)
    - `prop=coordinates|pageimages|description|info`, `inprop=url`, `piprop=thumbnail`, `pithumbsize=400`
  - Response `query.pages[]`: `title`, `description` (e.g. «мечеть в Казани, Россия»), `coordinates[0].lat/lon`, `thumbnail.source`, `fullurl`, `pageid`. A `continue` object means more pages are available.
  - **Filter:** skip pages without `description`; those are usually areas («Казанская агломерация»), not places.
- **Verified 2026-10-25:** 20 places within 3 km of the Kazan Kremlin, all with photos, 11 with descriptions (Казанский кремль, Кул-Шариф, Благовещенский собор, Хазинэ…).

### `nominatim.search_amenity`: places of one type (museums, theatres, cinemas)

- **`GET https://nominatim.openstreetmap.org/search`** with structured params: `amenity`* (place type: `museum`, `theatre`, `cinema`, `zoo`, `park`, `restaurant`…), `city`*, `country`, `format=jsonv2`*, `limit` (≤40), `accept-language=ru`. Same policy as geocoding: ≤1 req/s, User-Agent.
- **Filter by `type`**: results can include stops named after a place («Театр…» bus stop).
- **Verified 2026-10-25:** `amenity=museum` → 17 museums in Kazan; `theatre` → 5; `cinema` → only 2, so coverage is partial. For a full list, seed once from Overpass (step 11).

### Connected services

- **`nominatim`**:
  - `GET https://nominatim.openstreetmap.org/search?q=&format=jsonv2&limit=1&accept-language=ru` (+ `countrycodes`)
  - Returns `lat` and `lon` as strings, plus `display_name`, `type`, `boundingbox`.
  - Policy: ≤1 req/s, identifying User-Agent, cache results, no bulk use.
- **`open_meteo`** (all verified):
  - Forecast `GET https://api.open-meteo.com/v1/forecast`: `latitude`*, `longitude`*, `daily=temperature_2m_max,temperature_2m_min,precipitation_sum`, `timezone=auto`, ≤16 days.
  - History `GET https://archive-api.open-meteo.com/v1/archive`: `start_date`*, `end_date`*. Use the same dates last year for trips beyond the forecast.
  - Sea `GET https://marine-api.open-meteo.com/v1/marine`: `hourly=sea_surface_temperature`, `start_date` / `end_date` or `forecast_days` 0–8.
- **`osrm`**:
  - `GET https://router.project-osrm.org/route/v1/driving/{lon},{lat};{lon},{lat}?overview=false` → `routes[0].duration` (s), `distance` (m).
  - Demo server, so cache everything.
- **`cbr`**: `GET https://www.cbr.ru/scripts/XML_daily.asp?date_req=DD/MM/YYYY` → XML (windows-1251): `Valute[CharCode, Nominal, Value]`, with a comma as the decimal separator.
- **`osm_overpass`** (seed script only):
  - `POST https://overpass-api.de/api/interpreter` with `data=[out:json][timeout:25];node["tourism"="hotel"](around:R,lat,lon);out;`
  - Use exponential backoff, run once per city, store results in `data/`.
  - For whole regions, use a Geofabrik extract + `osmium` instead (Geofabrik reachable ✓).

---

## 7. Mapping `TripRequest` → provider parameters

| `TripRequest` | `aviasales` | `yandex_rasp.search` | `hotels_mock.search_*` | `wikipedia` / `nominatim.search_amenity` | `open_meteo` |
|---|---|---|---|---|---|
| `origin.city` | `origin` = IATA (places2) | `from` = `c…` (nearest_settlement on city coords) | — | — | — |
| `destination.city` | `destination` = IATA | `to` = `c…` | `region_id` (multicomplete) or `latitude`/`longitude` | `ggscoord` = `lat\|lon` of stop or hotel / `city` | `latitude`, `longitude` |
| `dates.departure_date` | `departure_at` | `date` (outbound) | `checkin` = **arrival** date (overnight trips shift it!) | — | `start_date` |
| `dates.departure_month` | `departure_at` = `YYYY-MM` | 2–3 sample dates | — | — | archive: same month last year |
| `dates.return_date` | `return_at`, `one_way=false` | `date` of return leg, `from`/`to` swapped | `checkout` | — | `end_date` |
| `travelers.*` | price × passengers (per-person prices) | — | `guests[]`: split into rooms (≤6 adults, ≤4 children each) | — | — |
| `travelers.residency` | — | — | `residency` | — | — |
| `transport.modes` | call only if `plane` is in modes | `transport_types` (one call per mode) | — | — | — |
| `transport.direct_only` | `direct=true` | `transfers=false` | — | — | — |
| `hotel.*` | — | — | `filter.star_rating`, `kind`, `meal_type`, `price_to` (= per night × nights), `upsells` | — | — |
| `leisure.interests` | — | — | — | sources by interest (below) | — |
| `currency` | `currency` (lowercase) | — | `currency` | — | — |
| `locale` | `locale` (places2) | `lang=ru_RU` | `language` | `ru.wikipedia.org` / `accept-language` | — |

**Leisure sources by interest:** `architecture`, `history`, sights in general → `wikipedia.geosearch` around the hotel or stop; `museums` → Wikipedia + `nominatim.search_amenity` (`museum`); `theatres` → `theatre`; `cinemas` → `cinema`; `parks`/`nature` → `park`; `food` → `restaurant`, `cafe`; `kids` → `zoo`, `theme_park`. `festivals`/`concerts` → `events_mock`.

**Layovers:** when a gap between two transport segments is ≥ `leisure.min_layover_hours`, call `wikipedia.geosearch` around the transfer station. Keep only places reachable and back within the gap: OSRM time × 2 + 1 h buffer.

---

## 8. Normalized results (provider-service → travel-service)

Every provider's answer becomes one of these, so travel-service and the frontend never see provider formats:

```text
TransportOption { id, provider, mode (plane|train|suburban|bus|water|helicopter), carrier, number,
                  from {code, name, city, lat, lon, tz}, to {…}, departure, arrival (RFC 3339 with offset),
                  duration_min, transfers, price {amount, currency, per_person, source: real|cached|synthetic},
                  deep_link }
HotelOffer      { hid, provider, name, stars, kind, lat, lon, address, checkin_time, checkout_time,
                  rate {book_hash, room_name, meal, free_cancellation_before, total {amount, currency},
                        taxes_at_hotel, source}, amenities[], distance_m {center, beach} }
LeisureItem     { xid, provider, kind, name, lat, lon, rate, url, image, starts_at?, price? }
WeatherSummary  { lat, lon, from, to, t_max_avg, t_min_avg, precipitation_mm, sea_temp_avg? }
```

- `price.source` gives the frontend the real / cached / synthetic badge the jury asked for («если данные моковые, честно скажите»).
- Times always carry an offset, plus an IANA time zone (from `cities.json`). With that, travel-service can check connections and build the map graph: nodes = stops, edges = segments.

---

## 9. Implementation steps

| # | Step | Output | Effort |
|---|---|---|---|
| 1 | Get keys: Travelpayouts token, Yandex Rasp key (both done). Add env names to `.env.example` | keys work | 1 h |
| 2 | `scripts/smoke_providers.sh`: one `curl` per endpoint with the Kazan example; save responses | `testdata/<provider>/*.json` | 1–2 h |
| 3 | `TripRequest` Go struct with tags; `go generate` → `configs/trip_request.schema.json`; 5 example files | deliverable 1a | 3 h |
| 4 | Review with Соня (LLM fills it via the schema) and Василина (scenarios, handlers) | agreed contract | 30 min |
| 5 | Request structs + generated schemas + examples for every endpoint in §6; `configs/providers.json` | deliverable 1b, 1c | 4–5 h |
| 6 | Resolver: places2 + `cities.json` (IATA, coords, tz), Rasp `nearest_settlement`, hotels `multicomplete`; cached | codes for builders | 3 h |
| 7 | `BuildRequests(trip, codes) RequestSet` + golden tests (example trip → expected requests) | builders | 3 h |
| 8 | Generic client: auth from registry, query/JSON encoding, timeouts, retries on 429/5xx with backoff, per-provider rate limiter (`golang.org/x/time/rate`), TTL cache, User-Agent | client | 4 h |
| 9 | Normalizers + tests on the saved `testdata` | normalized results | 4 h |
| 10 | Write `docs/providers/<provider>.md` from the template (alongside steps 5–9; include the working `curl`s) | deliverable 2 | 3 h |
| 11 | `hotels_mock`: seed script (Overpass with backoff or Geofabrik) for ~30 cities, rate generator, ETG-shaped handlers with ETG validation rules | hotel provider | 1–1.5 days |
| 12 | Leisure: interests → sources (Wikipedia, Nominatim by type), layover search, `events_mock` | leisure provider | 4 h |
| 13 | `PROVIDERS_OFFLINE=1` mode that serves `testdata` and cache, so the defense never depends on public APIs | demo safety | 2 h |

**Order:** 1–2 first (real responses shape everything), then 3–4 (unblocks Соня and Василина), then 5–10, then 11–13. Steps 1–10 are about 3 days for one person; 11–13 add 1.5–2 days.

---

## 10. Risks and open questions

- **Aviasales cache gaps:** rare routes or dates come back empty → retry with the month (`YYYY-MM`), then fall back to `transport_mock`.
- **Rasp has no prices for most trains and buses**, and no month search → real times + synthetic prices; sample dates for month queries.
- **Unknown free limits** (Rasp daily cap) → read them from the cabinets on day 1 and cache aggressively.
- **Events:** no free events API works (KudaGo blocked, Timepad token-only) → `events_mock`.
- **Overpass is not for live traffic**: offline seeding only.
- **Where does `hotels_mock` live?** Recommended: inside provider-service (less infrastructure). A separate service also works.
- **Attribution** (Яндекс Расписания, Википедия CC BY-SA, OpenStreetMap, Open-Meteo CC BY 4.0) → a footer on the frontend (Даня Казанович).
- **Overnight trips and time zones:** hotel `checkin` = arrival date, not departure date; always keep offsets.

---

## Sources

- Yandex Rasp API search method: https://yandex.ru/dev/rasp/doc/ru/reference/schedule-point-point · access: https://yandex.ru/dev/rasp/doc/ru/concepts/access · product page: https://yandex.ru/dev/rasp/raspapi
- Travelpayouts Aviasales Data API: https://support.travelpayouts.com/hc/en-us/articles/203956163-Aviasales-Data-API · open-source clients used to confirm v3 params: https://github.com/MikkoParkkola/trvl (`internal/flights/travelpayouts.go`), https://github.com/theYahia/WWmcp (`servers/travelpayouts/src/tools/flights.ts`)
- ETG / RateHawk API v3 (full text): https://docs.emergingtravel.com/llms-full.txt · integration guide: https://docs.emergingtravel.com/docs/integration-guide/
- Wikipedia geosearch: https://www.mediawiki.org/wiki/Extension:GeoData#list=geosearch
- Open-Meteo terms: https://open-meteo.com/en/terms · marine API: https://open-meteo.com/en/docs/marine-weather-api
- Go libraries: https://github.com/go-playground/validator · https://github.com/google/go-querystring · https://github.com/invopop/jsonschema · https://github.com/santhosh-tekuri/jsonschema
