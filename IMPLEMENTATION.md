# Каркас Godnota

Собран запускаемый скелет AI-агента для планирования поездки: protobuf-контракты, девять процессов и HTTP gateway. Бизнес-логика пока заглушки: доменный сервис пишет вход в лог и возвращает пустое значение, gRPC-контроллер всегда отвечает успехом.

Проверено локально: `go run` всех сервисов, запросы через gateway на `http://127.0.0.1:8080` возвращают HTTP 200 и пустой JSON. В логах видно и сам запрос, и вызовы по цепочке (например, поиск идёт в map, route, provider и ai). Docker daemon в этой среде не был запущен, поэтому `docker compose up --build` здесь не прогонялся.

## Контракты

Исходники лежат в `api/`, сгенерированный Go-код - в `pkg/generated`. Повторная генерация: `sh scripts/gen-proto.sh`.

| Сервис | RPC | Сообщения |
|---|---|---|
| session-service | CreateSession, GetSession, GetContext, SaveContext | сессия, история, `TripConfig` |
| user-service | GetUser, UpdatePreferences | профиль, язык, валюта |
| order-service | CreateOrder, GetOrder, PayOrder, CancelOrder | заказ и статусы |
| travel-service | PlanTrip, SubmitAnswers, SelectIdea, GetConfig, UpdateConfig, ConfirmConfig, SearchRoutes, GetRoutes | сценарий планирования |
| ai-service | GenerateOrUpdateConfig, UpdateConfig, RefineConfig, RankRoutes | пустые, поля позже |
| map-service | MapToProviderConfigs, MapRoutesToFrontend | пустые, поля позже |
| route-service | SearchRoutes | пустые, поля позже |
| provider-service | SearchTransport, SearchHotels, SearchActivities | пустые, поля позже |

Общая модель поездки (`TripConfig`: откуда, куда, даты, гости, бюджет, транспорт, жильё, досуг) описана в `api/common/v1/trip.proto`.

## Структура

Каждый микросервис - отдельный Go-модуль с гексагональной раскладкой:

- `cmd/app` - процесс
- `internal/controller` - gRPC, всегда `OK` и пустой ответ
- `internal/usecase` - доменный сервис, логирует запрос и возвращает нулевое значение
- `internal/entity` - доменные модели
- `internal/repository` - in-memory заглушка репозитория
- `internal/adapter` - исходящие gRPC-клиенты там, где сервис кого-то вызывает

`travel-service` оркестрирует session, ai, map и route. `route-service` на поиске вызывает provider. Gateway снаружи ходит только в session, travel, user и order.

Рядом: `gateway/`, `docker-compose.yml`, `.env`, `.env.example`, `Dockerfile` у каждого сервиса.

## Gateway

Реализация повторяет BFF из референсного проекта: таблица HTTP-маршрутов, сбор JSON из path и тела, `protojson` и unary gRPC. Swagger UI: [http://localhost:8080/docs](http://localhost:8080/docs), спецификация: `/openapi.json`. Корень `/` перенаправляет на `/docs`. `GET /health` отвечает `{"status":"ok"}`.

| Метод | Путь | Куда уходит |
|---|---|---|
| POST | `/api/sessions` | session.CreateSession |
| GET | `/api/sessions/{session_id}` | session.GetSession |
| POST | `/api/sessions/{session_id}/messages` | travel.PlanTrip |
| POST | `/api/sessions/{session_id}/answers` | travel.SubmitAnswers |
| POST | `/api/sessions/{session_id}/ideas` | travel.SelectIdea |
| GET | `/api/sessions/{session_id}/config` | travel.GetConfig |
| PATCH | `/api/sessions/{session_id}/config` | travel.UpdateConfig |
| POST | `/api/sessions/{session_id}/confirm` | travel.ConfirmConfig |
| POST | `/api/sessions/{session_id}/search` | travel.SearchRoutes |
| GET | `/api/sessions/{session_id}/routes` | travel.GetRoutes |
| POST | `/api/orders` | order.CreateOrder |
| GET | `/api/orders/{order_id}` | order.GetOrder |
| POST | `/api/orders/{order_id}/pay` | order.PayOrder |
| POST | `/api/orders/{order_id}/cancel` | order.CancelOrder |
| GET | `/api/users/{user_id}` | user.GetUser |
| PATCH | `/api/users/{user_id}` | user.UpdatePreferences |

## Запуск

```bash
cp .env.example .env
docker compose up --build
```

Gateway: [http://localhost:8080/docs](http://localhost:8080/docs).

Пример:

```bash
curl -s -X POST http://localhost:8080/api/sessions \
  -H 'content-type: application/json' \
  -d '{"user_id":"user-123"}'
```

Ответ заглушки: `{"session_id":""}`.

Логи входящих вызовов:

```bash
docker compose logs -f travel-service
docker compose logs -f session-service
```

Строка доменного сервиса выглядит так: `travel-service: plan trip session_id=... text=...`. Рядом gRPC-интерцептор пишет метод и payload.
