# Каркас Godnota

Запускаемый набор сервисов AI-агента для планирования поездки: protobuf-контракты, процессы и HTTP gateway. Контекст диалога хранит `session-service`. Планирование в `travel-service` читает и записывает его по `user_id` и `session_id`. Остальные сценарии (ai, map, route, provider, user, order и черновой путь `PlanTrip`) по-прежнему в основном логируют вызов и возвращают пустой ответ.

## Контракты

Исходники лежат в `api/`, сгенерированный Go-код - в `pkg/generated`. Повторная генерация: `sh scripts/gen-proto.sh`.

| Сервис | RPC | Сообщения |
|---|---|---|
| session-service | CreateSession, GetSession, GetAgentContext, SaveContext | сессия; история, резюме и `TripConfig` чата |
| user-service | GetUser, UpdatePreferences | профиль, язык, валюта |
| order-service | CreateOrder, GetOrder, PayOrder, CancelOrder | заказ и статусы |
| travel-service | PlanTrip, SubmitAnswers, SelectIdea, GetConfig, UpdateConfig, ConfirmConfig, SearchRoutes, GetRoutes | сценарий планирования, в каждом запросе есть `user_id` |
| ai-service | GenerateOrUpdateConfig, UpdateConfig, RefineConfig, RankRoutes | пустые, поля позже |
| map-service | MapToProviderConfigs, MapRoutesToFrontend | пустые, поля позже |
| route-service | SearchRoutes | пустые, поля позже |
| provider-service | SearchTransport, SearchHotels, SearchActivities | пустые, поля позже |

Общая модель поездки (`TripConfig`: откуда, куда, даты, гости, бюджет, транспорт, жильё, досуг) описана в `api/common/v1/trip.proto`.

`GetAgentContext` возвращает историю текущего чата, его резюме, резюме последних чатов того же пользователя (по умолчанию 4) и один конфиг поездки. `SaveContext` меняет только переданные поля: непустой `messages` дописывается в историю, переданный `summary` заменяет резюме, переданный `cfg` заменяет конфиг. `chat_id` в этом сценарии равен `session_id`. Как вызывать сервис, описано в `session-service/AGENTS.md`.

История и резюме лежат в репозитории истории, конфиг — в отдельном репозитории, один на чат. `SESSION_REPOSITORY=memory` держит их в процессе, `SESSION_REPOSITORY=postgres` — в PostgreSQL. В `docker-compose.yml` для этого поднят `session-db`. Запись сессии `CreateSession` / `GetSession` остаётся в памяти процесса и в эти репозитории не пишется.

## Структура

Каждый микросервис - отдельный Go-модуль с гексагональной раскладкой:

- `cmd/app` - процесс
- `internal/controller` - gRPC
- `internal/usecase` - сценарий
- `internal/entity` - доменные модели
- `internal/repository` - хранилище
- `internal/adapter` - исходящие клиенты там, где сервис кого-то вызывает

`travel-service` оркестрирует session, ai, map и route. `route-service` на поиске вызывает provider. Gateway снаружи ходит только в session, travel, user и order.

Рядом: `gateway/`, `docker-compose.yml`, `.env`, `.env.example`, `Dockerfile` у каждого сервиса.

## Gateway

Реализация повторяет BFF из референсного проекта: таблица HTTP-маршрутов, сбор JSON из path и тела, `protojson` и unary gRPC. Swagger UI: [http://localhost:8080/docs](http://localhost:8080/docs), спецификация: `/openapi.json`. Корень `/` перенаправляет на `/docs`. `GET /health` отвечает `{"status":"ok"}`.

Рабочие маршруты требуют `Authorization: Bearer`. Токен выдают `POST /api/auth/register` и `POST /api/auth/login`. Gateway подставляет `user_id` вошедшего пользователя в создание сессии, создание заказа и все travel-запросы. Если клиент прислал другой `user_id`, запрос отклоняется. Перед travel- и order-вызовом gateway проверяет, что сессия или заказ принадлежат этому пользователю.

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
curl -s -X POST http://localhost:8080/api/auth/register \
  -H 'content-type: application/json' \
  -d '{"email":"a@example.com","password":"long-demo-password"}'

curl -s -X POST http://localhost:8080/api/sessions \
  -H 'content-type: application/json' \
  -H 'authorization: Bearer <access_token>' \
  -d '{}'
```

`user_id` в тело создания сессии дописывает gateway. В ответе приходит `session_id`.

Логи входящих вызовов:

```bash
docker compose logs -f travel-service
docker compose logs -f session-service
```

Строка доменного сервиса выглядит так: `travel-service: plan trip session_id=... text=...`. Рядом gRPC-интерцептор пишет метод и payload.
