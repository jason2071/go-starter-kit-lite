# Clean Architecture Lite

Dependency direction:

```text
cmd/api
  ├── routes.go      registers handlers and middleware
  └── main.go        wires concrete dependencies

Handler (Fiber)  --->  Usecase  --->  Domain
                          ^
                          |
                   Repository (GORM)

Middleware (HTTP) ---> Usecase token port
```

## Packages

- `internal/domain` — pure business entities. No Fiber/GORM tags.
- `internal/usecase` — application services, request/response types, and the ports they use.
- `internal/repository` — PostgreSQL/GORM implementations of repository interfaces.
- `internal/handler/http` — Fiber app setup, request handlers, validation and HTTP responses.
- `internal/middleware` — reusable HTTP authentication, authorization and request logging.
- `internal/config` — Fiber, CORS and request-ID configuration.
- `internal/platform` — config, database, logger and JWT/password implementation.
- `internal/shared` — small reusable primitives such as pagination.
- `cmd/api/main.go` — composition root / dependency injection.
- `cmd/api/routes.go` — endpoint registration.
- `docs/openapi.yaml` and `docs/swagger-ui.html` — API specification and Swagger UI.

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
domain/category.go
usecase/category_service.go
usecase/category_types.go
repository/category_repository.go
handler/http/category_handler.go
cmd/api/routes.go
```
