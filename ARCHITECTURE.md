# Clean Architecture Lite

Dependency direction:

```text
cmd/api
  ├── app.go         builds the Fiber app and HTTP middleware
  ├── routes.go      creates HTTP dependencies and registers endpoints
  └── main.go        starts and stops the application

Handler (Fiber)  --->  Usecase  --->  Domain
                          ^
                          |
             Repository (GORM implementation)

Middleware (HTTP) ---> Usecase token port
```

## Packages

- `internal/domain` — pure business entities, repository interfaces, and their contract errors. No Fiber/GORM tags.
- `internal/usecase` — application services, request/response types, plus non-repository ports such as `TokenManager` and `PasswordHasher`.
- `internal/repository` — PostgreSQL/GORM implementations. It implements interfaces from `domain`.
- `internal/models` — GORM persistence models and table mappings.
- `internal/handler` — request handlers, validation and HTTP responses.
- `internal/middleware` — reusable HTTP authentication, authorization and request logging.
- `internal/config` — Fiber, CORS and request-ID configuration.
- `internal/platform` — config, database, logger and JWT/password implementation.
- `internal/shared` — small reusable primitives such as pagination.
- `cmd/api/main.go` — application lifecycle: config, database connection and shutdown.
- `cmd/api/app.go` — Fiber app and HTTP middleware setup.
- `cmd/api/routes.go` — HTTP composition root: repository/use case/handler construction and endpoint registration.
- `docs/openapi.yaml` and `docs/swagger-ui.html` — API specification and Swagger UI.

## Repository boundary

The domain owns repository contracts because they describe business data:

```text
domain/UserRepository           interface (what the app needs)
repository/UserRepository       GORM/Postgres implementation (how it is done)
models/UserModel                 GORM table mapping
```

This lets use cases depend only on `domain.UserRepository`; neither `domain`
nor `usecase` imports `repository`. GORM tags remain in `internal/models`;
queries remain in `internal/repository`.

## Naming rule

Names intentionally include the resource when ambiguity is possible:

```text
FindUserByID
ListUsers
CreateRefreshToken
RevokeRefreshToken
RegisterUser
GetCurrentUser
```

This makes a new feature such as Category predictable:

```text
domain/category.go                  entity and repository contract
usecase/category_service.go
usecase/category_types.go
repository/category_repository.go
models/category.go
handler/category.go
cmd/api/app.go
cmd/api/routes.go
```
