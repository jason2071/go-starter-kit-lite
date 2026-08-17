# Clean Architecture Lite

Dependency direction:

```text
Handler (Fiber)  --->  Usecase / Ports  --->  Domain
                           ^
                           |
                    Repository (GORM)

Platform provides config, database, logging and security implementations.
cmd/api/main.go wires concrete implementations together.
```

## Packages

- `internal/domain` — pure business entities. No Fiber/GORM tags.
- `internal/usecase` — application services and repository/security interfaces.
- `internal/repository` — PostgreSQL/GORM implementations of repository interfaces.
- `internal/handler/http` — HTTP handlers, routes, middleware, validation and responses.
- `internal/platform` — config, database, logger and JWT/password implementation.
- `internal/shared` — small reusable primitives such as pagination.
- `cmd/api/main.go` — composition root / dependency injection.

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
usecase/category_repository.go
repository/category_repository.go
handler/http/category_handler.go
```
