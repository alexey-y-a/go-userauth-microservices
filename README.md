# Микросервисная система аутентификации и работы с пользователями, реализованная как набор независимых сервисов с общими библиотеками.

## Цель проекта

- Реализовать систему аутентификации и авторизации с JWT.

---

## Технологии

- Go 1.25
- PostgreSQL, 
- Kafka, 
- ClickHouse, 
- Docker, docker-compose
- Prometheus (метрики)
- JSON-логирование, graceful shutdown
- Helm
- Ansible

---

## Архитектура

Проект состоит из трёх основных сервисов и набора общих библиотек:

- **libs/**
    - `logger` — обёртка над стандартным логгером (JSON‑формат, единый стиль логов).
    - `errors` — типизированные ошибки и обёртки.
    - `httpclient` — HTTP‑клиент с таймаутами и логированием.
    - `jwt` — генерация и валидация JWT access‑токенов (поддержка `JWT_SECRET`).
    - `password` — безопасный хеш паролей (bcrypt) + сравнение.
    - `requestid` — генерация и прокидывание request‑id через контекст/HTTP‑заголовки.
    - `metrics` — вспомогательные функции для экспорта метрик Prometheus.

- **services/auth-service/**
    - Отвечает за регистрацию и логин пользователей.
    - Хранит пользователей в Postgres (логин, email, password_hash).
    - Выдаёт JWT access‑токены при валидных кредах.
    - Имеет сервисный слой (`internal/service`) и слой хранения (`internal/storage`).
    - Интеграционные тесты с реальным Postgres через `testcontainers-go`.

- **services/user-service/**
    - CRUD‑операции над пользователями (например, profile/info).
    - Работает с той же Postgres‑БД, что и auth‑service (общая схема `users`).
    - Экспортирует HTTP‑эндпойнты для работы с профилями.

- **services/gateway-service/**
    - Внешний API‑шлюз для клиентов.
    - Проксирует запросы в auth‑service и user‑service.
    - Выполняет аутентификацию по JWT (middleware).
    - Добавляет request‑id, логирование и метрики.

- **deploy/**
    - `docker-compose.yml` — локальный запуск всех сервисов + инфраструктуры.
    - `deploy/prometheus/prometheus.yml` — конфиг Prometheus для сбора метрик.
    - `deploy/helm/go-userauth-microservices` — Helm‑чарт для деплоя в Kubernetes:
        - Deployment/Service для auth, user, gateway.
        - StatefulSet/Service для Postgres.
        - Secret для `JWT_SECRET`.

Взаимодействие:

1. Клиент обращается к `gateway-service` по HTTP (например, `/api/auth/register`, `/api/auth/login`, `/api/users/me`).
2. gateway вызывает auth‑service и user‑service по внутренним HTTP‑адресам.
3. auth‑service работает с Postgres (создание пользователя, проверка логина/пароля, генерация JWT).
4. user‑service читает/обновляет данные пользователя в Postgres.
5. Все HTTP‑сервисы экспортируют метрики Prometheus на `/metrics`.

---

## Запуск без Docker

### Требования

- Go 1.22+ (или актуальная версия Go).
- make (для удобства).
- Локальный Postgres (или переменная `POSTGRES_DSN`).

---

#### Локальный запуск сервисов
```
export POSTGRES_DSN="postgres://auth_user:auth_pass@localhost:5432/auth_db?sslmode=disable"
```

#### Запуск auth‑service:
```
go run ./services/auth-service/cmd/auth
# слушает, например, :8081 (см. internal/config/config.go)
```

#### Запуск user‑service:
```
go run ./services/user-service/cmd/user
# слушает, например, :8082
```

#### Запуск gateway‑service:
```
go run ./services/gateway-service/cmd/gateway
# слушает, например, :8080
```
---

#### Пример HTTP‑запросов
```
curl -s -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","email":"alice@example.com","password":"secret123"}'
```

#### Логин (получение JWT)
```
curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"secret123"}'
```

#### Ответ содержит access_token, который затем используется в Authorization: Bearer <token>.

Доступ к защищённому endpoint’у
```
curl -s http://localhost:8080/api/users/me \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```
---

## Запуск с Docker + Prometheus
```
docker compose up --build
```

#### После старта:

`gateway: http://localhost:8080`

`auth-service: http://localhost:8081`

`user-service: http://localhost:8082`

`prometheus: http://localhost:9090`

#### Проверка метрик
```
# gateway
curl -s http://localhost:8080/metrics | grep http_requests_total

# auth-service
curl -s http://localhost:8081/metrics | grep http_requests_total

# user-service
curl -s http://localhost:8082/metrics | grep http_requests_total
```
---

## Helm‑дeплой в Kubernetes

В проекте есть Helm‑чарт: deploy/helm/go-userauth-microservices.
Он разворачивает:

auth‑service, user‑service, gateway‑service (Deployment + Service);

Postgres (StatefulSet + Service);

Secret с JWT_SECRET.

---

## Makefile

Минимально полезные цели:

```makefile
PROJECT_NAME := go-userauth-microservices
DOCKER_COMPOSE := docker compose

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  make test          - run all Go tests"
	@echo "  make fmt           - run gofmt on all Go files"
	@echo "  make tidy          - run go mod tidy"
	@echo "  make build         - build all service binaries"
	@echo "  make up            - start stack via docker compose"
	@echo "  make down          - stop stack via docker compose"

.PHONY: test
test:
	go test ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: fmt
fmt:
	gofmt -w $$(find . -type f -name '*.go' -not -path "./vendor/*")

.PHONY: build
build:
	mkdir -p bin
	go build -o bin/auth-service ./services/auth-service/cmd/auth
	go build -o bin/user-service ./services/user-service/cmd/user
	go build -o bin/gateway-service ./services/gateway-service/cmd/gateway

.PHONY: up
up:
	$(DOCKER_COMPOSE) up -d

.PHONY: down
down:
	$(DOCKER_COMPOSE) down

```

---

## Структура проекта

```
.
├── libs/                                  # Общие библиотеки для всех сервисов
│   ├── clickhouse/                        # Подключение к ClickHouse (обёртка над клиентом)
│   │   └── clickhouse.go
│   ├── errors/                            # Общие ошибки и обёртки
│   │   └── errors.go
│   ├── httpclient/                        # HTTP-клиент с таймаутами, логированием и request id
│   │   └── httpclient.go
│   ├── jwt/                               # Работа с JWT (access-токены)
│   │   └── jwt.go
│   ├── kafka/                             # Утилиты для работы с Kafka (producer/consumer)
│   │   └── kafka.go
│   ├── logger/                            # Общий логгер (структурированные JSON-логи)
│   │   └── logger.go
│   ├── metrics/                           # Вспомогательные функции для Prometheus-метрик
│   │   └── metrics.go.go
│   ├── password/                          # Хэширование и проверка паролей (bcrypt)
│   │   ├── password.go
│   │   └── password_test.go
│   └── requestid/                         # Генерация и прокидывание request-id
│       ├── requestid.go
│       └── requestid_test.go
└── services/
    ├── auth-service/                      # Сервис аутентификации (регистрация, логин, JWT)
    │   ├── Dockerfile
    │   ├── cmd/auth/                      # Точка входа (main.go)
    │   │   └── main.go
    │   └── internal/
    │       ├── config/                    # Загрузка и валидация конфигурации сервиса
    │       │   └── config.go
    │       ├── http/                      # HTTP-хендлеры и middleware (REST API auth-сервиса)
    │       │   ├── handlers.go
    │       │   └── middleware_requestid.go
    │       ├── service/                   # Бизнес-логика: Register, Login, работа с JWT/паролями
    │       │   ├── service.go
    │       │   ├── service_mock_test.go   # Тесты с моками (testify/mock) поверх storage.Store
    │       │   └── service_test.go        # Unit-тесты сервисного слоя с in-memory/fake Store
    │       └── storage/                   # Слой хранения (интерфейс + Postgres-реализация)
    │           ├── postgres/
    │           │   ├── store.go           # Подключение к Postgres, initSchema, базовые настройки пула
    │           │   ├── users.go           # Реализация CreateUser / GetUserByUsername в Postgres
    │           │   ├── postgres_helper_test.go      # helper для integration-тестов с testcontainers
    │           │   └── postgres_integration_test.go # Интеграционные тесты Postgres.Store (docker + testcontainers)
    │           └── storage.go             # Интерфейс Store, in-memory реализация, модель User
    ├── user-service/                      # Сервис работы с пользователями (профили и т.п.)
    │   ├── Dockerfile
    │   ├── cmd/user/                      # Точка входа (main.go)
    │   │   └── main.go
    │   └── internal/
    │       ├── config/                    # Конфигурация user-сервиса
    │       │   └── config.go
    │       ├── http/                      # HTTP-хендлеры и middleware user-сервиса
    │       │   ├── handlers.go
    │       │   └── middleware_requestid.go
    │       └── storage/                   # Слой хранения пользователей в Postgres
    │           ├── postgres/
    │           │   ├── store.go           # Подключение к Postgres
    │           │   └── users.go           # Операции над пользователями
    │           └── storage.go             # Интерфейс Store и доменная модель
    └── gateway-service/                   # Внешний API-шлюз (объединяет auth и user)
        ├── Dockerfile
        ├── cmd/gateway/                   # Точка входа (main.go)
        │   └── main.go
        └── internal/
            ├── config/                    # Конфигурация gateway (URL внутренних сервисов, порты и т.д.)
            │   └── config.go
            ├── http/                      # HTTP-хендлеры, контекст, middleware, маршрутизация
            │   ├── context.go
            │   ├── handlers.go
            │   ├── middleware.go
            │   └── middleware_requestid.go
            └── middleware/
                └── auth.go                # JWT-auth middleware (проверка токена, прокидывание user в контекст)

```