include .env

migrate-up:
	migrate -path migrations -database "${DB_URL}" up
migrate-down:
	migrate -path migrations -database "${DB_URL}" down
db-start:
	docker compose up -d
db-stop:
	docker compose down