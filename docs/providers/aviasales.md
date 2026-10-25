# aviasales — цены на авиабилеты

**Travelpayouts / Aviasales Data API v3.** Отдаёт самые дешёвые билеты, которые пользователи Aviasales находили за последние ~2–7 дней. Это **кэш**, а не живое наличие мест. Для хакатона этого достаточно: бронирование мы имитируем.

**Статус:** ✓ проверено 2026-10-25, HTTP 200, 6 рейсов Москва → Казань на 12–14 декабря.

## Доступ

- **Токен:** бесплатный аккаунт партнёра на travelpayouts.com → токен в профиле. Переменная `TRAVELPAYOUTS_TOKEN`.
- **Как передавать:** заголовок `X-Access-Token` (или параметр `token`).
- **Без токена:** `401 Unauthorized`.
- **Лимиты:** официальные не выясняли; клиент ограничен 5 запросами в секунду (`providers.json`).

## Ручка `prices_for_dates`

`GET https://api.travelpayouts.com/aviasales/v3/prices_for_dates`

| Параметр | Обязат. | Формат | Пример |
|---|---|---|---|
| `origin` | да | IATA города или аэропорта, `^[A-Z]{3}$` | `MOW` |
| `destination` | да | то же | `KZN` |
| `departure_at` | нет | `YYYY-MM-DD` или `YYYY-MM` (весь месяц) | `2026-12-12` |
| `return_at` | нет | то же; не передавать для билета в одну сторону | `2026-12-14` |
| `one_way` | нет | `true`/`false`; `true` даёт по одному билету на группу дат | `false` |
| `direct` | нет | `true` — только прямые | `true` |
| `currency` | да | ISO 4217 **строчными** | `rub` |
| `sorting` | нет | `price` \| `route` | `price` |
| `unique` | нет | bool | |
| `limit` | нет | 1–100 (наш предел) | `30` |
| `page` | нет | ≥1 | `1` |

Конфиг: `configs/examples/aviasales.prices_for_dates.json`. Схема: `configs/providers/aviasales.prices_for_dates.schema.json`.

```bash
curl -s -H "X-Access-Token: $TRAVELPAYOUTS_TOKEN" \
  "https://api.travelpayouts.com/aviasales/v3/prices_for_dates?origin=MOW&destination=KZN&departure_at=2026-12-12&return_at=2026-12-14&one_way=false&direct=true&currency=rub&sorting=price&limit=30&page=1"
```

### Ответ (настоящий, сокращён; полный — `testdata/aviasales/prices_for_dates.json`)

```json
{
  "success": true,
  "currency": "rub",
  "data": [
    {
      "origin": "MOW", "destination": "KZN",
      "origin_airport": "SVO", "destination_airport": "KZN",
      "departure_at": "2026-12-12T20:35:00+03:00",
      "return_at": "2026-12-14T17:20:00+03:00",
      "airline": "N4", "flight_number": "59",
      "price": 9665,
      "transfers": 0, "return_transfers": 0,
      "duration": 215, "duration_to": 105, "duration_back": 110,
      "gate": "Авиасейлс",
      "link": "/search/MOW1212KZN14121?t=DP1797107700…&expected_price=9665"
    }
  ]
}
```

| Поле | Значение |
|---|---|
| `departure_at`, `return_at` | местное время со смещением; смещение не терять |
| `price` | целое в `currency`; при заданном `return_at` — цена **туда-обратно** |
| `duration_to`, `duration_back` | **минуты** в пути туда и обратно; `duration` — их сумма |
| `transfers`, `return_transfers` | число пересадок |
| `airline` | IATA-код авиакомпании (`N4` — Nordwind, `SU` — Аэрофлот) |
| `link` | относительный путь: дописывать к `https://www.aviasales.ru` |

## Маппинг из `TripRequest`

| Поле `TripRequest` | Параметр |
|---|---|
| `origin.city` | `origin` — IATA через [travelpayouts_places](travelpayouts_places.md) |
| `destination.city` | `destination` |
| `dates.departure_date` | `departure_at`, или `departure_month` → `departure_at=YYYY-MM` |
| `dates.return_date` | `return_at`, `one_way=false` |
| `transport.direct_only` | `direct` |
| `currency` | `currency` строчными |

Вызывать только если в `transport.modes` есть `plane`.

## Подводные камни

- **Это кэш.** Для редких направлений и дат ответ может быть пустым (`data: []`). Повторить с месяцем (`YYYY-MM`), потом брать синтетическую цену.
- **Цена за одного взрослого** (так устроен кэш Aviasales). На семью умножаем сами.
