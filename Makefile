.PHONY: start start-db gen-docs help

help:
	@echo "Available commands:"
	@echo "---------------------"
	@echo "start: Start the application server"
	@echo "start-db: Start the containerized database instance"
	@echo "gen-docs: Generate the OpenAPI swagger documentation"

start:
# 	go run cmd/main.go
	docker compose up

dev:
	docker compose up --watch

start-db:
	docker compose up -d

gen-docs:
	swag init -g ./cmd/api/main.go -o ./docs
