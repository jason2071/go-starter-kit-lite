# Clean Architecture Lite

Dependency direction:

```text
cmd/api
  ├── routes.go      registers handlers and middleware
  └── main.go        wires concrete dependencies

Handler (Fiber)  --->  Usecase  --->  Domain
                          ^
                          |
             Repository (GORM implementation)

Middleware (HTTP) ---> Usecase token port
```

## Packages

- `internal/domain` — pure business entities, repository interfaces, and their contract errors. No Fiber/GORM tags.
- `internal/usecase` — application services, request/response types, plus non-repository ports such as `TokenManager` and `PasswordHasher`.
- `internal/repository` — PostgreSQL/GORM implementations and GORM models. It implements interfaces from `domain`.
- `internal/handler/http` — Fiber app setup, request handlers, validation and HTTP responses.
- `internal/middleware` — reusable HTTP authentication, authorization and request logging.
- `internal/config` — Fiber, CORS and request-ID configuration.
- `internal/platform` — config, database, logger and JWT/password implementation.
- `internal/shared` — small reusable primitives such as pagination.
- `cmd/api/main.go` — composition root / dependency injection.
- `cmd/api/routes.go` — endpoint registration.
- `docs/openapi.yaml` and `docs/swagger-ui.html` — API specification and Swagger UI.

## Repository boundary

The domain owns repository contracts because they describe business data:

```text
domain/UserRepository           interface (what the app needs)
repository/UserRepository       GORM/Postgres implementation (how it is done)
repository/UserModel             GORM table mapping
```

This lets use cases depend only on `domain.UserRepository`; neither `domain`
nor `usecase` imports `repository`. GORM tags and queries remain in
`internal/repository`.

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
domain/category_repository.go
usecase/category_service.go
usecase/category_types.go
repository/category_repository.go
repository/category_model.go
handler/http/category_handler.go
cmd/api/routes.go
```
