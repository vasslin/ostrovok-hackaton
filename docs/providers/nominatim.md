# nominatim - точечный геокодинг

**Публичный OpenStreetMap Nominatim** принимает точечные запросы, например название города или адрес → координаты для Resolver, погоды и Википедии. Его структурный `search` умеет искать по `amenity` и городу, но возвращает лучшие совпадения, а не список всех объектов. Использовать его как каталог музеев, театров и кинотеатров нельзя. [Официальная справка Search API](https://nominatim.org/release-docs/latest/api/Search/) и [политика публичного сервера](https://operations.osmfoundation.org/policies/nominatim/).

**Статус:** оба примера дали HTTP 200 2026-10-25. Это подтверждает формат запроса на тот момент, а не разрешение собирать POI по городам. **Ключ не нужен.**

## Правила (обязательно)

- **Не чаще 1 запроса в секунду суммарно для всего приложения**, включая всех пользователей и экземпляры сервиса. Значение `rps: 1` в `providers.json` нужно обеспечивать общим ограничителем, а не отдельным ограничителем на каждый экземпляр.
- **Идентифицирующий `User-Agent` с контактной ссылкой** задан в `providers.json`.
- **Кэшировать результаты** (у нас на 30 дней); запрещены автодополнение и систематические запросы для выгрузки всех POI в зоне. Для каталогов использовать заранее полученный OSM extract или собственный индекс.
- **Атрибуция:** «© OpenStreetMap contributors».

## Ручка `search` - геокодинг

`GET https://nominatim.openstreetmap.org/search`

| Параметр | Обязат. | Формат | Пример |
|---|---|---|---|
| `q` | да | 1-255 символов | `Казань` |
| `format` | да | `json` \| `jsonv2` \| `geojson` | `jsonv2` |
| `limit` | нет | 1-40 | `1` |
| `accept-language` | нет | язык ответа | `ru` |
| `countrycodes` | нет | ISO alpha-2 строчными через запятую | `ru` |

```bash
curl -s -A "godnota-hackathon/0.1 (https://github.com/ostrovok-hackathon-2026/godnota)" -G "https://nominatim.openstreetmap.org/search" \
  --data-urlencode "q=Казань" --data-urlencode "format=jsonv2" --data-urlencode "limit=1" --data-urlencode "accept-language=ru"
```

Ответ (настоящий, сокращён; полный - `testdata/nominatim/search.json`):

```json
[{"lat": "55.7946485", "lon": "49.1115022", "category": "place", "type": "city", "name": "Казань",
  "display_name": "Казань, городской округ Казань, Татарстан, Приволжский федеральный округ, Россия",
  "boundingbox": ["55.6031357", "55.9307390", "48.8205846", "49.3812469"]}]
```

**`lat` и `lon` приходят строками**, их нужно переводить в числа.

## Ручка `search_amenity` - пример формата, не источник каталога

Тот же адрес `/search`, но со структурными параметрами вместо `q` (их нельзя смешивать с `q`). Пример оставлен для воспроизводимости прототипа; автоматический обход по типам и городам для построения досуга нарушает политику публичного сервера. Подходящий источник для списка мест - локальный OSM extract, полученный вне живого пользовательского запроса.

| Параметр | Обязат. | Формат | Пример |
|---|---|---|---|
| `amenity` | да | тип места: `museum`, `theatre`, `cinema`, `zoo`, `park`, `restaurant`, `cafe`… | `museum` |
| `city` | да | город | `Казань` |
| `country` | нет | страна | `Россия` |
| `format` | да | `jsonv2` | `jsonv2` |
| `limit` | нет | 1-40 | `40` |
| `accept-language` | нет | | `ru` |

Конфиг: `configs/examples/nominatim.search_amenity.json`.

```bash
curl -s -A "godnota-hackathon/0.1 (https://github.com/ostrovok-hackathon-2026/godnota)" -G "https://nominatim.openstreetmap.org/search" \
  --data-urlencode "amenity=museum" --data-urlencode "city=Казань" --data-urlencode "country=Россия" \
  --data-urlencode "format=jsonv2" --data-urlencode "limit=40" --data-urlencode "accept-language=ru"
```

Ответ (настоящий, 2 из 17 музеев; полный - `testdata/nominatim/search_amenity.json`):

```json
[
  {"name": "Национальный музей Республики Татарстан", "category": "tourism", "type": "museum", "lat": "55.7958566", "lon": "49.1096731"},
  {"name": "Музей А. М. Горького и Ф. И. Шаляпина", "category": "tourism", "type": "museum", "lat": "55.7935621", "lon": "49.1296637"}
]
```

Проверка 2026-10-25 по Казани: `museum` вернул 17 совпадений, `theatre` - 5, `cinema` - 2. Это **не количество** таких мест в городе.

## Подводные камни

- **Неполнота по определению API.** Даже успешный ответ не даёт полного списка. Для демокаталога нужен офлайн-набор OSM объектов; для актуального большого объёма - свой индекс или сервис с разрешённым POI-поиском.
- **Мусор в выдаче:** туда попадают остановки с названием места («Театр…», `type: bus_stop`). Оставлять только нужные `type`.
- **Не больше 40 результатов** на запрос.
