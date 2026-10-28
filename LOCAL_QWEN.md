# Бесплатный локальный Qwen

Модель выполняется на вашем компьютере. API-ключ, банковская карта и платный облачный аккаунт не нужны. Дневной квоты запросов нет: нагрузка ограничена скоростью компьютера. После первоначального скачивания модель обрабатывает запросы без интернета и VPN. Для установки и скачивания нужен доступ к сайтам Ollama и реестру моделей; доступность загрузки из конкретной сети заранее не гарантируется.

Используется `qwen3:4b` (примерно 2,5 GB на диске). Ориентир: 16 GB RAM; на 8 GB нужно закрыть тяжёлые приложения и проверить доступную память. Скорость и качество надо проверить на сценариях команды. Не скачивайте веса в Git.

Команды ниже выполняются из корня клонированного репозитория. `travel-qwen:4b` — локальное имя Qwen3 4B с окном контекста 16 384 токена, заданным в `ai-service/ollama/Modelfile`. Дополнительного скачивания при создании этого имени нет.

## 1. Установить Ollama на компьютер

macOS с Homebrew:

```bash
brew install ollama
brew services start ollama
ollama pull qwen3:4b
ollama create travel-qwen:4b -f ai-service/ollama/Modelfile
```

Windows: установите Ollama с https://ollama.com/download/windows, запустите приложение, затем выполните в PowerShell:

```powershell
ollama pull qwen3:4b
ollama create travel-qwen:4b -f ai-service/ollama/Modelfile
```

Linux: установите Ollama по официальной инструкции https://docs.ollama.com/linux, запустите сервис, затем выполните `ollama pull qwen3:4b` и `ollama create travel-qwen:4b -f ai-service/ollama/Modelfile`.

На Apple Silicon используйте нативный Ollama: контейнер Docker не получает доступ к Metal так же, как процесс macOS. Homebrew-сервис запускается автоматически после входа пользователя. Альтернатива без автозапуска — `ollama serve` в отдельном терминале.

## 2. Запустить AI-сервис в Docker

Из корня репозитория:

```bash
cp .env.example .env
docker compose up --build -d
```

Если `.env` уже существует, сохраните свои секреты и измените только:

```dotenv
LLM_PROVIDER=ollama
LLM_BASE_URL=http://host.docker.internal:11434/v1
LLM_MODEL=travel-qwen:4b
LLM_TIMEOUT=180s
```

`LLM_API_KEY` в локальном режиме не используется и не отправляется в Ollama. `SERVICE_API_KEY` по-прежнему защищает HTTP/gRPC вашего AI-сервиса; задайте его перед доступом других пользователей.

Docker Desktop на macOS/Windows обращается к Ollama через `host.docker.internal`. На Linux Ollama по умолчанию слушает только loopback: для обращения из контейнера настройте слушающий адрес и firewall по https://docs.ollama.com/faq либо используйте вариант без Docker ниже. Не открывайте порт Ollama 11434 в интернет — у локального API нет встроенной авторизации.

Откройте http://localhost:8080/docs, нажмите Authorize при заданном SERVICE_API_KEY и выполните `/v1/planning/analyze` или `/v1/ranking/routes`. Новый формат и примеры описаны в docs/STAGE1.md. Первый ответ может идти дольше: модель загружается в память. HTTP — порт 8080, gRPC — 9090. gRPC-клиенту задавайте deadline около 190 секунд.

## 3. Запуск без Docker

Нужны Go 1.26+ и запущенный Ollama. В терминале из корня репозитория:

```bash
cd ai-service
export LLM_PROVIDER=ollama
export LLM_BASE_URL=http://localhost:11434/v1
export LLM_MODEL=travel-qwen:4b
export LLM_TIMEOUT=180s
go run ./cmd/server
```

PowerShell:

```powershell
cd ai-service
$env:LLM_PROVIDER="ollama"
$env:LLM_BASE_URL="http://localhost:11434/v1"
$env:LLM_MODEL="travel-qwen:4b"
$env:LLM_TIMEOUT="180s"
go run ./cmd/server
```

При нативном запуске приложение читает переменные окружения, а не `.env`. Файлы Swagger по умолчанию ищутся в `ai-service/web`; запускайте команду именно из `ai-service`.

## Проверка и ошибки

```bash
ollama list
ollama ps
curl http://localhost:11434/api/tags
curl http://localhost:8080/ready
docker compose logs ai-service
```

`/ready` подтверждает настройку провайдера, но не проверяет наличие скачанной модели. Настоящая проверка — вызов `/v1/config` и `/v1/rank` через Swagger.

- 502: проверьте, что Ollama запущен, модель скачана и URL доступен из контейнера. Невалидный JSON модели также возвращает 502.
- 504: запрос не успел за LLM_TIMEOUT. Очередь тоже входит в это время; увеличьте таймаут (до 10 минут) и deadline клиента.
- Первый запрос медленный: модель загружается; повторите после загрузки.
- Не хватает памяти: попробуйте `ollama pull qwen3:1.7b` и измените LLM_MODEL; меньшую модель отдельно проверьте на качестве извлечения и ранжирования.

Изменив `.env`, пересоздайте контейнер: `docker compose up -d --force-recreate`. После изменения LLM_TIMEOUT выше 180 секунд увеличьте `stop_grace_period` в Compose хотя бы до таймаута + 20 секунд.

В локальном режиме сервис отправляет запросы в модель по одному, отключает thinking и передаёт JSON Schema для структурированного ответа. Одновременные пользователи ждут в очереди. Модель не ищет отели в интернете: отели передаются в `hotels` вместе с предпочтениями в `config`. Ответы локальной модели могут отличаться от Groq; 4B — компромисс по памяти, а не гарантия качества большой облачной модели.

## Вернуться к облачному API

```dotenv
LLM_PROVIDER=groq
LLM_BASE_URL=https://api.groq.com/openai/v1
LLM_MODEL=openai/gpt-oss-20b
LLM_API_KEY=<ключ>
LLM_TIMEOUT=45s
```

Для других OpenAI-совместимых провайдеров используйте LLM_PROVIDER=openai и соответствующие URL, модель и ключ. Никакого автоматического переключения на платный API нет.

Источники: https://ollama.com/library/qwen3:4b, https://docs.ollama.com/api/openai-compatibility, https://docs.ollama.com/faq.

## Проверка двух задач одной командой

Из корня репозитория, после запуска AI-сервиса:

```bash
python3 scripts/check-local.py
```

В Windows используйте `python scripts/check-local.py`. Скрипт делает два настоящих локальных запроса на синтетических данных: извлечение параметров и сравнение двух отелей. Печатает время и ответы, проверяет основные поля и предпочтение подходящего отеля. Сервисный ключ читает из `.env`, но не печатает его. Это короткий smoke test, а не полная оценка качества модели.
