lint:
	gofumpt -w .
	go mod tidy
	golangci-lint run ./...

up:
	docker-compose up -d

down:
	docker-compose down

rebuild:
	docker-compose up -d --remove-orphans --force-recreate --build auth-service