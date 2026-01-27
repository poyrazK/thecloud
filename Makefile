.PHONY: run test test-coverage doccheck build migrate clean stop swagger reset clean-docker

# Configuration
PROJECT_PREFIX = cloud-
COMPOSE_FILES = -f docker-compose.yml
TEST_TIMEOUT ?= 2m

run: stop
	docker compose $(COMPOSE_FILES) up -d --build
	@echo "Services are running. API at http://localhost:8080"

dev: build
	docker compose $(COMPOSE_FILES) up -d --build
	docker compose $(COMPOSE_FILES) logs -f api

stop:
	@echo "Stopping services..."
	docker compose $(COMPOSE_FILES) stop
	@fuser -k 8080/tcp 2>/dev/null || true

# Nuclear cleanup for stale resources
clean-docker:
	@echo "Cleaning up stale Docker resources..."
	@docker compose $(COMPOSE_FILES) down --remove-orphans 2>/dev/null || true
	@# Remove any remaining containers with our prefix
	@docker ps -a --filter "name=$(PROJECT_PREFIX)" -q | xargs -r docker rm -f
	@# Remove project network specifically if it's stuck
	@docker network rm $(PROJECT_PREFIX)network 2>/dev/null || true
	@echo "Pruning stale networks..."
	@docker network prune -f

reset: clean-docker
	@echo "Performing fresh start..."
	rm -rf bin
	docker compose $(COMPOSE_FILES) up -d --build
	@echo "System reset complete."

test:
	go test ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@rm coverage.out

test-e2e:
	go test -timeout $(TEST_TIMEOUT) ./tests/...

doccheck:
	go run ./cmd/doccheck --root .

swagger:
	@$(HOME)/go/bin/swag init -d cmd/api,internal/handlers -g main.go -o docs/swagger --parseDependency --parseInternal

build:
	mkdir -p bin
	go build -o bin/api ./cmd/api
	go build -o bin/cloud cmd/cloud/*.go
	go build -o bin/storage-node cmd/storage-node/main.go

install: build
	mkdir -p $(HOME)/.local/bin
	cp bin/cloud $(HOME)/.local/bin/cloud
	@./scripts/setup_path.sh

migrate:
	@echo "Running migrations..."
	@docker compose up -d postgres
	@sleep 2
	@go run cmd/api/main.go --migrate-only

clean: clean-docker
	rm -rf bin
