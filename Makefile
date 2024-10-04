include .envrc
MIGRATIONS_PATH := ./cmd/migrate/migrations
DB_DSN := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)

.PHONY: migration
migration:
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

.PHONY: migrate/up
migrate/up:
	@echo "Running database 'up'-migrations..."
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_DSN) up

.PHONY: migrate/down
migrate/down:
	@echo "Running database 'down'-migration(s)..."
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_DSN) down $(filter-out $@,$(MAKECMDGOALS))