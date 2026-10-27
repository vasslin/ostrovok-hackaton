# wikipedia — досуг: что посмотреть рядом

**API Википедии, `generator=geosearch`.** Источник статей о местах рядом с точкой: название, координаты и ссылка; описание и миниатюра доступны не у каждой статьи. Используем вокруг отеля, остановки и пересадки.

**Статус:** ✓ проверено 2026-10-25, HTTP 200: 20 мест в 3 км от Казанского кремля, у всех фото, у 11 описание. **Ключ не нужен.**

## Доступ и правила

- **Без ключа.** Wikimedia просит понятный `User-Agent` — он задан в `providers.json`.
- **Атрибуция текста:** ссылка на статью и указание лицензии CC BY-SA 4.0. **Фотографии имеют собственные лицензии и авторов**: перед показом в карточке нужно получить страницу файла или лицензионные метаданные и указать автора, источник и лицензию для конкретного изображения. [Правила повторного использования Wikimedia Commons](https://commons.wikimedia.org/wiki/Commons:Reusing_content_outside_Wikimedia/licenses).
- **Лимиты:** с 2026 года для автоматического клиента только с корректным `User-Agent` указан ориентир 200 запросов в минуту и не более трёх одновременных запросов. В каталоге наш предел 3 запроса в секунду; соблюдать `Retry-After` при 429 и 503. [Официальные лимиты Wikimedia](https://www.mediawiki.org/wiki/Wikimedia_APIs/Rate_limits).

## Ручка `geosearch`

`GET https://ru.wikipedia.org/w/api.php`

| Параметр | Обязат. | Формат | Пример |
|---|---|---|---|
| `action` | да | всегда `query` | `query` |
| `generator` | да | всегда `geosearch` | `geosearch` |
| `ggscoord` | да | `lat\|lon` | `55.7987\|49.1059` |
| `ggsradius` | да | метры, 10–10000 | `3000` |
| `ggslimit` | нет | 1–50 (наш предел, чтобы у всех было фото) | `20` |
| `prop` | нет | что вернуть | `coordinates\|pageimages\|description\|info` |
| `inprop` | нет | `url` — добавить ссылку на статью | `url` |
| `piprop` | нет | `thumbnail` \| `original` | `thumbnail` |
| `pithumbsize` | нет | ширина фото, 50–1000 px | `400` |
| `format` | да | `json` | `json` |
| `formatversion` | нет | `2` — страницы списком | `2` |

Конфиг: `configs/examples/wikipedia.geosearch.json`.

```bash
curl -s -A "godnota-hackathon/0.1 (https://github.com/ostrovok-hackathon-2026/godnota)" -G "https://ru.wikipedia.org/w/api.php" \
  --data-urlencode "action=query" --data-urlencode "generator=geosearch" \
  --data-urlencode "ggscoord=55.7987|49.1059" --data-urlencode "ggsradius=3000" --data-urlencode "ggslimit=20" \
  --data-urlencode "prop=coordinates|pageimages|description|info" --data-urlencode "inprop=url" \
  --data-urlencode "piprop=thumbnail" --data-urlencode "pithumbsize=400" \
  --data-urlencode "format=json" --data-urlencode "formatversion=2"
```

### Ответ (настоящий, первое место; полный — `testdata/wikipedia/geosearch.json`)

```json
{
  "continue": {"cocontinue": "4055604|331392569", "continue": "||pageimages|description|info"},
  "query": {
    "pages": [
      {
        "pageid": 119678,
        "title": "Казанский кремль",
        "description": "древнейшая часть города и цитадель Казани",
        "coordinates": [{"lat": 55.79861, "lon": 49.10639, "primary": true, "globe": "earth"}],
        "thumbnail": {"source": "https://thumb.wikimedia.org/…/500px-Kazan_Kremlin_-_panoramio_%286%29.jpg…", "width": 400, "height": 266},
        "fullurl": "https://ru.wikipedia.org/wiki/%D0%9A%D0%B0%D0%B7%D0%B0%D0%BD%D1%81%D0%BA%D0%B8%D0%B9_%D0%BA%D1%80%D0%B5%D0%BC%D0%BB%D1%8C"
      }
    ]
  }
}
```

Ещё в выдаче: «Кул-Шариф — мечеть в Казани, Россия», «Благовещенский собор Казанского кремля», «Башня Сююмбике», «Хазинэ — музей изобразительного искусства в Казанском кремле».

| Поле | Значение |
|---|---|
| `title` | название места |
| `description` | короткое описание; **у части страниц его нет** |
| `coordinates[0].lat/lon` | координаты |
| `thumbnail.source` | фото для карточки |
| `fullurl` | ссылка на статью |
| `continue` | есть ещё страницы; передать его поля в следующий запрос, если нужно больше |

## Как использовать

- **Фильтровать по смыслу и типу объекта:** отсутствие `description` само по себе не доказывает, что статья не о достопримечательности; использовать его как сигнал для ранжирования или дополнять другим источником.
- **Порядок не по расстоянию.** Расстояние до отеля считаем сами по координатам.
- **Пересадки:** если пересадка ≥ `leisure.min_layover_hours`, искать вокруг вокзала или аэропорта. Оставлять только места, куда можно успеть туда и обратно (время из [osrm.md](osrm.md) × 2 + час запаса).

## Маппинг из `TripRequest`

| Интерес (`leisure.interests`) | Источник |
|---|---|
| `architecture`, `history`, достопримечательности | этот API |
| `museums` | этот API + локальный каталог OSM POI |
| `theatres`, `cinemas`, `parks`, `food`, `kids` | локальный каталог OSM POI по типу места; публичный Nominatim не подходит для систематического сбора объектов |
| `festivals`, `concerts` | бесплатного рабочего API нет → `events_mock` |
