.PHONY: dev test build up down logs migrate

dev:
	docker compose up --build

up:
	docker compose up --build -d

down:
	docker compose down

logs:
	docker compose logs -f

test:
	cd apps/api && go test ./...
	npm test

build:
	cd apps/api && go build ./cmd/api
	npm run build

migrate:
	docker compose run --rm migrate
