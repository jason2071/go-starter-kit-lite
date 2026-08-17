# Clean Architecture Lite

Dependency direction:

```text
Handler (Fiber)       Repository (GORM)      Platform
       \                    |                  /
        \                   v                 /
                  Usecase / Ports
                        |
                        v
                      Domain
```

Rules:
- `domain` contains pure entities only; no Fiber/GORM tags.
- `usecase` contains business logic and interfaces (ports).
- `repository` implements persistence ports using GORM/PostgreSQL.
- `handler/http` owns HTTP parsing, handlers, routes, middleware and status codes.
- `platform` owns config, DB setup, logger, JWT/bcrypt/refresh-token primitives.
- `cmd/api/main.go` is the composition root (dependency wiring), so no separate bootstrap package is needed.

Lite means fewer packages/files, not weaker dependency boundaries.
