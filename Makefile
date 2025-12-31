.PHONY: run test build migrate clean

run:
	docker compose up -d
	go run cmd/compute-api/main.go

test:
	go test ./...

build:
	mkdir -p bin
	go build -o bin/compute-api cmd/compute-api/main.go
	go build -o bin/cloud cmd/cloud-cli/main.go

migrate:
	@echo "Running migrations..."
	# Placeholder for migration command

clean:
	rm -rf bin
	docker compose down
