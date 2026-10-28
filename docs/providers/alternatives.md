# Additional provider research

Access and coverage below are based on provider documentation. Hotelbeds evaluation access was subsequently checked as described in the [run report](../evidence/2026-10-28/RUN_REPORT.ru.md); no live hotel inventory has been verified.

| Source | Useful endpoint | Access and limitations |
|---|---|---|
| [Hotelbeds hotels](https://developer.hotelbeds.com/documentation/hotels/booking-api/) | `POST /hotel-api/1.0/hotels` | Registration gives evaluation credentials; the [evaluation quota](https://developer.hotelbeds.com/documentation/getting-started/) is 50 requests/day. An evaluation search near Aveiro returned EUR rates; other destinations returned no results in the recorded checks. Production requires provider approval. |
| [Hotelbeds activities](https://developer.hotelbeds.com/documentation/activities/booking-api/availability/availability-/) | `POST /activity-api/3.0/activities/availability` | Search by destination or GPS and dates; the activities API needs its own evaluation key. Not implemented. |
| [Hotelbeds transfers](https://developer.hotelbeds.com/documentation/transfers/booking-api/search-availability/availability-simple/) | Availability search | Separate evaluation credentials and confirmed pickup/drop-off details are needed. Not implemented. |
| [Ostrovok B2B](https://docs-api.ostrovok.ru/docs/integration-guide/) | `POST /api/b2b/v3/search/serp/geo/` | Contract-specific Basic credentials; production certification. Its test-hotel booking can carry financial obligations: do not book during this demo. |
| [Booking.com Demand](https://developers.booking.com/demand/docs/open-api/3.3/demand-api/accommodations/accommodations/search) | `POST /3.3/accommodations/search` | Requires an API token and affiliate ID. Inventory, partner eligibility and demo-city coverage are unverified. |
| [Expedia Rapid](https://developers.expediagroup.com/docs/products/rapid/setup/getting-started) | Lodging availability | Requires partner approval, API key and signed requests; evaluation and production are separate. Coverage is unverified. |
| [Ticketmaster Discovery](https://developer.ticketmaster.com/products-and-docs/apis/discovery/) | `GET /discovery/v2/events.json` | API key required; events and optional price ranges, not a general cinema/museum catalogue. Country coverage must be checked; no implementation here. |
| [Yandex Rasp](https://yandex.ru/dev/rasp/doc/ru/reference/schedule-point-point) | `GET /v3.0/search/` | Schedules for trains and water transport. When present, `tickets_info.places[].price.whole/cents` and currency describe a ticket quote; an empty list does not imply a free ticket. |

Do not scrape public Nominatim for POI lists. Its [usage policy](https://operations.osmfoundation.org/policies/nominatim/) forbids systematic POI downloads. The supplied OSM snapshot is limited and attributed; use an appropriately licensed extract or a permitted Overpass instance to refresh it.
