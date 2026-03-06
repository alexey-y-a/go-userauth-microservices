MODULE_PATH := github.com/alexey-y-a/go-userauth-microservices
PROJECT_NAME := go-userauth-microservices
DOCKER_COMPOSE := docker compose

# Цель по умолчанию
.DEFAULT_GOAL := help

.PHONY: help
help:
	@echo "Available make targets:"
	@echo "  make test              - run all Go tests"
	@echo "  make fmt               - run gofmt on all Go files"
	@echo "  make tidy              - run go mod tidy"
	@echo "  make build             - build all service binaries (local)"
	@echo "  make run-auth          - run auth-service locally"
	@echo "  make run-user          - run user-service locally"
	@echo "  make run-gateway       - run gateway-service locally"
	@echo "  make docker-build      - build Docker images via docker compose"
	@echo "  make up                - start all services via docker compose"
	@echo "  make down              - stop all services via docker compose"
	@echo "  make logs              - show docker compose logs"
	@echo "  make test-integration  - run integration tests (Postgres, testcontainers)"
	@echo "  make helm-template     - render Helm manifests"
	@echo "  make helm-install      - install Helm chart (requires JWT_SECRET env)"

# ---------- Go команды ----------

.PHONY: test
test:
	go test ./...

.PHONY: fmt
fmt:
	gofmt -w $$(find . -type f -name '*.go' -not -path "./vendor/*")

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: build
build:
	mkdir -p bin
	go build -o bin/auth-service ./services/auth-service/cmd/auth
	go build -o bin/user-service ./services/user-service/cmd/user
	go build -o bin/gateway-service ./services/gateway-service/cmd/gateway

# ---------- Локальный запуск без Docker ----------

.PHONY: run-auth
run-auth:
	go run ./services/auth-service/cmd/auth

.PHONY: run-user
run-user:
	go run ./services/user-service/cmd/user

.PHONY: run-gateway
run-gateway:
	go run ./services/gateway-service/cmd/gateway

# ---------- Docker / Docker Compose ----------

.PHONY: docker-build
docker-build:
	$(DOCKER_COMPOSE) build

.PHONY: up
up:
	$(DOCKER_COMPOSE) up -d

.PHONY: down
down:
	$(DOCKER_COMPOSE) down

.PHONY: logs
logs:
	$(DOCKER_COMPOSE) logs -f

# ---------- Интеграционные тесты ----------

.PHONY: test-integration
test-integration:
	go test ./services/auth-service/internal/storage/postgres -tags=integration -v

# ---------- Helm ----------

.PHONY: helm-template
helm-template:
	helm template userauth ./deploy/helm/go-userauth-microservices

.PHONY: helm-install
helm-install:
	@if [ -z "$(JWT_SECRET)" ]; then \
		echo "ERROR: JWT_SECRET is not set. Run 'export JWT_SECRET=your_secret' first"; \
		exit 1; \
	fi
	helm install userauth ./deploy/helm/go-userauth-microservices \
		--set jwt.secretValue="$(JWT_SECRET)"


