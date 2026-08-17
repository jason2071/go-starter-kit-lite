# Go Fiber Clean Architecture Lite

A compact production-ready starter using **Go + Fiber v2 + GORM + PostgreSQL** with standard Clean Architecture boundaries, but less boilerplate than the full/classic version.

## Included

- Fiber v2
- GORM + PostgreSQL
- golang-migrate compatible SQL migrations
- validator
- JWT access tokens
- rotating refresh tokens stored as SHA-256 hashes
- RBAC (`user`, `admin`)
- `slog` JSON structured logging
- request ID
- pagination / search / filter / whitelist sort
- graceful shutdown
- Docker / Docker Compose
- Air
- Makefile
- unit + integration test examples
- OpenAPI + Swagger UI

## Structure

```text
cmd/api/main.go                # composition root / DI
internal/domain/               # pure entities
internal/usecase/              # business logic + ports
internal/repository/           # GORM/Postgres implementation
internal/handler/http/         # Fiber handlers, routes, middleware
internal/platform/             # config/db/logger/security
internal/shared/               # small reusable primitives
migrations/
docs/
integration/
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
make deps
make migrate-up
make dev
```

Install tools if needed:

```bash
go install github.com/air-verse/air@latest
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

## Endpoints

```text
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/refresh
POST /api/v1/auth/logout
GET  /api/v1/auth/me
GET  /api/v1/users        # admin
GET  /healthz
GET  /readyz
GET  /docs
```

## Promote a user to admin

After registering a user, run:

```sql
INSERT INTO user_roles (user_id, role_id)
SELECT 'USER_UUID', id FROM roles WHERE name = 'admin'
ON CONFLICT DO NOTHING;
```

## Why Lite?

The HTTP package is the handler layer: it owns routes, request parsing, validation, middleware, and HTTP responses. The use-case layer stays independent of Fiber, while repositories and platform packages remain replaceable infrastructure adapters.
