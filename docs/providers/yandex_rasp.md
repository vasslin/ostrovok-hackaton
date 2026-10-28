# yandex_rasp - расписания транспорта

**API Яндекс.Расписаний v3.0.** Настоящие расписания: самолёты, поезда, электрички, автобусы, **водный транспорт** (корабли, паромы, «Метеоры»), вертолёты. Цен почти нет.

**Статус:** ✓ проверено 2026-10-25, все 4 ручки HTTP 200.

## Доступ

- **Ключ:** бесплатный, в кабинете разработчика Яндекса (продукт «API Яндекс.Расписаний»). Переменная `YANDEX_RASP_API_KEY`. [Порядок доступа](https://yandex.ru/dev/rasp/doc/ru/concepts/access).
- **Как передавать:** параметр `apikey` (или заголовок `Authorization`).
- **Домен: `https://api.rasp.yandex-net.ru`.** Старый `api.rasp.yandex.net` может не работать.
- **Активация:** документация говорит, что ключ активен сразу после выпуска. При `429 «Ключ не активен»` проверить состояние продукта и ключа в кабинете; наблюдение с первым ключом не является общим правилом API.
- **Дневной лимит** смотреть в кабинете; клиент ограничен 2 запросами в секунду.
- **Атрибуция обязательна:** на каждом экране с данными расписаний непосредственно рядом с ними показать текст «Данные предоставлены сервисом Яндекс.Расписания» с активной ссылкой на `http://rasp.yandex.ru/`; размер и цвет шрифта - как у основного текста. Ручка `copyright` даёт материалы для оформления. [Условия использования, пункт 2.3.5](https://yandex.ru/legal/timetable_api/ru/).

## Ручка `search` - рейсы между двумя точками

`GET /v3.0/search/`

| Параметр | Обязат. | Формат | Пример |
|---|---|---|---|
| `from`, `to` | да | код Яндекса: город `c213`, станция `s2000003`; или другая система через `system` | `c213`, `c43` |
| `date` | нет | `YYYY-MM-DD`; без даты - только прямые рейсы на все дни | `2026-12-12` |
| `transport_types` | нет | `plane` \| `train` \| `suburban` \| `bus` \| `water` \| `helicopter`; без него - все | `train` |
| `transfers` | нет | `true` - добавить маршруты с пересадками | `false` |
| `system` | нет | `yandex` \| `iata` \| `sirena` \| `express` \| `esr` | |
| `limit`, `offset` | нет | `limit` ≤100 (по умолчанию 100) | `100` |
| `lang` | нет | `ru_RU` \| `uk_UA` | `ru_RU` |
| `format` | нет | `json` \| `xml` | `json` |
| `result_timezone` | нет | IANA; по умолчанию у каждой станции свой пояс | |

```bash
curl -s "https://api.rasp.yandex-net.ru/v3.0/search/?apikey=$YANDEX_RASP_API_KEY&from=c213&to=c43&date=2026-12-12&transport_types=train&transfers=false&limit=100&lang=ru_RU&format=json"
```

### Ответ (настоящий, один из 10 поездов; полный - `testdata/yandex_rasp/search.json`)

```json
{
  "search": {"from": {"type": "settlement", "title": "Москва", "code": "c213"},
             "to": {"type": "settlement", "title": "Казань", "code": "c43"}, "date": "2026-12-12"},
  "pagination": {"total": 10, "limit": 100, "offset": 0},
  "segments": [
    {
      "departure": "2026-12-12T20:40:00+03:00",
      "arrival": "2026-12-13T08:00:00+03:00",
      "duration": 40800.0,
      "from": {"code": "s2000003", "title": "Москва (Казанский вокзал)", "station_type": "train_station", "transport_type": "train"},
      "to": {"code": "s9623141", "title": "Казань-Пасс."},
      "thread": {"uid": "002J_0_2", "number": "002Й", "title": "Москва - Казань", "transport_type": "train",
                 "carrier": {"code": 112, "title": "РЖД/ФПК"}},
      "has_transfers": false,
      "stops": "",
      "tickets_info": {"et_marker": false, "places": []},
      "start_date": "2026-12-12"
    }
  ],
  "interval_segments": []
}
```

| Поле | Значение |
|---|---|
| `departure`, `arrival` | ISO 8601 со смещением, время станции |
| `duration` | **секунды**, дробное число (`40800.0` = 11 ч 20 мин) |
| `thread.uid` | id рейса, нужен для ручки `thread` |
| `tickets_info.places[]` | `{name, currency, price: {whole, cents}}` - **у всех 10 поездов пусто** |
| `interval_segments` | рейсы без точного времени («каждые 15 минут») |

## Ручка `nearest_settlement` - код города по координатам

`GET /v3.0/nearest_settlement/?lat=55.7879&lng=49.1233&distance=30&lang=ru_RU&format=json`

Параметры: `lat`, `lng` - обязательны; `distance` - радиус в км (у нас ≤50).

Так Resolver получает код Яндекса для города. Ответ (настоящий):

```json
{"type": "settlement", "title": "Казань", "code": "c43", "lat": 55.795793, "lng": 49.106585, "distance": 1.36}
```

## Ручка `thread` - остановки рейса

`GET /v3.0/thread/?uid=002J_0_2&date=2026-12-12&lang=ru_RU&format=json`

`uid` берётся из `search`. Ответ (настоящий, сокращён):

```json
{
  "number": "002Й", "title": "Москва - Казань", "days": "ежедневно",
  "stops": [
    {"station": {"title": "Москва (Казанский вокзал)", "code": "s2000003"}, "arrival": null, "departure": "2026-12-12 20:40:00", "stop_time": null},
    {"station": {"title": "Муром-1", "code": "s9612176"}, "arrival": "2026-12-13 01:04:00", "departure": "2026-12-13 01:06:00", "stop_time": 120},
    {"station": {"title": "Казань-Пасс.", "code": "s9623141"}, "arrival": "2026-12-13 08:00:00", "departure": null}
  ]
}
```

**Внимание: здесь другой формат времени** - `2026-12-12 20:40:00`, без `T` и без смещения. Разбирать в часовом поясе станции. `stop_time` - в секундах.

## Ручка `copyright` - атрибуция

`GET /v3.0/copyright/?format=json` → `{"copyright": {"text": "Данные предоставлены сервисом Яндекс.Расписания", "url": "http://rasp.yandex.ru/", "logo_…": …}}`. Выводить обязательный текст со ссылкой рядом с расписаниями на каждой странице, включая мобильную карту маршрута.

## Маппинг из `TripRequest`

| Поле `TripRequest` | Параметр |
|---|---|
| `origin.city`, `destination.city` | `from`, `to` - коды через `nearest_settlement` по координатам города |
| `dates.departure_date` | `date` прямого рейса |
| `dates.return_date` | `date` обратного рейса, `from` и `to` меняются местами |
| `transport.modes` | `transport_types`, один запрос на вид транспорта |
| `transport.direct_only` | `transfers=false` |

## Подводные камни

- **Цен нет** → цены на поезда, автобусы и паромы берём из синтетического `transport_mock`, а время - настоящее отсюда.
- **Нет поиска по месяцу.** Для `departure_month` запрашивать 2-3 конкретные даты.
- **Ночные рейсы** прибывают на следующий день: поезд 002Й уходит 12-го и прибывает 13-го. Заезд в отель = дата **прибытия**.
- **`search` с датой** возвращает и рейс в 00:20 следующего дня. Фильтровать по `start_date`, если нужно.
