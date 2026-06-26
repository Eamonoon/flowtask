.PHONY: dev dev-server dev-web test build clean

dev: ## Start all services (requires docker-compose up -d postgres redis first)
	@echo "Starting all services..."
	@make -j2 dev-server dev-web

dev-server: ## Start Go backend
	@echo "Starting Go backend..."
	cd server && go run cmd/server/main.go

dev-web: ## Start Next.js frontend
	@echo "Starting Next.js frontend..."
	cd web && npm run dev

test: ## Run all tests
	cd server && go test ./... -v
	cd web && npm run test

test-server: ## Run backend tests
	cd server && go test ./... -v

test-web: ## Run frontend tests
	cd web && npm run test

build: build-server build-web ## Build all

build-server: ## Build Go backend
	cd server && go build -o ../bin/server cmd/server/main.go

build-web: ## Build Next.js frontend
	cd web && npm run build

docker-up: ## Start PostgreSQL and Redis
	docker-compose up -d postgres redis

docker-down: ## Stop all containers
	docker-compose down

clean: ## Clean build artifacts
	rm -rf bin/
	cd web && rm -rf .next out/

lint: lint-server lint-web ## Lint all

lint-server: ## Lint Go code
	cd server && golangci-lint run

lint-web: ## Lint frontend code
	cd web && npm run lint

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
