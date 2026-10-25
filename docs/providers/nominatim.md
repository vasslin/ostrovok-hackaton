# nominatim — геокодинг и места по типу

**OpenStreetMap Nominatim.** Две роли:
1. **Геокодинг:** название города или адреса → координаты (для Resolver, погоды, Википедии).
2. **Поиск по типу места:** музеи, театры, кинотеатры, парки, рестораны в городе. Дополняет [Википедию](wikipedia.md), где у кинотеатров и ресторанов статей обычно нет.

**Статус:** ✓ обе ручки проверены 2026-10-25, HTTP 200. **Ключ не нужен.**

## Правила (обязательно)

- **Не чаще 1 запроса в секунду** — в `providers.json` стоит `rps: 1`.
- **Понятный `User-Agent`** — задан в `providers.json`.
- **Кэшировать результаты** (у нас на 30 дней); массовые выгрузки запрещены.
- **Атрибуция:** «© OpenStreetMap contributors».

## Ручка `search` — геокодинг

`GET https://nominatim.openstreetmap.org/search`

| Параметр | Обязат. | Формат | Пример |
|---|---|---|---|
| `q` | да | 1–255 символов | `Казань` |
| `format` | да | `json` \| `jsonv2` \| `geojson` | `jsonv2` |
| `limit` | нет | 1–40 | `1` |
| `accept-language` | нет | язык ответа | `ru` |
| `countrycodes` | нет | ISO alpha-2 строчными через запятую | `ru` |

```bash
curl -s -A "godnota-hackathon/0.1" -G "https://nominatim.openstreetmap.org/search" \
  --data-urlencode "q=Казань" --data-urlencode "format=jsonv2" --data-urlencode "limit=1" --data-urlencode "accept-language=ru"
```

Ответ (настоящий, сокращён; полный — `testdata/nominatim/search.json`):

```json
[{"lat": "55.7946485", "lon": "49.1115022", "category": "place", "type": "city", "name": "Казань",
  "display_name": "Казань, городской округ Казань, Татарстан, Приволжский федеральный округ, Россия",
  "boundingbox": ["55.6031357", "55.9307390", "48.8205846", "49.3812469"]}]
```

**`lat` и `lon` приходят строками**, их нужно переводить в числа.

## Ручка `search_amenity` — места по типу

Тот же адрес `/search`, но со структурными параметрами вместо `q` (их нельзя смешивать с `q`).

| Параметр | Обязат. | Формат | Пример |
|---|---|---|---|
| `amenity` | да | тип места: `museum`, `theatre`, `cinema`, `zoo`, `park`, `restaurant`, `cafe`… | `museum` |
| `city` | да | город | `Казань` |
| `country` | нет | страна | `Россия` |
| `format` | да | `jsonv2` | `jsonv2` |
| `limit` | нет | 1–40 | `40` |
| `accept-language` | нет | | `ru` |

Конфиг: `configs/examples/nominatim.search_amenity.json`.

```bash
curl -s -A "godnota-hackathon/0.1" -G "https://nominatim.openstreetmap.org/search" \
  --data-urlencode "amenity=museum" --data-urlencode "city=Казань" --data-urlencode "country=Россия" \
  --data-urlencode "format=jsonv2" --data-urlencode "limit=40" --data-urlencode "accept-language=ru"
```

Ответ (настоящий, 2 из 17 музеев; полный — `testdata/nominatim/search_amenity.json`):

```json
[
  {"name": "Национальный музей Республики Татарстан", "category": "tourism", "type": "museum", "lat": "55.7958566", "lon": "49.1096731"},
  {"name": "Музей А. М. Горького и Ф. И. Шаляпина", "category": "tourism", "type": "museum", "lat": "55.7935621", "lon": "49.1296637"}
]
```

Проверка 2026-10-25 по Казани: `museum` — 17 мест, `theatre` — 5, `cinema` — **всего 2**.

## Подводные камни

- **Неполнота.** Кинотеатров находится мало. Для полного списка — разовая выгрузка из Overpass (шаг 11 плана).
- **Мусор в выдаче:** туда попадают остановки с названием места («Театр…», `type: bus_stop`). Оставлять только нужные `type`.
- **Не больше 40 результатов** на запрос.
