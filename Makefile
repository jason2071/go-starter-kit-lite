APP=go-starter-kit-lite
DB_URL?=postgres://postgres:postgres@localhost:5432/app?sslmode=disable
MIGRATE?=migrate

.PHONY: run dev deps fmt vet test test-integration migrate-up migrate-down docker-up docker-down
run:
	go run ./cmd/api

dev:
	air

deps:
	go mod tidy

fmt:
	gofmt -w ./cmd ./internal ./integration

vet:
	go vet ./...

test:
	go test ./...

test-integration:
	TEST_DATABASE_URL=$(DB_URL) go test -tags=integration ./integration/...

migrate-up:
	$(MIGRATE) -path migrations -database "$(DB_URL)" up

migrate-down:
	$(MIGRATE) -path migrations -database "$(DB_URL)" down 1

docker-up:
	docker compose up --build

docker-down:
	docker compose down
