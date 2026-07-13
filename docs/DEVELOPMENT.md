# Development

## Requirements

- Go 1.25+
- Node.js 20.9+ (Node.js 24 recommended)
- PostgreSQL 16+
- Redis 7+
- Docker with Compose for the one-command stack

## Local Services

Start PostgreSQL and Redis, then configure:

```bash
cp .env.example .env
```

Run migrations with `migrate/migrate` or `make migrate`, then:

```bash
cd apps/api
go run ./cmd/api
```

In another terminal:

```bash
npm install
npm run dev
```

The Next.js BFF expects `API_URL=http://localhost:8080/api/v1`. Compose sets it to the internal `api` service address automatically.

## Tests

```bash
cd apps/api
go test ./...

cd ../..
npm test
npm run lint
npm run build
```

Application tests use in-memory fakes to verify business policy without a database. HTTP tests cover authentication middleware. Configuration and security packages have focused unit tests. Add PostgreSQL integration tests when introducing complex queries or transactions.

## Migrations

Create paired files in `apps/api/migrations`:

```text
000002_feature.up.sql
000002_feature.down.sql
```

Never edit an applied migration in a shared environment. Add a new migration instead.

## Configuration

All runtime configuration is loaded from environment variables. Production validates a minimum JWT secret length. `.env` is ignored by Git; `.env.example` documents supported settings.

## API Changes

Update these together:

1. Domain/application contract
2. HTTP handler and tests
3. `apps/api/openapi/openapi.yaml`
4. Frontend types and callers
