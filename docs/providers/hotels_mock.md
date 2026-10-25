# hotels_mock — отели (наш сервис)

Бесплатного API отелей с реальными ценами нет: Hotellook закрыт (404), ETG/RateHawk — только по договору. Поэтому **мы сами делаем сервис** с тем же контрактом, что у **ETG / RateHawk B2B API v3** (их документация публична: docs.emergingtravel.com):

- **отели реальные:** названия и координаты из OpenStreetMap;
- **цены и тарифы синтетические**, детерминированные: одинаковый запрос даёт одинаковую цену.

Если позже появится настоящий поставщик, его подключают адаптером, без переписывания остального кода.

**Статус:** контракт и конфиги готовы, **сервис ещё не написан** (шаг 11 плана). Smoke-проверка его пропускает.

## Общее

- **Адрес:** `http://provider-service:8080/hotels` — меняется в `providers.json`.
- **Авторизация:** HTTP Basic (`HOTELS_KEY_ID:HOTELS_KEY`), как в ETG.
- **Ручки:** все `POST` с телом JSON.
- **Обёртка ответа** как в ETG: `{"data": …, "debug": null, "error": null, "status": "ok"}`, при ошибке — `"status": "error", "error": "<код>"`.
- **Заголовки лимитов:** `X-RateLimit-SecondsNumber`, `X-RateLimit-RequestsNumber`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`.

## Ограничения (скопированы из ETG, проверяются тегами и `Check()`)

| Что | Правило |
|---|---|
| `checkin`, `checkout` | `YYYY-MM-DD`; заезд не позже чем через 730 дней; выезд после заезда и **не позже 30 дней** |
| `residency` | ISO 3166-1 alpha-2 **строчными** (`ru`) |
| `currency` | ISO 4217 заглавными (`RUB`) |
| `guests` | один объект = один номер; 1–9 номеров; взрослых 1–6 на номер; детей до 4 на номер, возраст 0–17 |
| `hid` | целое, до 10 цифр |
| `hash` (prebook) | 1–256 символов, живёт 24 часа после страницы отеля |
| `search_hash` | живёт 38 минут |
| `partner_order_id` | UUID, уникальный для каждой попытки бронирования |
| `user.comment` | 1–256 символов |
| `payment_type.type` | `now` \| `hotel` \| `deposit` |

## Порядок вызовов

```
multicomplete (регион) → search_region / search_geo / search_hotels (список, 1–2 дешёвых тарифа)
  → hotelpage (все тарифы выбранного отеля) → prebook (перепроверка цены; изменение показать пользователю!)
  → booking_form → booking_finish → booking_status (опрашивать до финального статуса)   … → cancel
```

## Ручки

| Ручка | Путь | Тело (обязательные — *) | Пример конфига |
|---|---|---|---|
| `multicomplete` | `/api/b2b/v3/search/multicomplete/` | `query`*, `language` → до 5 регионов и 5 отелей | `hotels_mock.multicomplete.json` |
| `search_region` | `/api/b2b/v3/search/serp/region/` | `checkin`*, `checkout`*, `residency`*, `guests`*, `region_id`*, `language`, `currency`, `hotels_limit`, `filter`, `upsells` | `hotels_mock.search_region.json` |
| `search_geo` | `/api/b2b/v3/search/serp/geo/` | то же, но вместо `region_id`: `latitude`*, `longitude`*, `radius`* (м, ≤70000) | `hotels_mock.search_geo.json` |
| `search_hotels` | `/api/b2b/v3/search/serp/hotels/` | то же, но вместо `region_id`: `hids`* (1–300) | `hotels_mock.search_hotels.json` |
| `hotelpage` | `/api/b2b/v3/search/hp/` | поля поиска + `hid`* | `hotels_mock.hotelpage.json` |
| `prebook` | `/api/b2b/v3/hotel/prebook/` | `hash`*, `price_increase_percent` (0–100) | `hotels_mock.prebook.json` |
| `booking_form` | `/api/b2b/v3/hotel/order/booking/form/` | `partner_order_id`*, `book_hash`*, `language`*, `user_ip`* | `hotels_mock.booking_form.json` |
| `booking_finish` | `/api/b2b/v3/hotel/order/booking/finish/` | `user`* `{email*, phone, comment}`, `partner`* `{partner_order_id*}`, `language`*, `rooms`* `[{guests:[{first_name*, last_name*, is_child, age}]}]`, `payment_type`* `{type*, amount*, currency_code*}` | `hotels_mock.booking_finish.json` |
| `booking_status` | `/api/b2b/v3/hotel/order/booking/finish/status/` | `partner_order_id`* | `hotels_mock.booking_status.json` |
| `cancel` | `/api/b2b/v3/hotel/order/cancel/` | `partner_order_id`* | `hotels_mock.cancel.json` |

Поля `filter`:
- `star_rating[]` — 0–5;
- `kind[]`;
- `meal_type` — `nomeal` \| `breakfast` \| `half-board` \| `full-board` \| `all-inclusive`;
- `price_from`, `price_to` — **цена за всё проживание**, `price_to` > `price_from`;
- `has_transfer`.

Поля `upsells`: `early_checkin.time`, `late_checkout.time` — оба в формате `HH:MM`.

Пример запроса поиска рядом с Кремлём (`configs/examples/hotels_mock.search_geo.json`):

```json
{"checkin": "2026-12-12", "checkout": "2026-12-14", "residency": "ru", "language": "ru", "currency": "RUB",
 "latitude": 55.7987, "longitude": 49.1059, "radius": 2000, "hotels_limit": 20,
 "guests": [{"adults": 2, "children": []}],
 "filter": {"star_rating": [3, 4, 5], "meal_type": "breakfast", "price_to": 16000}}
```

## Ответ поиска (контракт, который реализуем; по образцу ETG)

```json
{
  "status": "ok", "error": null, "debug": null,
  "data": {
    "hotels": [
      {
        "hid": 1000001,
        "rates": [
          {
            "search_hash": "s-1000001-std-bb-20261212-20261214",
            "room_name": "Стандартный двухместный",
            "daily_prices": ["7100.00", "7100.00"],
            "meal_data": {"value": "breakfast", "has_breakfast": true, "no_child_meal": false},
            "allotment": 3,
            "amenities_data": ["wifi"],
            "payment_options": {"payment_types": [{
              "type": "now", "amount": "14200.00", "currency_code": "RUB",
              "show_amount": "14200.00", "show_currency_code": "RUB",
              "cancellation_penalties": {
                "free_cancellation_before": "2026-12-10T21:00:00",
                "policies": [
                  {"start_at": null, "end_at": "2026-12-10T21:00:00", "amount_charge": "0.00", "amount_show": "0.00"},
                  {"start_at": "2026-12-10T21:00:00", "end_at": null, "amount_charge": "7100.00", "amount_show": "7100.00"}
                ]
              },
              "tax_data": {"taxes": []}
            }]}
          }
        ]
      }
    ]
  }
}
```

`hotelpage` возвращает те же тарифы, но с `book_hash` и `match_hash`. Для `prebook` нужен `book_hash`. Время отмены (`free_cancellation_before`, `start_at`, `end_at`) — в **UTC**.

## Поведение мока, которое нужно реализовать

- **Отели:** реальные из OSM для ~30 городов; разовая выгрузка через Overpass с паузами или из выгрузки Geofabrik.
- **Цены:** детерминированные, `seed = hid + даты`: уровень цен города × звёзды × сезон × шум.
- **Prebook:** в 5–10% случаев меняет цену, чтобы агент умел показывать изменение пользователю.
- **Занятость:** часть дат занята («sold out»).
- **Бронирование:** статус `processing` → `ok`.

## Маппинг из `TripRequest`

| Поле `TripRequest` | Параметр |
|---|---|
| `destination.city` | `region_id` через `multicomplete` или `latitude`/`longitude` |
| `dates.departure_date` | `checkin` — **дата прибытия** (после ночного поезда — следующий день) |
| `dates.return_date` | `checkout` |
| `travelers.adults`, `children_ages` | `guests[]` — раскидать по номерам (≤6 взрослых, ≤4 детей) |
| `travelers.residency` | `residency` |
| `hotel.star_rating_min` | `filter.star_rating` |
| `hotel.meal` | `filter.meal_type` |
| `hotel.price_per_night_max` | `filter.price_to` = цена за ночь × число ночей |
| `hotel.early_checkin`, `late_checkout` | `upsells` |
