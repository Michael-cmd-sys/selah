.PHONY: up down dev build test clean

ENV_FILE := .env

# Copy .env.example if .env doesn't exist
$(ENV_FILE):
	cp .env.example $(ENV_FILE)
	@echo "Created .env from .env.example — set JWT_SECRET before running."

## up: start Postgres in the background
up:
	docker-compose up -d
	@echo "Waiting for Postgres to be ready..."
	@until docker-compose exec -T postgres pg_isready -U selah -d selah > /dev/null 2>&1; do sleep 1; done
	@echo "Postgres is ready."

## down: stop and remove containers
down:
	docker-compose down

## dev: start DB + run the server (builds first)
dev: $(ENV_FILE) up build
	@fuser -k 8080/tcp 2>/dev/null || true
	./bin/selah

## build: compile the server binary
build:
	go build -o bin/selah ./cmd/server

## test: run all tests
test:
	go test ./... -count=1

## clean: remove binary and stop containers
clean: down
	rm -f ./bin/selah
