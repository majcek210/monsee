.PHONY: up down build logs shell-backend shell-db shell-redis \
        migrate migrate-new sqlc lint test test-int create-admin metrics asynqmon

# ── Docker Compose ─────────────────────────────────────────────────────────────
up:
	docker compose up -d

down:
	docker compose down

build:
	docker compose up --build -d

logs:
	docker compose logs -f $(svc)

# ── Shells ─────────────────────────────────────────────────────────────────────
shell-backend:
	docker compose exec backend sh

shell-db:
	docker compose exec postgres psql -U statususer -d statusdb

shell-redis:
	docker compose exec redis redis-cli

# ── Database / sqlc ────────────────────────────────────────────────────────────
migrate:
	docker compose exec backend /migrate

migrate-new:
	@read -p "Migration name: " name; \
	seq=$$(ls backend/db/migrations/*.up.sql 2>/dev/null | wc -l | tr -d ' '); \
	seq=$$(printf "%06d" $$((seq + 1))); \
	touch backend/db/migrations/$${seq}_$${name}.up.sql; \
	touch backend/db/migrations/$${seq}_$${name}.down.sql; \
	echo "Created $${seq}_$${name}.up.sql and $${seq}_$${name}.down.sql"

sqlc:
	cd backend && sqlc generate

# ── Code quality ───────────────────────────────────────────────────────────────
lint:
	cd backend && golangci-lint run ./...

test:
	cd backend && go test ./... -race -cover

test-int:
	cd backend && go test ./... -race -tags integration

# ── Helpers ────────────────────────────────────────────────────────────────────
create-admin:
	docker compose exec backend /create-admin

metrics:
	open http://localhost:9112/metrics || xdg-open http://localhost:9112/metrics

asynqmon:
	open http://localhost:9001 || xdg-open http://localhost:9001
