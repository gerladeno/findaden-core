lint:
	gofumpt -w .
	go mod tidy
	golangci-lint run ./...

test:
	docker-compose -f test-db-docker-compose.yml up -d
	sleep 5
	go test -race ./...
	docker-compose -f test-db-docker-compose.yml down

up:
	docker-compose up -d

down:
	docker-compose down

rebuild:
	docker-compose up -d --remove-orphans --force-recreate --build core-service