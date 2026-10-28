# AI-сервис: этап 1 и контракт ранжирования

В полном стеке прямой AI HTTP доступен на http://127.0.0.1:8090, внутренний gRPC — ai-service:50055. Порты 8080/9090 ниже относятся к отдельному запуску AI-модуля. Полный UI запускается по README.md.

Актуальная интеграция сверена с `origin/main` (`adc122a`): `configs/trip_request.schema.json`, `configs/ai_response.schema.json`, `api/common/v1/trip.proto`, `api/provider/v1/parts.proto`. Копии JSON Schema лежат в `ai-service/schemas`. HTTP — 8080, gRPC — 9090. Локальная модель — Ollama `travel-qwen:4b`, инструкция в `LOCAL_QWEN.md`.

## Этап 1

- `POST /v1/planning/analyze` — сообщение, текущий конфиг и контекст.
- `POST /v1/planning/update` — правки, ответы на вопросы, выбор идеи.
- `POST /v1/planning/refine` — попытка завершения без выдумывания недостающих фактов.
- `POST /v1/planning/recommendations` — 2–3 идеи, не проверенные предложения.

Каждый вызов делает максимум один запрос к модели, без скрытых повторов. Валидация некорректных входных ответов выполняется до LLM. Контекст хранит session-service: вызовы чтения/сохранения выполняет travel-service, см. SESSION_INTEGRATION.md. Прямой AI API остаётся stateless.

`config` принимает сырой частичный TripRequest либо обёртку:

```json
{
  "version": "1.0",
  "request": {
    "schema_version": "1.0",
    "locale": "ru",
    "currency": "RUB",
    "origin": {"city": "Москва"},
    "destination": {"city": "Сочи"},
    "dates": {"departure_date": "2026-11-10", "return_date": "2026-11-17"},
    "travelers": {"adults": 2, "pets": [{"kind": "dog", "weight_kg": 5}]},
    "hotel": {"rooms": 1, "free_cancellation": true}
  },
  "resolved": {}
}
```

Legacy `config.trip` принимается для совместимости; одновременно `trip` и `request` запрещены. Неоднозначные старые поля `flexible_dates=true` и `max_wait_hours` отклоняются при преобразовании — используйте канонический request. Поддерживается одно выбранное направление.

Ответ analyze/update/refine по умолчанию имеет **ровно четыре поля**, как в схеме main:

```json
{
  "status": "questions",
  "config": {"schema_version": "1.0", "locale": "ru", "currency": "RUB", "origin": {"city": "Москва"}, "dates": {}, "travelers": {"adults": 0}, "completeness": {"missing_required": ["/destination/city", "/dates/departure_date", "/dates/return_date", "/travelers/adults"]}},
  "advice": [],
  "questions": [
    {"id": "destination_city", "field": "/destination/city", "text": "В какой город хотите поехать?", "required": true},
    {"id": "dates_departure_date", "field": "/dates/departure_date", "text": "Укажите точную дату отправления в формате ГГГГ-ММ-ДД.", "required": true}
  ]
}
```

`config` в ответе — сам TripRequest. Во время уточнения он частичный: `adults: 0` означает неизвестное значение; его нельзя передавать в поиск. `generate` выставляет код после проверки полей, дат, состава гостей и ограничений. Даты должны находиться в пределах года, пребывание — до 30 ночей. Для поиска нужна точная дата и выбранный город; одного месяца или списка идей недостаточно. Для поездки туда-обратно/с проживанием требуется дата возвращения либо число ночей, из которого код выводит дату. Бюджет необязателен. При неизвестном возрасте упомянутых детей нужно уточнение.

`advice` означает, что пользователь просит идеи и ещё не выбрал город. Вызывайте recommendations отдельно: он возвращает `advice` с `id`, `city`, `reason`, а также подробным описанием, `source=ai_generated_idea` и конфигом. Идея не содержит подтверждённых цен, наличия, погоды или условий въезда. Выбор отправляется как `selected_idea: {"id":"...","city":"Казань"}` либо полной идеей из recommendations.

Чтобы получить метаданные, добавьте `?include_metadata=true`. Тогда ответ — `{result, metadata}`: `result` соответствует четырём полям, `metadata` содержит `summary`, `missing_fields`, `assumptions`, `warnings`, `config_wrapper`. В следующий вызов передавайте `result.config` или `metadata.config_wrapper`, а `metadata.summary` — в `context.summary`. Это отдельный режим диагностики; обычный ответ дополнительных полей не имеет.

Вопросы используют JSON-пути `/dates/departure_date`, `/travelers/adults`. Ответы передаются в порядке приоритета через `answers: [{question_id, text}]` вместе с соответствующими `questions`. Максимум два отображаемых вопроса, 20 входных вопросов/ответов. История: до 20 сообщений и 3000 символов summary; recommendations принимает до пяти прошлых сессий.

Новая модель данных сохраняет питомцев, гражданство, гибкость дат, требования к отелю и досугу. Извлечение основных полей использует существующий проверяемый LLM-конвейер; литеральный разбор обрабатывает явные числа, даты, детей, питомцев, число ночей/номеров, отмену и часть удобств. Для редких требований используйте структурированные поля request. Неоднозначные формулировки требуют проверки пользователем. `resolved` в модель не передаётся; при изменении request сбрасывается, и backend должен разрешить коды заново. `completeness` вычисляется кодом.

## Ранжирование

`POST /v1/ranking/routes`, `/v1/ranking/join-options`, `/v1/ranking/final-routes` используют единый механизм проверки и оценки готовых кандидатов. Он не строит маршруты и стыковки. Route-service формирует кандидатов и подтверждает стыковки отдельно. Пример запроса — `examples/ranking/routes.json`, он также встроен в Swagger.

Вход: `config`, `context`, `parts`, `candidates`, `limit`, `current_date`. `parts` совпадает с нормализованным `SearchPartsResponse`: transport_parts, hotel_parts, errors, limitations. Кандидат содержит ID и ссылки `transport_part_ids`, `hotel_part_ids`, `connections_verified`. Можно передавать кандидатов только с отелями; полная цена поездки тогда будет неизвестна из-за отсутствия транспорта. Максимум 20 кандидатов и 100 частей каждого вида; весь запрос ограничен 32 KiB и в HTTP, и в gRPC.

Деньги передаются в минимальных единицах, `amount_minor` — JSON-строка int64, как в protobuf. Допускается целое JSON-число, но клиентам рекомендуется строка для сохранения точности. Дробные значения, экспоненциальная запись и переполнение отклоняются. Для RUB/USD/EUR 100 минимальных единиц составляют одну основную.

- Транспорт: `per_person_one_way` или `per_person_round_trip`; цена умножается только на взрослых. Для детей отдельного тарифа в текущем контракте нет — итог остаётся неизвестным.
- Отель: `stay_all_rooms`, полная стоимость проживания всех номеров; повторно на гостей или ночи не умножается.
- Неизвестная, нулевая, несопоставимая цена, разные валюты или неполное покрытие поездки → `price_complete=false`, `total_price` отсутствует.
- Итоговая сумма считается кодом с защитой от переполнения. Конвертации валют нет. Жёсткий общий бюджет проверяется только при сопоставимой полной сумме; другие области бюджета помечаются для проверки.
- Хронология, даты, города при наличии данных, пересадки, вид транспорта, звёзды и бесплатная отмена проверяются до модели. Нарушения попадают в `rejected`.
- Источник, время получения, предупреждения, ошибки поставщиков и limitations сохраняются в `parts`. Синтетика/кэш, неподтверждённые стыковки, дополнительные требования к отелю и питомцам отмечаются в risks/`requires_verification`. Это не подтверждение наличия или бронирования.

Модель возвращает только ID и субъективный `score` 0–100 для каждого допустимого кандидата. Новые, повторные, пропущенные ID и недопустимые оценки дают 502. Причины и риски формирует код. `limit=0` — все допустимые результаты, 1–20 — верхние N; отклонённые кандидаты остаются в rejected. Все отклонённые кандидаты возвращаются без вызова модели.

## gRPC

Контракт — `ai-service/api/ai/v1/ai.proto`, сервис `godnota.ai.v1.AIService`:

`GenerateOrUpdateConfig`, `UpdateConfig`, `RefineConfig`, `GenerateRecommendations`, `RankRoutes`, `RankJoinOptions`, `RankFinalRoutes`.

gRPC использует обёртку `TripConfig.request/resolved` и сохраняет отдельные поля метаданных в ответах этапа 1. `TripConfig` и `SearchPartsResponse` соответствуют main; сообщения кандидатов и результатов добавлены в AI-контракт этой ветки, их надо использовать при генерации клиентов в общем проекте. Три Rank RPC пока выполняют одинаковую проверку и оценку переданного набора.

Старый `travel.v1.TravelAI` и HTTP `/v1/config`, `/v1/rank` сохранены. У них прежние упрощённые модели — для новой интеграции используйте новые ручки. Доступ защищается `Authorization: Bearer SERVICE_API_KEY` и gRPC metadata `authorization`. Пустой ключ отключает проверку.

Генерация Go-кода:

```sh
docker build -f ai-service/Dockerfile.proto -t godnota-protoc ai-service
docker run --rm -v "$PWD/ai-service:/src" godnota-protoc
```

## Проверка

```sh
cd ai-service
go test -race ./...
go vet ./...
```

После запуска Ollama и контейнера, из корня репозитория:

```sh
python3 scripts/check-stage1.py
curl -H 'Content-Type: application/json' --data-binary @examples/ranking/routes.json http://localhost:8080/v1/ranking/routes
```

При заданном SERVICE_API_KEY добавьте Bearer к curl. Скрипт читает ключ из локального .env без вывода его значения. Отчёт со своими тестовыми данными записывает в `/tmp/godnota-stage1-smoke.json`. Старый `STAGE1_REPORT.json` фиксирует предыдущий legacy-контракт и не является проверкой новой версии.

Подготовка зашифрованных настроек для организаторов — `docs/SECRETS.md`.
