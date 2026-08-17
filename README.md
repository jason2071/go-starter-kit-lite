# Go Fiber Clean Architecture Lite

Compact starter using **Go + Fiber v2 + GORM + PostgreSQL**.

This refactored edition keeps the original API behavior but uses explicit file and function names so the codebase is easier to follow and extend.

## Included

- Fiber v2
- GORM + PostgreSQL
- golang-migrate SQL migrations
- validator
- JWT access tokens
- rotating refresh tokens stored as SHA-256 hashes
- RBAC (`user`, `admin`)
- `slog` structured logging
- request ID
- pagination / search / filter / whitelist sort
- graceful shutdown
- Docker / Docker Compose
- Air
- unit + integration test examples
- OpenAPI + Swagger UI

## Structure

```text
cmd/api/
├── main.go                 # composition root and dependency wiring
└── routes.go               # HTTP endpoint registration

docs/
├── openapi.yaml            # OpenAPI specification
└── swagger-ui.html         # Swagger UI page

internal/
├── config/
│   └── config.go           # Fiber, CORS and request-ID configuration
├── domain/
│   ├── user.go
│   └── refresh_token.go
├── usecase/
│   ├── auth_service.go
│   ├── auth_types.go
│   ├── user_service.go
│   ├── user_types.go
│   └── error.go
├── repository/
│   ├── user_repository.go
│   └── refresh_token_repository.go
├── handler/http/
│   ├── app.go
│   ├── auth_handler.go
│   ├── dependencies.go
│   ├── response.go
│   ├── user_handler.go
│   └── system_handler.go
├── middleware/
│   ├── auth.go
│   └── request_logger.go
├── platform/
│   ├── config.go
│   ├── database.go
│   ├── logger.go
│   └── security.go
└── shared/
    └── pagination.go
```

## Quick start

```bash
cp .env.example .env
docker compose up --build
```

- API: http://localhost:8080
- Swagger: http://localhost:8080/docs
- Health: http://localhost:8080/healthz
- Ready: http://localhost:8080/readyz

## Local development

```bash
go mod tidy
make migrate-up
make dev
```

## Endpoints

```text
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/refresh
POST /api/v1/auth/logout
GET  /api/v1/auth/me
GET  /api/v1/users
GET  /healthz
GET  /readyz
GET  /docs
GET  /openapi.yaml
```

## Adding a new CRUD feature

For `Category`, follow the same naming pattern:

```text
internal/domain/category.go
internal/usecase/category_service.go
internal/usecase/category_types.go
internal/repository/category_repository.go
internal/handler/http/category_handler.go
```

Then wire the repository/service in `cmd/api/main.go` and register routes in `cmd/api/routes.go`.
