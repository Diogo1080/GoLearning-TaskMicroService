.PHONY: build up down logs clean rebuild migrate-up

# Apply all pending database migrations using the configured environment.
migrate-up:
	go run ./cmd/migrate

# Build all services
build:
	docker-compose build

# Start services
up:
	docker-compose up -d

# Start with logs
logs:
	docker-compose up

# Stop services
down:
	docker-compose down

# View logs
view-logs:
	docker-compose logs -f

# View auth logs
logs-auth:
	docker-compose logs -f auth-service

# View web logs
logs-web:
	docker-compose logs -f web-service

# Clean everything (including volumes)
clean:
	docker-compose down -v

# Rebuild and restart
rebuild:
	docker-compose down && docker-compose build && docker-compose up -d

# Shell into web service
shell-web:
	docker-compose exec web-service sh

# Shell into auth service
shell-auth:
	docker-compose exec auth-service sh