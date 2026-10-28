# Одна команда из корня репозитория: make
# UI:  http://localhost:5173
# API: http://127.0.0.1:8080
# Ctrl+C останавливает фронт, ai-service, user-service и стек cmd/dev.

.PHONY: run build test smoke help

MODE ?= demo
ENV_FILE ?= .env
WEB_PORT ?= 5173
AI_HTTP_PORT ?= 8090
BIN_DIR := bin

STACK := session-service map-service provider-service route-service travel-service order-service gateway

.DEFAULT_GOAL := run

help:
	@printf '%s\n' \
	  'make            собрать и запустить UI, API, ai-service и user-service' \
	  'make build      положить бинарники в bin/' \
	  'make test       go run ./cmd/dev test' \
	  'make smoke      проверить уже запущенный API' \
	  'make MODE=live  тот же запуск с живыми провайдерами' \
	  'Открывайте UI как http://localhost:$(WEB_PORT) — gateway пускает этот origin.'

build:
	@mkdir -p $(BIN_DIR)
	@for name in $(STACK); do \
		echo build $$name; \
		CGO_ENABLED=0 go build -o $(BIN_DIR)/$$name ./$$name/cmd/app; \
	done
	@echo build user-service
	@CGO_ENABLED=0 go build -o $(BIN_DIR)/user-service ./user-service/cmd/app
	@echo build ai-service
	@CGO_ENABLED=0 go build -o $(BIN_DIR)/ai-service ./ai-service/cmd/server

test:
	go run ./cmd/dev test

smoke:
	go run ./cmd/dev smoke

run:
	@set -eu; \
	root="$$(pwd)"; \
	mkdir -p "$$root/data/logs"; \
	side_pids=""; \
	cleanup() { \
		for pid in $$side_pids; do \
			kill "$$pid" 2>/dev/null || true; \
			pkill -P "$$pid" 2>/dev/null || true; \
		done; \
	}; \
	trap cleanup EXIT INT TERM; \
	if [ -f "$(ENV_FILE)" ]; then \
		while IFS= read -r line || [ -n "$$line" ]; do \
			case "$$line" in ''|\#*) continue ;; esac; \
			case "$$line" in export\ *) line=$${line#export } ;; esac; \
			key=$${line%%=*}; val=$${line#*=}; \
			key=$$(printf '%s' "$$key" | tr -d '[:space:]'); \
			case "$$key" in ''|*[!A-Za-z0-9_]*) continue ;; esac; \
			case "$$val" in \"*\") val=$${val#\"}; val=$${val%\"} ;; \'*\') val=$${val#\'}; val=$${val%\'} ;; esac; \
			export "$$key=$$val"; \
		done < "$(ENV_FILE)"; \
	fi; \
	if command -v node >/dev/null 2>&1; then \
		PORT=$(WEB_PORT) node frontend/serve.mjs >> "$$root/data/logs/frontend.log" 2>&1 & \
	else \
		python3 -m http.server $(WEB_PORT) --bind 127.0.0.1 --directory frontend/dist >> "$$root/data/logs/frontend.log" 2>&1 & \
	fi; \
	side_pids="$$side_pids $$!"; \
	GRPC_PORT=50053 go run ./user-service/cmd/app >> "$$root/data/logs/user-service.log" 2>&1 & \
	side_pids="$$side_pids $$!"; \
	APP_PORT=$(AI_HTTP_PORT) GRPC_PORT=50055 WEB_DIR="$$root/ai-service/web" \
		go run ./ai-service/cmd/server >> "$$root/data/logs/ai-service.log" 2>&1 & \
	side_pids="$$side_pids $$!"; \
	wait_port() { \
		name="$$1"; port="$$2"; pid="$$3"; i=0; \
		while [ "$$i" -lt 120 ]; do \
			if python3 -c "import socket,sys; socket.create_connection(('127.0.0.1', int(sys.argv[1])), 0.2).close()" "$$port" 2>/dev/null; then \
				return 0; \
			fi; \
			if ! kill -0 "$$pid" 2>/dev/null; then \
				printf '%s завершился. Журнал: data/logs/%s.log\n' "$$name" "$$name" >&2; \
				tail -n 40 "$$root/data/logs/$$name.log" >&2 || true; \
				return 1; \
			fi; \
			i=$$((i + 1)); sleep 1; \
		done; \
		printf '%s не открыл порт %s. Журнал: data/logs/%s.log\n' "$$name" "$$port" "$$name" >&2; \
		tail -n 40 "$$root/data/logs/$$name.log" >&2 || true; \
		return 1; \
	}; \
	set -- $$side_pids; \
	wait_port frontend $(WEB_PORT) "$$1"; \
	wait_port user-service 50053 "$$2"; \
	wait_port ai-service 50055 "$$3"; \
	printf 'UI:  http://localhost:%s\n' $(WEB_PORT); \
	printf 'API: http://127.0.0.1:8080\n'; \
	printf 'Журналы: data/logs/\n'; \
	go run ./cmd/dev run -mode "$(MODE)" -env-file "$(ENV_FILE)"
