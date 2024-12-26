include .envrc

MIGRATIONS_PATH := ./cmd/migrate/migrations
DATABASE_CONNSTR := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)

.DEFAULT_GOAL := run

.PHONY: run
run:
	@air

.PHONY: air/build
air/build:
	@echo "Starting up infrastructure (PostgreSQL and Redis) before running the application..." && \
	make infra/start && sleep 5 && \
	echo "Generating Swagger Docs..." && make docs > /dev/null 2>&1 && \
	echo "Building the application..." && go build -o ./tmp/main ./cmd/api && \
	echo "Application built successfully! Running the application with 'air'..."

.PHONY: air/stop
air/stop:
	@echo "Shutting down infrastructure..."; make infra/stop && sleep 2 && echo "Done!"

.PHONY: db/start
db/start:
	@docker compose up -d gophersocial-postgres

.PHONY: db/stop
db/stop:
	@docker compose down gophersocial-postgres

.PHONY: db/migrate/new
db/migrate/new:
	@read -p "Please name your new migration: " name; \
	name=$$(echo "$$name" | tr '[:upper:]' '[:lower:]' | tr ' ' '_'); \
	echo "Using migration name: $$name"; \
	@output=$$(migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))"$$name" 2>&1); \
	up_file=$$(echo "$$output" | grep -o '[^ ]*\.up\.sql'); \
	down_file=$$(echo "$$output" | grep -o '[^ ]*\.down\.sql'); \
	up_filename=$$(echo $$up_file | sed 's|.*/||'); \
	down_filename=$$(echo $$down_file | sed 's|.*/||'); \
	echo "Created the following files:"; \
	echo " - up: $(MIGRATIONS_PATH)/$$up_filename"; \
	echo " - down: $(MIGRATIONS_PATH)/$$down_filename";

# Avoid "No rule to make target" error for arbitrary names passed after the target
%:
	@:

.PHONY: db/migrate/up
db/migrate/up:
	@make db/start
	@echo "Running database 'up'-migrations..."
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DATABASE_CONNSTR) up
	@make db/stop

.PHONY: db/migrate/down
db/migrate/down:
	@make db/start
	@echo "Running database 'down'-migration(s)..."
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DATABASE_CONNSTR) down $(filter-out $@,$(MAKECMDGOALS))
	@make db/stop

.PHONY: db/migrate/force
db/migrate/force:
	@make db/start
	@echo "Running database 'up'-migrations with force..."
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DATABASE_CONNSTR) force $(filter-out $@,$(MAKECMDGOALS))
	@make db/stop

.PHONY: db/seed
db/seed:
	@make db/start
	@echo "Seeding database..."
	@go run ./cmd/migrate/seed/main.go
	@make db/stop

.PHONY: db/reset
db/reset:
	@make db/start
	@echo "Resetting database..."
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DATABASE_CONNSTR) down -all
	@make db/migrate/up
	@make db/seed
	@make db/stop

.PHONY: infra/start
infra/start:
	@docker compose up -d --remove-orphans --quiet-pull 2>/dev/null

.PHONY: infra/stop
infra/stop:
	@docker compose down 2>/dev/null

.PHONY: docs
docs:
	@swag init -d cmd,internal -g api/main.go && swag fmt

.PHONY: test
test:
	@go test -v ./...
