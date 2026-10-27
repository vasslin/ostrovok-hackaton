# Hotel search API research

The repository has no hotel mock and no executable hotel adapter. The real option found is ETG / RateHawk B2B API v3. Access requires a partner agreement and credentials; none are configured here, so hotel search is intentionally absent from `configs/providers.json` and `BuildTripPlan`.

## Access and base URLs

- Production: `https://api.ratehawk.com`
- Sandbox: `https://api-sandbox.ratehawk.com`
- Authentication: HTTP Basic over HTTPS.
- Official docs: [integration guide](https://docs.emergingtravel.com/docs/integration-guide/), [authorization](https://docs.emergingtravel.com/docs/fundamentals/authorization/), [API best practices](https://docs.emergingtravel.com/docs/best-practices-for-apiv3/).

Do not point the old internal mock URL at RateHawk. The prior mock request structs were removed because they simplified fields and should not be treated as a production contract. Implement and validate a dedicated adapter against the partner documentation and sandbox before adding a registry entry.

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
- The existing trip schema keeps hotel preferences. Until a real adapter is available, the plan reports hotel search as unavailable and preserves those preferences.
