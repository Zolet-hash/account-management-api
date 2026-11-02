# Account Management API

A Go 1.24 modular monolith for managing users, exchange accounts, deposits, withdrawals, aggregate balances and daily exchange transaction volume. The service uses Gin, PostgreSQL/pgx, `NUMERIC(18,2)` balances, structured `slog` logs, OpenTelemetry traces and Prometheus metrics.

## Layout

- `cmd/api`: API startup and dependency wiring; `cmd/seed`: repeatable exchange seeder; `cmd/migrate`: migration CLI wrapper.
- `internal/domain`, `model`, `service`, `handler`, `repository/postgres`: business types/rules and transport/persistence boundaries.
- `migrations`: reversible SQL migrations. `scripts/seed.sql`: idempotent exchange catalog.
- `deployments`: Docker, Compose, collector, Prometheus and Grafana setup.
- `tests/load`: k6 flow and concurrency scenarios; `openapi.yaml`: OpenAPI contract.
- `docs`: architecture, decisions, API and testing notes.

## Local setup

Prerequisites: Go 1.24+, PostgreSQL 16+, `golang-migrate` CLI. Copy `.env.example` to `.env`, then export its values in your shell. Create the `accounts` database, download the pinned modules, and run:

```sh
go mod download
make migrate
make seed
make run
```

The API listens on `:8080` by default. For a clean container setup run `docker compose --env-file .env -f deployments/docker-compose.yml up --build`; Compose starts PostgreSQL, runs migrations and seeds the exchanges before starting the API, Prometheus, Grafana and OTEL Collector. Local Grafana is at `http://localhost:3000`, Prometheus at `http://localhost:9090`, API health at `http://localhost:8080/health`, and readiness at `/ready`.

Database credentials are supplied through `POSTGRES_PASSWORD`; the Compose default is only suitable for local development. Configure a secret in production. API startup does not apply schema changes; migration is an explicit deployment step.

## Configuration

`DATABASE_URL`, `LISTEN_ADDR`, `OTEL_EXPORTER_OTLP_ENDPOINT`, and `SHUTDOWN_TIMEOUT_SECONDS` are supported. PostgreSQL pool sizing/timeouts are configured in `internal/infrastructure/database`. Tracing export can be disabled by leaving the OTLP endpoint empty. `/metrics` exposes Prometheus counters and histograms. Logs and traces intentionally omit document numbers and other personal fields.

## API

See [docs/api.md](docs/api.md) and [openapi.yaml](openapi.yaml). User and account writes return 201; successful transfers return 200 and require an `Idempotency-Key` header. Exchange creation is not exposed: seed rows are the catalog source. Errors have a stable `{error:{code,message}}` envelope.

## Tests and quality

```sh
go test ./...
go test -race ./...
go vet ./...
gofmt -w cmd internal tests
```

Repository integration coverage is intended to run against testcontainers when Docker is available; see [docs/testing.md](docs/testing.md). Run load scenarios with `BASE_URL=http://localhost:8080 k6 run tests/load/transactions.js`. The script exercises 50, 100 and 500 concurrent VUs, including writes, reads and reporting. k6 outputs actual throughput, average/p95/p99 latency and error rate; inspect `db` connection utilization alongside those values. No performance measurements are claimed until a run is performed.

## Design notes

Transfers lock the account row inside a SQL transaction and update balance with the immutable transaction insert. Idempotency scope is account plus key. See [docs/architecture.md](docs/architecture.md) and [docs/decisions.md](docs/decisions.md) for flows, precision, date and age assumptions.
