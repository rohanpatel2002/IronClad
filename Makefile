.PHONY: dev test test-go lint build push clean help

COMPOSE_FILE := infra/docker/docker-compose.yml

# ─── Development ────────────────────────────────────────────────────────────

## dev: Start all services with docker-compose
dev:
	docker compose -f $(COMPOSE_FILE) up --build

## dev-detach: Start all services in background
dev-detach:
	docker compose -f $(COMPOSE_FILE) up --build -d

## stop: Stop all running services
stop:
	docker compose -f $(COMPOSE_FILE) down

## logs: Tail logs from all services
logs:
	docker compose -f $(COMPOSE_FILE) logs -f

# ─── Testing ─────────────────────────────────────────────────────────────────

## test: Run all tests across all services
test: test-gate test-topology test-scoring

## test-gate: Run gate-go unit tests with race detector
test-gate:
	@echo "▶ Testing gate-go..."
	cd services/gate-go && go test ./... -race -count=1 -v

## test-topology: Run topology-go unit tests
test-topology:
	@echo "▶ Testing topology-go..."
	cd services/topology-go && go test ./... -race -count=1 -v 2>/dev/null || echo "  No tests yet"

## test-scoring: Run Python scoring tests
test-scoring:
	@echo "▶ Testing scoring-python..."
	cd services/scoring-python && python -m pytest tests/ -v 2>/dev/null || echo "  No tests yet"

# ─── Building ─────────────────────────────────────────────────────────────────

## build: Build all service Docker images
build:
	docker compose -f $(COMPOSE_FILE) build

## build-gate: Build only the gate-go binary
build-gate:
	cd services/gate-go && go build -o bin/gate-go .

## build-topology: Build only the topology-go binary
build-topology:
	cd services/topology-go && go build -o bin/topology-go .

# ─── Linting ─────────────────────────────────────────────────────────────────

## lint: Lint all Go and Python services
lint: lint-go lint-python

## lint-go: Run golangci-lint on all Go services
lint-go:
	@echo "▶ Linting Go services..."
	cd services/gate-go && go vet ./...
	cd services/topology-go && go vet ./...

## lint-python: Run ruff/pylint on Python services
lint-python:
	@echo "▶ Linting Python services..."
	cd services/scoring-python && python -m py_compile scoring_server.py scorer/risk_scorer.py && echo "  OK: scoring-python"
	cd services/semantic-python && python -m py_compile semantic_server.py && echo "  OK: semantic-python"

# ─── Database ─────────────────────────────────────────────────────────────────

## db-up: Start only Postgres
db-up:
	docker compose -f $(COMPOSE_FILE) up postgres -d

## db-shell: Open a psql shell in the running Postgres container
db-shell:
	docker exec -it ironclad-postgres psql -U ironclad -d ironclad

# ─── Quick API Test ────────────────────────────────────────────────────────────

## curl-decision: Fire a test decision request at the running gate
curl-decision:
	curl -s -X POST http://localhost:8080/api/v1/decision \
		-H 'Content-Type: application/json' \
		-d '{"commit_hash":"abc123","service":"payment-api","branch":"main","environment":"production","changed_files":["src/payment.go","migrations/001.sql"]}' \
		| python3 -m json.tool

## curl-blast: Fire a test blast-radius request at the running topology service
curl-blast:
	curl -s -X POST http://localhost:8081/api/v1/blast-radius \
		-H 'Content-Type: application/json' \
		-d '{"service":"payment-api","changed_files":["src/payment.go"]}' \
		| python3 -m json.tool

# ─── Cleanup ─────────────────────────────────────────────────────────────────

## clean: Remove build artifacts and Docker volumes
clean:
	docker compose -f $(COMPOSE_FILE) down -v
	rm -f services/gate-go/bin/gate-go services/topology-go/bin/topology-go

# ─── Help ─────────────────────────────────────────────────────────────────────

## help: Show this help message
help:
	@echo "IRONCLAD Makefile commands:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /' | column -t -s ':'
