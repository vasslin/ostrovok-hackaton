# open_meteo — погода и температура моря

**Open-Meteo.** Три ручки на трёх разных хостах (поэтому в `providers.json` три записи):

| Ключ в реестре | Что | Хост |
|---|---|---|
| `open_meteo.forecast` | прогноз на ≤16 дней | `api.open-meteo.com` |
| `open_meteo_archive.archive` | история погоды с 1940 года (ERA5) | `archive-api.open-meteo.com` |
| `open_meteo_marine.marine` | **температура моря** (история с 2022, прогноз 0–8 дней) | `marine-api.open-meteo.com` |

**Статус:** ✓ все три проверены 2026-10-25, HTTP 200. **Ключ не нужен.**

## Правила

- **Бесплатно для некоммерческого использования:** меньше 10 000 запросов в день, 5 000 в час, 600 в минуту. Коммерческий продукт или реклама требуют другого плана. [Условия Open-Meteo](https://open-meteo.com/en/terms).
- **Атрибуция:** «Weather data by Open-Meteo.com», CC BY 4.0.

## Какую ручку брать

- **Поездка в ближайшие 16 дней** → `forecast`.
- **Поездка дальше горизонта прогноза** → `archive` за те же даты прошлых лет как исторический ориентир; явно показать год и не называть эти значения прогнозом.
- **Пляжный отдых** → `marine` для текущего горизонта либо исторический `sea_surface_temperature` за тот же сезон с пометкой «история». [Официальный Marine API](https://open-meteo.com/en/docs/marine-weather-api).

## `forecast`

`GET https://api.open-meteo.com/v1/forecast`

Параметры:
- `latitude`*, `longitude`*;
- `daily` — через запятую: `temperature_2m_max,temperature_2m_min,precipitation_sum`; `hourly`;
- `timezone=auto`;
- `forecast_days` (1–16) или `start_date` / `end_date` (`YYYY-MM-DD`).

```bash
curl -s "https://api.open-meteo.com/v1/forecast?latitude=55.79&longitude=49.12&daily=temperature_2m_max,temperature_2m_min,precipitation_sum&timezone=auto&forecast_days=7"
```

Ответ (настоящий, Казань, сокращён):

```json
{
  "latitude": 55.79072, "longitude": 49.10105, "timezone": "Europe/Moscow",
  "daily_units": {"time": "iso8601", "temperature_2m_max": "°C", "temperature_2m_min": "°C", "precipitation_sum": "mm"},
  "daily": {
    "time": ["2026-10-25", "2026-10-26", "2026-10-27"],
    "temperature_2m_max": [10.9, 10.5, 10.7],
    "temperature_2m_min": [2.4, 1.7, 1.8],
    "precipitation_sum": [0.0, 0.0, 0.0]
  }
}
```

## `archive`

`GET https://archive-api.open-meteo.com/v1/archive`

Параметры: `latitude`*, `longitude`*, `start_date`*, `end_date`* (`YYYY-MM-DD`), `daily`, `hourly`, `timezone`.

```bash
curl -s "https://archive-api.open-meteo.com/v1/archive?latitude=55.79&longitude=49.12&start_date=2025-12-12&end_date=2025-12-14&daily=temperature_2m_max,temperature_2m_min,precipitation_sum&timezone=auto"
```

Ответ (настоящий: Казань, 12–14 декабря 2025):

```json
{"timezone": "Europe/Moscow",
 "daily": {"time": ["2025-12-12", "2025-12-13", "2025-12-14"],
           "temperature_2m_max": [0.8, 0.4, -12.0], "temperature_2m_min": [-1.5, -13.2, -14.3],
           "precipitation_sum": [4.6, 5.3, 0.2]}}
```

## `marine`

`GET https://marine-api.open-meteo.com/v1/marine`

Параметры:
- `latitude`*, `longitude`*;
- `hourly=sea_surface_temperature` (или `daily`);
- `start_date` / `end_date` либо `forecast_days` (0–8);
- `timezone`.

```bash
curl -s "https://marine-api.open-meteo.com/v1/marine?latitude=27.26&longitude=33.84&hourly=sea_surface_temperature&start_date=2025-11-01&end_date=2025-11-07&timezone=auto"
```

Ответ (настоящий: Хургада, 1–7 ноября 2025, 168 часовых значений; первые три):

```json
{"timezone": "Africa/Cairo",
 "hourly_units": {"time": "iso8601", "sea_surface_temperature": "°C"},
 "hourly": {"time": ["2025-11-01T00:00", "2025-11-01T01:00", "2025-11-01T02:00"],
            "sea_surface_temperature": [26.3, 26.3, 26.3]}}
```

Это настоящий ответ на «тёплое море в ноябре»: в Хургаде **26,3 °C**.

## Подводные камни

- **Время в ответе ISO 8601 без явного смещения** (`2025-11-01T00:00`) и с полем `timezone`. Для дат со сменой летнего времени документация рекомендует `timeformat=unixtime` и преобразование по IANA-зоне; `timezone=auto` само по себе не делает локальные метки однозначными.
- **Сетка ~5–10 км:** ответ приходит для ближайшей точки сетки, а не для твоих координат. Для моря брать точку **в воде**, а не на берегу.
- **Пропуски:** значения могут быть `null`, считать среднее без них.
