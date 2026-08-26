.PHONY: dev build test seed

dev:
	go run ./cmd/server

build:
	cd web && npm run build
	go build -o bin/sheep ./cmd/server

test:
	go test ./...

seed:
	rm -f data/sheep.json
	go run ./cmd/server

