GO := ./scripts/go
.PHONY: db-up db-native migrate seed-demo run worker worker-env frontend-install frontend-dev frontend-build frontend-check frontend-test test test-integration fmt vet backup

db-up:
	docker compose up -d db

db-native:
	./scripts/db-native.sh

migrate:
	$(GO) run ./cmd/migrate

seed-demo:
	$(GO) run ./cmd/seed

run:
	$(GO) run ./cmd/server

worker-env:
	python3 scripts/worker-env.py

worker: worker-env
	$(GO) run ./cmd/worker

frontend-install:
	npm ci --prefix frontend --cache /private/tmp/iml-npm-cache

frontend-dev:
	npm run dev --prefix frontend

frontend-build:
	npm run build --prefix frontend

frontend-check:
	npm run typecheck --prefix frontend
	npm run lint --prefix frontend

frontend-test:
	npm test --prefix frontend

test:
	$(GO) test ./cmd/... ./internal/...

test-integration:
	python3 scripts/test-integration.py

fmt:
	$(GO) fmt ./cmd/... ./internal/...

vet:
	$(GO) vet ./cmd/... ./internal/...

backup:
	python3 scripts/backup.py
