# travelpayouts_places - город → IATA-код

**Автокомплит Travelpayouts.** Название города или аэропорта → IATA-код, координаты, страна и русские падежи (готовый текст «в Казань» для интерфейса). Нужен Resolver'у: AI пишет «Казань», а Aviasales нужен `KZN`. [Официальная документация](https://support.travelpayouts.com/hc/en-us/articles/360002322572-Autocomplete-API-for-countries-cities-and-airports-by-Aviasales).

**Статус:** ✓ проверено 2026-10-25, HTTP 200. **Ключ не нужен.**

## Ручка `places2`

`GET https://autocomplete.travelpayouts.com/places2`

| Параметр | Обязат. | Формат | Пример |
|---|---|---|---|
| `term` | да | 1-100 символов, кириллица допустима | `Казань` |
| `locale` | да | наша схема допускает `ru` \| `en`; API поддерживает и другие языки | `ru` |
| `types[]` | да | `city` \| `airport` \| `country`, можно несколько раз | `city` |

Конфиг: `configs/examples/travelpayouts_places.places2.json`.

```bash
curl -s -G "https://autocomplete.travelpayouts.com/places2" \
  --data-urlencode "term=Казань" --data-urlencode "locale=ru" --data-urlencode "types[]=city"
```

### Ответ (настоящий, первый элемент; полный - `testdata/travelpayouts_places/places2.json`)

```json
[
  {
    "type": "city",
    "code": "KZN",
    "name": "Казань",
    "country_code": "RU",
    "country_name": "Россия",
    "coordinates": {"lon": 49.1233294, "lat": 55.7878944},
    "cases": {"vi": "в Казань", "ro": "Казани", "da": "Казани", "pr": "Казани", "tv": "Казанью", "su": "Казань"}
  }
]
```

- **Выбор результата:** API возвращает `weight`, но официальная документация называет его служебным полем и не гарантирует порядок. Сопоставлять `type`, страну и название; при неоднозначности попросить пользователя выбрать город.
- **Падежи** `cases`: `su` - именительный, `ro` - родительный, `da` - дательный, `vi` - винительный с предлогом («в Казань»), `tv` - творительный, `pr` - предложный.

## Статический справочник

`GET https://api.travelpayouts.com/data/ru/cities.json`, без ключа, ~3,5 МБ. Скачать один раз и держать локально.

У каждого города:
- `code`, `name`, `coordinates`;
- `time_zone` - IANA, нужен для перевода времени;
- `has_flightable_airport`.

Рядом лежат `airports.json`, `airlines.json`, `countries.json`.

## Подводные камни

- `code` бывает кодом **города** (`MOW`), а не аэропорта (`SVO`). Aviasales принимает оба, город предпочтительнее.
- Маленьких городов без аэропорта может не быть, тогда `[]`. Для поездов код города берём у Яндекса ([yandex_rasp.md](yandex_rasp.md), `nearest_settlement`).
