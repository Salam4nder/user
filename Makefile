.PHONY: test test-cover test-db test-db/down test-db/run run api up down logs logs-user logs-db proto lint nancy redis

test: 
	go test -v ./... && $(MAKE) test-db

test-cover:
	go tool cover -html=coverage.out

test-db:
	docker compose -f internal/db/docker-compose.yaml up -d --wait
	bash -c "trap '$(MAKE) test-db/down' EXIT; $(MAKE) test-db/run"

test-db/down:
	docker compose -f internal/db/docker-compose.yaml down -v

test-db/run:
	go test -tags testdb -v --coverprofile=coverage.out -coverpkg ./... ./internal/db

api:
	docker build -t user .

redis:
	docker run --name redis -p 6379:6379 -d redis-7alpine

up:
	docker-compose up -d

down:
	docker-compose down -v

logs:
	docker-compose logs -f

logs-user:
	docker-compose logs -f api

logs-db:
	docker-compose logs -f postgres

proto:
	rm -rf internal/grpc/gen/*.go
	protoc --proto_path=pkg/proto --go_out=internal/grpc/gen --go_opt=paths=source_relative \
    --go-grpc_out=internal/grpc/gen --go-grpc_opt=paths=source_relative \
     pkg/proto/*.proto

lint:
	docker run -t --rm -v $(shell pwd):/app -v ~/.cache/golangci-lint/v1.57.2:/root/.cache -w /app golangci/golangci-lint:v1.57.2 golangci-lint run -v

nancy:
	go list -json -deps ./... | docker run --rm -i sonatypecommunity/nancy:latest sleuth
