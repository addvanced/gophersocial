include .envrc

MIGRATIONS_PATH := ./cmd/migrate/migrations
DATABASE_CONNSTR := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)

.PHONY: db/start
db/start:
	@docker compose up -d

.PHONY: db/stop
db/stop:
	@docker compose down

.PHONY: db/migrate/create
db/migrate/create:
	@output=$$(migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS)) 2>&1); \
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

.PHONY: db/migrate/down
db/migrate/down:
	@make db/start
	@echo "Running database 'down'-migration(s)..."
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DATABASE_CONNSTR) down $(filter-out $@,$(MAKECMDGOALS))

.PHONY: db/migrate/force
db/migrate/force:
	@make db/start
	@echo "Running database 'up'-migrations with force..."
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DATABASE_CONNSTR) force $(filter-out $@,$(MAKECMDGOALS))

.PHONY: db/seed
db/seed:
	@make db/start
	@echo "Seeding database..."
	@go run ./cmd/migrate/seed/main.go

.PHONY: db/reset
db/reset:
	@make db/start
	@echo "Resetting database..."
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DATABASE_CONNSTR) down -all
	@make db/migrate/up
	@make db/seed

.PHONY: docs
docs:
	@swag init -d cmd,internal -g api/main.go && swag fmt