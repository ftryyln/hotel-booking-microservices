.PHONY: run down test lint swagger

run:
	docker-compose up --build

down:
	docker-compose down -v

test:
	go test ./... -cover

lint:
	go vet ./...

swagger:
	swag init -g cmd/api-gateway/main.go -o docs/swagger --parseInternal --parseDependency
