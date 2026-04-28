include .env

run:
	go run cmd/api/main.go
build:
	go build ./...
migrate-up:
	migrate -path migrations -database "${DB_URL}" up
migrate-down:
	migrate -path migrations -database "${DB_URL}" down
db-start:
	docker compose up -d
db-stop:
	docker compose down
sqlc-gen:
	sqlc generate