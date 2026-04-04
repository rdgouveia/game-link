include .env
export

NAME ?= ''

migrate-up:
	goose -dir ./sql/migrations postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir ./sql/migrations postgres "$(DATABASE_URL)" down

migrate-status:
	goose -dir ./sql/migrations postgres "$(DATABASE_URL)" status

migrate-create:
	goose -dir ./sql/migrations create $(NAME) sql

generate:
	sqlc generate