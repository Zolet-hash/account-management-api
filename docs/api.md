# API

Base path `/api/v1`. JSON errors use `{ "error": { "code": "...", "message": "..." } }`.

- `POST /users` — `{ "username": "...", "dateOfBirth": "YYYY-MM-DD", "documentNumber": "..." }`; returns 201.
- `POST /accounts` — `{ "userId": "UUID", "exchangeId": "UUID" }`; returns 201.
- `POST /accounts/{accountId}/deposit` and `/withdrawal` — `{ "amount": 12.34 }`; set `Idempotency-Key` header; returns transaction, 200.
- `GET /users/{userId}/balances` — per-exchange positive balances and total.
- `GET /transactions/volume?startDate=YYYY-MM-DD&endDate=YYYY-MM-DD` — inclusive UTC dates, empty days included.
- `GET /health` — process liveness. `GET /ready` — database readiness. `/metrics` exposes Prometheus metrics.

Exchange IDs are seeded via `scripts/seed.sql`; the seed records are Atlas `11111111-1111-4111-8111-111111111111` and Aurora `22222222-2222-4222-8222-222222222222`.

OpenAPI 3.1 endpoint definitions are in `openapi.yaml`.
