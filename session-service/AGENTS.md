# session-service: контекст AI-агента

Сервис отдаёт и обновляет контекст, который `travel-service` передаёт модели. Вызывай его по gRPC `godnota.session.v1.SessionService`, не через gateway.

## Что входит в контекст

`GetAgentContext(user_id, chat_id, previous_chats_limit)` возвращает:

- `chat_history` — сообщения этого чата по порядку добавления
- `chat_summary` — резюме этого чата
- `previous_chats_summary` — резюме других чатов того же пользователя, самые свежие первыми
- `cfg` — единственный конфиг поездки этого чата

`previous_chats_limit = 0` означает 4. Отрицательное значение — ошибка. Значение больше 20 обрезается до 20. Пустые резюме в предыдущие чаты не попадают. Текущий чат туда не входит.

Если чата ещё нет, история и резюме пустые. Это не ошибка. Конфиг при этом может уже быть.

## Как обновлять

`SaveContext(user_id, chat_id, messages?, summary?, cfg?)` меняет только переданные поля:

- непустой `messages` дописывается в конец истории. Повторная отправка той же реплики создаст дубль. У каждого сообщения обязательны `role` и `text`. Пустой список историю не трогает.
- `summary`, если поле есть в запросе, заменяет резюме этого `chat_id`. Пустая строка очищает резюме. Если поля нет, резюме остаётся прежним.
- `cfg`, если поле есть в запросе, записывает конфиг. У чата всегда один конфиг: повторная передача заменяет предыдущий. Если поля нет, конфиг остаётся прежним.

`user_id` и `chat_id` обязательны. Чат и конфиг принадлежат первому пользователю, который их создал. Чужой `user_id` для того же `chat_id` возвращает `PermissionDenied`.

`chat_id` — идентификатор диалога. В сценарии планирования это `session_id` из `CreateSession`.

`user_id` бери из запроса. Gateway после входа подставляет идентификатор вошедшего пользователя во все вызовы `travel-service` (`PlanTrip`, `SubmitAnswers`, `SelectIdea`, `GetConfig`, `UpdateConfig`, `ConfirmConfig`, `SearchRoutes`, `GetRoutes`), а также в `CreateSession` и `CreateOrder`. Планирование передаёт эту пару в `GetAgentContext` и `SaveContext` и не запрашивает владельца отдельным `GetSession`. Чужой `user_id` в теле HTTP-запроса gateway отклоняет до вызова сервиса.

## Где лежат данные

История, резюме и выборка предыдущих резюме идут в репозиторий истории. Конфиг — в отдельный репозиторий и читается по паре user id + chat id.

`SESSION_REPOSITORY=memory` хранит всё в процессе. `SESSION_REPOSITORY=postgres` пишет в PostgreSQL (`SESSION_POSTGRES_DSN`). В `docker-compose.yml` включён postgres. Схема создаётся при старте.

## Куда смотреть в коде

- контракт: `api/session/v1/session.proto`
- сценарии: `internal/usecase/session/context.go`
- память: `internal/repository/history/memory`, `internal/repository/tripconfig/memory`
- postgres: те же пакеты `postgres` и схема `internal/repository/pgdb/schema.sql`
- выбор хранилища: `internal/adapter/storage`

`CreateSession` и `GetSession` относятся к сессии и не пишут в эти два репозитория. Контекст агента читается через `GetAgentContext` и меняется через `SaveContext`.
