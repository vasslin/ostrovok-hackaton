# aviasales - цены на авиабилеты

**Travelpayouts / Aviasales Data API v3.** Возвращает найденные пользователями Aviasales цены за последние 48 часов. Это **кэш**, а не живое наличие мест; для демо бронирование имитируется. [Официальное описание метода](https://support.travelpayouts.com/hc/en-us/articles/203956163-Aviasales-Data-API).

**Статус:** ✓ проверено 2026-10-25, HTTP 200, 6 рейсов Москва → Казань на 12-14 декабря.

## Доступ

- **Токен:** бесплатный аккаунт партнёра на travelpayouts.com → токен в профиле. Переменная `TRAVELPAYOUTS_TOKEN`.
- **Как передавать:** заголовок `X-Access-Token` (или параметр `token`).
- **Без токена:** `401 Unauthorized`.
- **Лимиты:** для `prices_for_dates` официально 600 запросов в минуту; наш лимит 5 запросов в секунду консервативнее. При HTTP 429 соблюдать паузу. [Лимиты Travelpayouts](https://support.travelpayouts.com/hc/en-us/articles/4402565416594-API-rate-limits).

## Ручка `prices_for_dates`

`GET https://api.travelpayouts.com/aviasales/v3/prices_for_dates`

| Параметр | Обязат. | Формат | Пример |
|---|---|---|---|
| `origin` | да | IATA города или аэропорта, `^[A-Z]{3}$` | `MOW` |
| `destination` | да | то же | `KZN` |
| `departure_at` | нет | `YYYY-MM-DD` или `YYYY-MM` (весь месяц) | `2026-12-12` |
| `return_at` | нет | то же; не передавать для билета в одну сторону | `2026-12-14` |
| `one_way` | нет | `true`/`false`; `true` даёт по одному билету на группу дат | `false` |
| `direct` | нет | `true` - только прямые | `true` |
| `currency` | да | ISO 4217 **строчными** | `rub` |
| `sorting` | нет | `price` \| `route` | `price` |
| `unique` | нет | bool | |
| `limit` | нет | 1-100 в нашей схеме; максимум API - 1000 | `30` |
| `page` | нет | ≥1 | `1` |

Конфиг: `configs/examples/aviasales.prices_for_dates.json`. Схема: `configs/providers/aviasales.prices_for_dates.schema.json`.

```bash
curl -s -H "X-Access-Token: $TRAVELPAYOUTS_TOKEN" \
  "https://api.travelpayouts.com/aviasales/v3/prices_for_dates?origin=MOW&destination=KZN&departure_at=2026-12-12&return_at=2026-12-14&one_way=false&direct=true&currency=rub&sorting=price&limit=30&page=1"
```

### Ответ (настоящий, сокращён; полный - `testdata/aviasales/prices_for_dates.json`)

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
| `departure_at`, `return_at` | ISO 8601; разбирать по смещению в ответе и сохранять исходное значение |
| `price` | целое в `currency`; при заданном `return_at` - цена **туда-обратно** |
| `duration_to`, `duration_back` | **минуты** в пути туда и обратно; `duration` - их сумма |
| `transfers`, `return_transfers` | число пересадок |
| `airline` | IATA-код авиакомпании (`N4` - Nordwind, `SU` - Аэрофлот) |
| `link` | относительный путь поиска; для внешнего перехода использовать домен Aviasales, соответствующий рынку, и проверять назначение URL. Официальный пример использует `https://www.aviasales.com/` |

## Маппинг из `TripRequest`

| Поле `TripRequest` | Параметр |
|---|---|
| `origin.city` | `origin` - IATA через [travelpayouts_places](travelpayouts_places.md) |
| `destination.city` | `destination` |
| `dates.departure_date` | `departure_at`, или `departure_month` → `departure_at=YYYY-MM` |
| `dates.return_date` | `return_at`, `one_way=false` |
| `transport.direct_only` | `direct` |
| `currency` | `currency` строчными |

Вызывать только если в `transport.modes` есть `plane`.

## Подводные камни

- **Это кэш.** Для редких направлений и дат ответ может быть пустым (`data: []`). Можно повторить с месяцем (`YYYY-MM`); если дальше используется синтетическая цена, маркировать её как демоданные и не выдавать за доступный билет.
- **Нет гарантии цены и наличия.** Перед реальным оформлением нужен отдельный поиск с актуальной ценой и подтверждение пользователя; Data API этого не делает.
- **Цена за одного взрослого** (так устроен кэш Aviasales). На семью умножаем сами.
