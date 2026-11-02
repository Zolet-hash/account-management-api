.PHONY: run migrate seed test integration race vet fmt compose-up compose-down load
run:
	go run ./cmd/api
migrate:
	migrate -path migrations -database "$(DATABASE_URL)" up
seed:
	go run ./cmd/seed
test:
	go test ./...
integration:
	go test -tags=integration ./tests/integration
race:
	go test -race ./...
vet:
	go vet ./...
fmt:
	gofmt -w cmd internal tests
compose-up:
	docker compose -f deployments/docker-compose.yml up --build
compose-down:
	docker compose -f deployments/docker-compose.yml down
load:
	k6 run tests/load/transactions.js
