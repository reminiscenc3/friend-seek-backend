CMD      := ./cmd/api
MIGRATE  := ./cmd/migrate


ifneq (,$(wildcard .env))
include .env
export
endif


.PHONY: run
run: ## Run the service
	go run $(CMD)

.PHONY: migrate
migrate:
	go run $(MIGRATE)

.PHONY: docker-build
docker-build: ## Build the image
	docker compose build

.PHONY: up
up: ## Migrate and start the service in the background
	docker compose up -d --build

.PHONY: down
down: ## Stop the service
	docker compose down

.PHONY: clean
clean: down ## Stop the service and delete ./data with the database
	rm -rf ./data

.PHONY: logs
logs: ## Follow the service logs
	docker compose logs -f api
