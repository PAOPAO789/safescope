# SafeScope

SafeScope is a modern AI Security Workspace for organizing security projects, managing internet-facing assets, and preparing scanner and AI-assisted analysis workflows.

## Included

- Next.js 16 workspace UI with a server-side BFF and `HttpOnly` JWT session cookie
- Go 1.25 API using Gin and Clean Architecture boundaries
- PostgreSQL migrations and repository implementations
- Redis connectivity for health and future job orchestration
- JWT authentication and `admin` / `analyst` / `viewer` RBAC
- Project CRUD, asset CRUD, unified dashboard, and team role management
- Structured Zap logs, request IDs, CORS, graceful shutdown, and health checks
- OpenAPI 3.1 contract
- Docker Compose deployment and focused unit tests

## Quick Start

1. Copy `.env.example` to `.env`.
2. Replace `POSTGRES_PASSWORD`, update the password inside `DATABASE_URL`, and replace `JWT_SECRET`.
3. Run:

```bash
docker compose up --build
```

Open the workspace at `http://localhost:3000`. The API is available at `http://localhost:8080`, and its OpenAPI document is at `http://localhost:8080/openapi.yaml`.

The first registered user receives the `admin` role. Later users receive `analyst` by default.

## Commands

```bash
make dev       # build and run the full stack
make test      # run Go and web tests
make build     # build API and web
make migrate   # run pending database migrations
make down      # stop the stack
```

See [Development](docs/DEVELOPMENT.md), [Architecture](docs/ARCHITECTURE.md), and [Extensions](docs/EXTENSIONS.md) for implementation details.
