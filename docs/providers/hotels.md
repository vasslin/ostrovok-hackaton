# Hotel search API research

The provider service contains search-only adapters for Hotelbeds evaluation and ETG / RateHawk. They are disabled until configured and have not been verified against a real partner account. No booking endpoint is called. The generic `BuildTripPlan` retains hotel preferences; hotel requests are assembled separately by the service.

## Hotelbeds evaluation

[Hotelbeds](https://developer.hotelbeds.com/documentation/getting-started/) offers an evaluation key and secret for hotel, activities and transfers APIs after registration, currently with a 50-request daily evaluation quota. Evaluation access does not establish production rights or availability for any particular destination.

Set `HOTELBEDS_MODE=evaluation`, `HOTELBEDS_API_KEY` and `HOTELBEDS_SECRET` in the local environment. The adapter sends `POST https://api.test.hotelbeds.com/hotel-api/1.0/hotels`, with `Api-key` and `X-Signature`, the SHA-256 hex digest of API key + secret + current Unix seconds. The [request example](../../configs/hotelbeds/availability.example.json) and [schema](../../configs/hotelbeds/availability.schema.json) describe the supported single-room geolocation search. Source market must be supplied explicitly in `resolved.hotelbeds_source_market`, not guessed from citizenship. Arrival/check-out dates and room allocation must be confirmed by the caller.

Evaluation rates are labeled `sandbox_evaluation`, may be generic net rates and are not presented as confirmed consumer prices. Currency, taxes, rate rules and availability need verification before comparison; `RECHECK` rates require a separate check. The adapter does not execute check-rate or booking. Activities use [availability search](https://developer.hotelbeds.com/documentation/activities/booking-api/availability/availability-/); transfers have a [separate API](https://developer.hotelbeds.com/documentation/transfers/booking-api/search-availability/availability-simple/). Neither is implemented here.

## Alternative providers

- [Ostrovok API](https://docs-api.ostrovok.ru/docs/integration-guide/) offers hotel search and content through `api.ostrovok.ru`; credentials and access are contract-specific. Its documentation states that booking its test hotel can create financial obligations. Do not book it as part of this demo.
- [Booking.com Demand API](https://developers.booking.com/demand/docs/development-guide/authentication) requires an API key and affiliate ID; accommodation search and product availability are partner-dependent.
- [Expedia Rapid](https://developers.expediagroup.com/docs/products/rapid/setup/getting-started) requires partner approval, an API key and shared secret. Its evaluation endpoint and production access are separate.

None of these alternatives has been checked for coverage of the demo destinations or integrated with the frontend.

## Access and base URLs

- Production: `https://api.ratehawk.com`
- Sandbox: `https://api-sandbox.ratehawk.com`
- Authentication: HTTP Basic over HTTPS.
- Official docs: [integration guide](https://docs.emergingtravel.com/docs/integration-guide/), [authorization](https://docs.emergingtravel.com/docs/fundamentals/authorization/), [API best practices](https://docs.emergingtravel.com/docs/best-practices-for-apiv3/).

The RateHawk geo-search adapter uses the real HTTPS API with Basic authentication and explicit `RATEHAWK_MODE=sandbox` or `production`. It is outside the generic provider registry and requires partner credentials. It has not been validated against a real sandbox account; production readiness and provider-specific filters remain unverified.

## Documented request flow

These are `POST` endpoints with JSON bodies under the production base URL:

| Step | Path | Purpose |
|---|---|---|
| Region autocomplete | `/api/b2b/v3/search/multicomplete/` | Resolve a city/region or hotel query to provider IDs |
| Region search | `/api/b2b/v3/search/serp/region/` | Search rates by region, dates and room guests |
| Coordinate search | `/api/b2b/v3/search/serp/geo/` | Search rates near coordinates |
| Hotel ID search | `/api/b2b/v3/search/serp/hotels/` | Search a supplied list of hotel IDs |
| Hotel page | `/api/b2b/v3/search/hp/` | Fetch rates for a selected hotel |
| SERP prebook | `/api/b2b/v3/serp/prebook/` | Recheck a rate using its `search_hash` |
| Hotel page prebook | `/api/b2b/v3/hotel/prebook/` | Recheck a selected rate using its `book_hash` |
| Booking form | `/api/b2b/v3/hotel/order/booking/form/` | Validate a booking attempt and obtain payment options |
| Booking finish | `/api/b2b/v3/hotel/order/booking/finish/` | Start booking; this operation is asynchronous |
| Booking status | `/api/b2b/v3/hotel/order/booking/finish/status/` | Poll the booking result |
| Cancellation | `/api/b2b/v3/hotel/order/cancel/` | Request cancellation under the applicable tariff rules |

The flow branches after search: SERP rates use `search_hash` with SERP prebook; hotel-page rates use `book_hash` with hotel-page prebook. Confirm hash validity, retry rules, required fields and booking semantics against the current partner docs. Show any changed price or conditions and obtain user confirmation before continuing.

## Integration requirements

- Preserve dates, local arrival time, residency, currency and per-room guest composition. Hotel check-in must use the actual arrival date, which can differ from the transport departure date.
- Keep provider IDs and hashes server-side; they are not AI-authored preferences.
- Parse the provider's exact response schemas and normalize taxes, payment choices, cancellation penalties, meal plans and amount scales before calculating totals.
- Do not report availability, price or booking success without a successful provider response.
- The existing trip schema keeps hotel preferences. Both adapters currently support one explicitly assigned room, return limitations for unverified filters and never complete a reservation.
