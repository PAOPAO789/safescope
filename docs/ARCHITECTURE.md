# Architecture

## Monorepo

```text
apps/
  api/
    cmd/api/                    composition root
    internal/domain/            entities, errors, repository and provider ports
    internal/application/       authentication, RBAC, project, asset and dashboard use cases
    internal/infrastructure/    PostgreSQL and Redis adapters
    internal/interfaces/http/   Gin routes, middleware and HTTP DTOs
    internal/platform/          configuration, logging and JWT/password services
    migrations/                 versioned PostgreSQL schema
    openapi/                    OpenAPI contract
  web/
    app/                        Next.js routes and BFF route handlers
    components/                 workspace feature components
    lib/                        API client, server proxy and shared types
    test/                       frontend unit tests
```

Dependencies point inward: the domain has no framework imports; application services depend on domain ports; infrastructure and HTTP layers implement or consume those ports; `cmd/api` wires concrete adapters.

## Authentication

The API issues signed HS256 JWTs. The web BFF stores the token in an `HttpOnly`, `SameSite=Lax` cookie and forwards it to the API as a Bearer token. The browser UI never reads the JWT directly.

For TLS deployments, set `COOKIE_SECURE=true`. Use a secret manager for `JWT_SECRET` and rotate it through a controlled session invalidation process.

## Authorization

| Capability | Admin | Analyst | Viewer |
| --- | --- | --- | --- |
| View all projects | Yes | Owned | Owned |
| Create/edit projects | Yes | Owned | No |
| Create/edit assets | Yes | Owned | No |
| Manage user roles | Yes | No | No |

Project ownership is enforced in the application service and repository access check. The UI hides unavailable actions, but the API remains the authority.

## Data

Migrations are the schema source of truth. Projects own assets with cascading deletion. Asset values are unique by project, type, and value. JSONB metadata and array tags allow scanner-specific observations without changing the core entity for every adapter.

## Operations

The API emits JSON logs in production with request ID, method, path, status, and latency. `/healthz` probes PostgreSQL and Redis. Services use Compose health conditions, and the API shuts down gracefully on `SIGTERM`.
