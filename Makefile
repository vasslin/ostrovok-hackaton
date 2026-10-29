.PHONY: run build test smoke help e2e test-ai doctor model up down seed
.DEFAULT_GOAL := run
MODE ?= demo
ENV_FILE ?= .env
WEB_PORT ?= 0
AI_HTTP_PORT ?= 0
API_URL ?= http://127.0.0.1:8080
RUN_FLAGS ?=
OLLAMA_HOST ?= http://127.0.0.1:11434
OLLAMA_MODEL ?= travel-qwen:4b
help:
	@printf '%s\n' 'make: Ollama, модель $(OLLAMA_MODEL), UI, gateway и сервисы' 'make model: установить Ollama, запустить её и создать модель, если её нет' 'make build: сборка для текущей ОС и архитектуры' 'make test: тесты всех Go-модулей и go vet' 'make smoke: проверка запущенного gateway' 'make test-ai: реальная нейронка через gateway' 'make e2e: изолированный сценарий с демопровайдерами' 'make RUN_FLAGS=-random-ports: запуск на свободных портах' 'Ollama ставится сама: Homebrew на macOS, install.sh на Linux, winget на Windows.'
model:
	@set -eu; \
	if ! command -v brew >/dev/null 2>&1; then \
		if [ -x /opt/homebrew/bin/brew ]; then eval "$$(/opt/homebrew/bin/brew shellenv)"; \
		elif [ -x /usr/local/bin/brew ]; then eval "$$(/usr/local/bin/brew shellenv)"; fi; \
	fi; \
	if ! command -v ollama >/dev/null 2>&1; then \
		os=$$(uname -s 2>/dev/null || printf '%s' "$${OS:-}"); \
		case "$$os" in \
			Darwin) \
				command -v brew >/dev/null 2>&1 || { printf '%s\n' 'Нужен Homebrew: https://brew.sh' >&2; exit 1; }; \
				printf '%s\n' 'install ollama (Homebrew)'; \
				brew install ollama; \
				export PATH="$$(brew --prefix)/bin:$$PATH" ;; \
			Linux) \
				printf '%s\n' 'install ollama (https://ollama.com/install.sh)'; \
				if [ "$$(id -u)" -eq 0 ]; then curl -fsSL https://ollama.com/install.sh | sh; \
				else curl -fsSL https://ollama.com/install.sh | sudo sh; fi ;; \
			MINGW*|MSYS*|CYGWIN*|Windows_NT) \
				command -v winget >/dev/null 2>&1 || { printf '%s\n' 'Нужен winget или https://ollama.com/download/windows' >&2; exit 1; }; \
				printf '%s\n' 'install ollama (winget)'; \
				winget install --id Ollama.Ollama -e --accept-source-agreements --accept-package-agreements; \
				if [ -n "$${LOCALAPPDATA:-}" ]; then export PATH="$$LOCALAPPDATA/Programs/Ollama:$$PATH"; fi ;; \
			*) printf '%s\n' "Неизвестная ОС: $$os. Установите Ollama: https://ollama.com/download" >&2; exit 1 ;; \
		esac; \
		command -v ollama >/dev/null 2>&1 || { printf '%s\n' 'Ollama установлена, но команда ollama не видна в PATH. Откройте новый терминал и запустите make ещё раз.' >&2; exit 1; }; \
	fi; \
	mkdir -p data/logs; \
	if ! curl -fsS -m 2 "$(OLLAMA_HOST)/api/tags" >/dev/null 2>&1; then \
		printf '%s\n' 'start ollama'; \
		started=0; \
		if command -v brew >/dev/null 2>&1; then \
			brew services start ollama && started=1 || true; \
		fi; \
		if [ "$$started" -eq 0 ]; then \
			nohup ollama serve >> data/logs/ollama.log 2>&1 & \
		fi; \
		i=0; \
		while ! curl -fsS -m 2 "$(OLLAMA_HOST)/api/tags" >/dev/null 2>&1; do \
			i=$$((i + 1)); \
			if [ "$$i" -gt 30 ]; then \
				printf '%s\n' 'Ollama не открыла порт 11434. Журнал: data/logs/ollama.log' >&2; \
				exit 1; \
			fi; \
			sleep 1; \
		done; \
	fi; \
	if ! curl -fsS -m 5 "$(OLLAMA_HOST)/api/tags" | grep -F -q '"$(OLLAMA_MODEL)"'; then \
		printf '%s\n' 'pull qwen3:4b'; \
		ollama pull qwen3:4b; \
		printf '%s\n' 'create $(OLLAMA_MODEL)'; \
		ollama create "$(OLLAMA_MODEL)" -f ai-service/ollama/Modelfile; \
	fi; \
	printf '%s\n' "model: $(OLLAMA_MODEL)"
run: model
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

up:
	docker compose up --build

down:
	docker compose down

seed:
	@printf '%s\n' 'Схема PostgreSQL создаётся при старте session-service. Демо-билеты и отели синтетические и отдельной загрузки не требуют.'
