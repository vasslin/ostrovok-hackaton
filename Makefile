.PHONY: run build test smoke help e2e test-ai doctor
.DEFAULT_GOAL := run
MODE ?= demo
ENV_FILE ?= .env
WEB_PORT ?= 0
AI_HTTP_PORT ?= 0
API_URL ?= http://127.0.0.1:8080
RUN_FLAGS ?=
help:
	@printf '%s\n' 'make: UI, gateway, все сервисы и настоящий Qwen' 'make build: сборка для текущей ОС и архитектуры' 'make test: тесты всех Go-модулей и go vet' 'make smoke: проверка запущенного gateway' 'make test-ai: реальная нейронка через gateway' 'make e2e: изолированный сценарий с демопровайдерами' 'make RUN_FLAGS=-random-ports: запуск на свободных портах' 'Для Qwen сначала установите Ollama и travel-qwen:4b по LOCAL_QWEN.md.'
run:
	go run ./cmd/dev run -full -mode "$(MODE)" -env-file "$(ENV_FILE)" -web-port "$(WEB_PORT)" -ai-http-port "$(AI_HTTP_PORT)" $(RUN_FLAGS)
build:
	go run ./cmd/dev build
test:
	go run ./cmd/dev test
smoke:
	go run ./cmd/dev smoke -url "$(API_URL)"
e2e:
	go run ./cmd/dev e2e
test-ai:
	go run ./cmd/dev ai-smoke -url "$(API_URL)"

doctor:
	go run ./cmd/dev doctor -env-file "$(ENV_FILE)"
