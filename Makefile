lint:
	go mod tidy
	golangci-lint run ./...

up:
	docker-compose up -d db

down:
	docker_compose down

test: up
	go test -failfast -v ./...