.PHONY: help dev up down build test migrate backend-test frontend-dev backend-dev clean

help:
	@echo "AuctionXI — Available commands:"
	@echo "  make dev          Start all services via Docker Compose"
	@echo "  make up           Alias for dev"
	@echo "  make down         Stop all services"
	@echo "  make build        Build all Docker images"
	@echo "  make backend-test Run Go unit tests"
	@echo "  make backend-dev  Run Go API locally"
	@echo "  make frontend-dev Run Next.js dev server"
	@echo "  make migrate      Run database migrations"
	@echo "  make clean        Remove build artifacts"

dev up:
	docker compose up --build

down:
	docker compose down

build:
	docker compose build

backend-test:
	cd backend && go test ./...

backend-dev:
	cd backend && go run ./cmd/api

frontend-dev:
	cd frontend && npm run dev

migrate:
	cd backend && goose -dir migrations postgres "$${DATABASE_URL:-postgres://auctionxi:auctionxi@localhost:5432/auctionxi?sslmode=disable}" up

clean:
	rm -rf backend/bin frontend/.next frontend/node_modules
