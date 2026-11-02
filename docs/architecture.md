# Architecture

The API is a Go modular monolith. Gin handlers parse and validate transport input, the service applies business rules, and the PostgreSQL repository owns queries and transactional consistency. The command entry point loads environment configuration, opens the pool, configures telemetry and wires dependencies.

A transfer validates its account and exchange limit, then the repository begins a PostgreSQL transaction, locks the account row, checks balance/idempotency, updates the account and inserts an immutable transaction before commit. The row lock serializes concurrent transfers for an account. Readiness performs a bounded database ping; liveness reports process health independently.

Metrics are served at `/metrics`; OTLP HTTP traces are optional through `OTEL_EXPORTER_OTLP_ENDPOINT`. Request IDs are returned in `X-Request-ID`. Prometheus and Grafana are provided in Compose.
