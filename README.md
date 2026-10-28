# Travel AI service — «О!дин запрос»

**Этап 1 и ранжирование:** канонический TripRequest, уточнения, идеи и типизированные кандидаты доступны через `/v1/planning/*`, `/v1/ranking/*` и `godnota.ai.v1.AIService`. [Контракт, примеры и интеграция](docs/STAGE1.md).

**Основной запуск теперь бесплатный локальный: Ollama + Qwen3 4B.** Пошаговая инструкция для macOS, Windows и Linux: [LOCAL_QWEN.md](LOCAL_QWEN.md). Примеры настройки Groq ниже относятся к альтернативному облачному режиму.

AI-компонент для travel-service: переводит сообщение путешественника в конфиг поиска и ранжирует отели, переданные provider-service. Реальные билеты и отели должен получать provider-service. Бронирование и полный пользовательский сценарий не входят в этот компонент.

Настройки по умолчанию находятся в Compose; `.env.example` содержит пустые значения по требованиям сдачи. Подготовка `secrets.env.age`: [SECRETS.md](docs/SECRETS.md). Основные новые примеры и формат ответа описаны в [STAGE1.md](docs/STAGE1.md); упрощённые ручки ниже оставлены для совместимости.

Интеграция с PR #4: travel-service загружает контекст перед обработкой и сохраняет новые сообщения, summary и конфиг после успешного ответа. [Запуск с session-service и PostgreSQL](docs/SESSION_INTEGRATION.md).

## Выбор модели

Проверено по документации 26.10.2026:

| Вариант | Бесплатная квота | Вывод |
| --- | --- | --- |
| Groq, `openai/gpt-oss-20b` | 30 запросов/мин, 1 000 запросов/день, 8 000 токенов/мин, 200 000 токенов/день | Основной вариант: достаточно для 30–40 коротких вызовов |
| OpenRouter, бесплатные модели | До 50 запросов/день без покупки кредитов | Резерв: небольшой запас на тесты, повторы и демо |
| Gemini API | Зависит от модели и проекта; актуальный лимит виден в AI Studio | Не закладываем неподтверждённую квоту в план |

Источники: [Groq Free Plan Limits](https://console.groq.com/docs/rate-limits), [возможности GPT OSS 20B](https://console.groq.com/docs/model/openai/gpt-oss-20b), [OpenRouter: бесплатное использование](https://openrouter.ai/blog/tutorials/any-coding-agent/), [Gemini: лимиты](https://ai.google.dev/gemini-api/docs/rate-limits).

Лимиты общие для организации/аккаунта, а не на каждого пользователя. У Groq точные значения нужно проверить в [консоли](https://console.groq.com/settings/limits). Бесплатная квота не гарантирует доступность или качество. Проверьте регистрацию и доступность из сети, где будет демо; доступ с вашей машины здесь не подтверждён.

Расчёт: 7 пользователей × 6 вызовов = 42 вызова. Если каждый укладывается в 4 000 суммарных токенов, это около 168 000 токенов/день. Лимит 200 000 оставляет запас 32 000 токенов. Reasoning и длинный контекст также расходуют токены: измерьте `usage` на реальном сценарии. 8 000 токенов/мин могут ограничить параллельное демо раньше дневной квоты.

**Сообщение пользователя и запрос к модели — разные единицы.** Предлагаемый сценарий: 1 вызов на извлечение параметров, 1–3 на уточнения/изменения, 1 на ранжирование. Если после каждого из 6 сообщений вызывать оба endpoint, получится до 12 LLM-вызовов. Сервис делает один вызов на endpoint, без скрытых повторов, поиска и платных инструментов. Автоматического переключения провайдера нет.

## Быстрый старт

Для локального запуска установите Ollama и скачайте Qwen по [инструкции](LOCAL_QWEN.md). Для альтернативного облачного запуска Groq нужны Docker/Compose, доступ в интернет и ключ Groq. Ориентир для контейнера: 1 CPU, 128–256 MB RAM; для сборки выделить Docker около 1 GB.

Для облачного режима задайте LLM_PROVIDER=groq, LLM_BASE_URL=https://api.groq.com/openai/v1, LLM_MODEL=openai/gpt-oss-20b и LLM_TIMEOUT=45s.

1. Создать аккаунт [GroqCloud](https://console.groq.com/) и ключ на [API Keys](https://console.groq.com/keys). Оставаться на Free Plan.
2. Скопировать настройки и записать ключ в `.env`:

```bash
cp .env.example .env
# LLM_API_KEY=<ваш ключ Groq>
# SERVICE_API_KEY=<секрет для вызовов из backend>
docker compose up --build
```

Сервис: http://localhost:8080. В облачном режиме без ключа контейнер запускается, `/health` работает, но AI-endpoint возвращают 503. Никаких моковых рекомендаций вместо модели нет. `.env` исключён из Git. Ключ модели хранится только в backend, не во frontend. Перед публикацией обязательно задать SERVICE_API_KEY; пустое значение отключает авторизацию. В текущем Compose порт доступен на всех интерфейсах машины.

## API и интеграция

- `GET /` — переход в Swagger UI.
- `GET /docs` — Swagger UI: примеры, Authorize, Execute, ответы и время запроса.
- `GET /openapi.json` — OpenAPI 3.0.3 для импорта в Postman/Insomnia и генерации клиентов.
- `GET /health` — процесс работает, `{"status":"ok"}`.
- `GET /ready` — 200, если ключ задан; не проверяет его действительность или доступность провайдера.
- `POST /v1/config` — сообщение + текущий конфиг → обновлённый конфиг, `missing_fields`, максимум 2 вопроса.
- `POST /v1/rank` — конфиг + 1–10 отелей → все переданные ID в порядке убывания score, объяснения и риски.

Backend обращается из Compose к `http://ai-service:8080`, с хоста — к `http://localhost:8080`. Для защищённых POST передаётся `Authorization: Bearer <SERVICE_API_KEY>`.

```bash
curl http://localhost:8080/v1/config \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <SERVICE_API_KEY>' \
  -d '{"message":"Из Москвы в Сочи с 10 по 17 ноября 2026, двое взрослых, до 80000 рублей, тихий отель у моря","config":{}}'
```

Ожидаемая структура ответа (точные пожелания и формулировки зависят от модели):

```json
{
  "config": {
    "origin": "Москва", "destination": "Сочи",
    "departure_date": "2026-11-10", "return_date": "2026-11-17",
    "adults": 2, "children_ages": [], "budget": 80000,
    "currency": "RUB", "preferences": ["тихий отель", "у моря"]
  },
  "missing_fields": [], "questions": []
}
```

Для следующего сообщения передавайте предыдущий `config` целиком. Сервис stateless: историю хранит user-service/travel-service. Относительные даты интерпретируются от текущей даты UTC; для неоднозначных условий модель должна уточнять. Пустой `children_ages` означает отсутствие указанных возрастов; если нужны детские тарифы, travel-service должен уточнить состав детей.

```bash
curl http://localhost:8080/v1/rank \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <SERVICE_API_KEY>' \
  -d '{"config":{"budget":80000,"currency":"RUB","preferences":["у моря"]},"hotels":[{"id":"hotel-1","name":"Морской отель","details":{"total_price":75000,"currency":"RUB","distance_to_beach_m":200,"cancellation":"non-refundable"}},{"id":"hotel-2","name":"Тихая бухта","details":{"total_price":78000,"currency":"RUB","distance_to_beach_m":500,"cancellation":"free until 2026-11-08"}}]}'
```

Структура: `{"results":[{"id":"hotel-1","score":85,"reason":"...","risks":["..."]}, ...]}`. ID возвращаются ровно по одному разу; несуществующие ID и некорректные оценки отклоняются. В `details` backend передаёт компактные факты с единицами измерения: цена всего проживания, валюта, даты, отель, длительность/пересадки, отмена, ограничения. Примеры отелей выше синтетические.

Связка: travel-service → `/v1/config` → уточнения → provider-service → `/v1/rank` → map-service/frontend. Цены, наличие и состыковки проверяются кодом backend и API поставщика; оценка LLM не заменяет эти проверки. Произвольный текст поставщиков не следует считать доверенным. Prompt содержит ограничения, но защита от prompt injection не гарантирована.

Ошибки: 400 — неверный вход; 401 — неверный сервисный секрет; 429 — лимит провайдера (пробрасывается Retry-After, если он задан); 502 — ошибка провайдера, неполный/невалидный ответ; 503 — ключ не задан; 504 — таймаут модели (включая очередь локального режима); 413 — тело больше 32 KiB; 415 — Content-Type не application/json. Вход ограничен 32 KiB, сообщение 6 000 Unicode-символов, выход 2 048 токенами. JSON проверяется по типам, даты/числа — кодом. Требуемые для поиска поля вычисляются сервисом. Полная семантическая корректность извлечения зависит от модели и требует проверки на ваших запросах.

## Проверка

```bash
cd ai-service
go test ./...
go vet ./...
```

Тесты используют локальный тестовый HTTP-провайдер, не тратят бесплатную квоту. Проверяются разбор/ранжирование, неверные ID и даты, плохой JSON, авторизация, отсутствие ключа и проброс 429. Интеграция с настоящей моделью требует ключа и отдельного запуска примеров выше.

```bash
docker compose ps
curl http://localhost:8080/health
curl http://localhost:8080/ready
docker compose logs ai-service
```

Смена провайдера через env: для OpenRouter `LLM_BASE_URL=https://openrouter.ai/api/v1`, `LLM_MODEL=openrouter/free`, ключ OpenRouter. Роутер выбирает бесплатную модель; поддержка JSON mode и качество зависят от выбранного endpoint, требуется отдельный smoke test. Для воспроизводимого демо лучше фиксировать конкретный доступный ID с суффиксом `:free`.

Публичный хостинг не настроен: этот репозиторий поднимает API локально через Docker. Для внешнего размещения нужен выбранный сервер/хостинг и секреты; бесплатность модели не означает бесплатность сервера. Для сдачи всей команды ещё понадобятся интеграция остальных сервисов и ссылка на скринкаст.

## Тестирование через Swagger UI

Откройте http://localhost:8080/docs. Файлы Swagger UI 5.17.14 обслуживаются самим сервисом, доступ к CDN браузеру не нужен. Если задан `SERVICE_API_KEY`, нажмите **Authorize**, введите его значение без `Bearer` и подтвердите. Ключ провайдера в Swagger вводить не нужно.

Разверните `POST /v1/config`, измените готовый JSON и нажмите **Execute**. Для изменения параметров скопируйте полученный `config` в следующий запрос. Аналогично проверьте `/v1/rank` на готовом примере из двух синтетических отелей. Примеры ответов в документации иллюстративные; Execute вызывает настоящую модель и расходует квоту. Авторизация в браузере не сохраняется после перезагрузки.

После изменения `.env` выполните `docker compose up -d --force-recreate`. При локальном запуске без Docker выполняйте `go run ./cmd/server` из `ai-service`, заранее экспортировав переменные окружения; `.env` автоматически читает Compose. Для другого рабочего каталога задайте `WEB_DIR` с абсолютным путём к `ai-service/web`.

Сервис проверяет настройки при запуске и завершает активные запросы при SIGTERM/SIGINT. Расход токенов доступен в `docker compose logs ai-service` как `model usage`; сообщения пользователей и ключи в лог не записываются.

## gRPC

HTTP/Swagger слушает `8080`, gRPC — `9090` (настройка `GRPC_PORT`). Общая бизнес-логика находится в `internal/service`, транспорты в `internal/transport/httpapi` и `internal/transport/grpcapi`, запуск в `internal/app`. В `cmd/server` только `main.go`.

Контракт: `ai-service/api/travel/v1/travel.proto`. Методы: `travel.v1.TravelAI/ExtractConfig`, `travel.v1.TravelAI/RankHotels`. Go-клиент сгенерирован в том же каталоге. Включены reflection и стандартный gRPC health. Внутри Compose адрес `ai-service:9090`; на хосте `localhost:9090`. Транспорт plaintext для внутренней сети/локального демо; при внешнем размещении используйте TLS на ingress.

Если задан SERVICE_API_KEY, передавайте metadata `authorization: Bearer <SERVICE_API_KEY>`. Health/reflection доступны без секрета, как HTTP-документация. Устанавливайте deadline на клиенте, например 50 секунд. Максимальный protobuf-запрос 32 KiB, JSON-представление также ограничено 32 KiB.

```bash
grpcurl -plaintext localhost:9090 list
grpcurl -plaintext -d '{"service":"travel.v1.TravelAI"}' localhost:9090 grpc.health.v1.Health/Check
grpcurl -plaintext -H 'authorization: Bearer <SERVICE_API_KEY>' \
  -d '{"message":"В Сочи на неделю в ноябре, двое взрослых","config":{}}' \
  localhost:9090 travel.v1.TravelAI/ExtractConfig
grpcurl -plaintext -H 'authorization: Bearer <SERVICE_API_KEY>' \
  -d '{"config":{"preferences":["у моря"]},"hotels":[{"id":"hotel-1","name":"Морской отель","details":{"total_price":75000,"currency":"RUB","check_in":"2026-11-10","check_out":"2026-11-17","distance_to_beach_m":200}},{"id":"hotel-2","name":"Тихая бухта","details":{"total_price":78000,"currency":"RUB","check_in":"2026-11-10","check_out":"2026-11-17","distance_to_beach_m":500}}]}' \
  localhost:9090 travel.v1.TravelAI/RankHotels
```

В Postman можно создать gRPC request к `localhost:9090` и загрузить методы через reflection. Swagger тестирует HTTP; оба транспорта вызывают один сервис. Старое поле `candidates` заменено на `hotels`; каждый отель содержит обязательные `id`, `name` и произвольный объект `details`. Результат содержит ID, score, reason, risks. Цена `total_price` — за весь период проживания, `price_per_night` — за ночь. Общий бюджет поездки не равен бюджету отеля; неизвестные условия должны попадать в риски.

Соответствие ошибок: неверный вход → InvalidArgument; секрет → Unauthenticated; квота → ResourceExhausted (trailer retry-after, если предоставлен); отсутствие ключа/ошибка провайдера → Unavailable; таймаут → DeadlineExceeded; отмена → Canceled.

Для регенерации protobuf (Docker из корня репозитория):

```bash
docker build -f ai-service/Dockerfile.proto -t godnota-protoc ai-service
docker run --rm -v "$PWD/ai-service:/src" godnota-protoc
```

Генератор использует зафиксированные protoc-gen-go v1.36.9 и protoc-gen-go-grpc v1.5.1. Сгенерированные файлы включены в репозиторий; обычная сборка сервиса не требует protoc.
