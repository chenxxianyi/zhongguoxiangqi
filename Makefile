.PHONY: dev-api dev-worker test vet build frontend-check

dev-api:
	go run ./cmd/api

dev-worker:
	go run ./cmd/worker

test:
	go test ./...

vet:
	go vet ./...

build:
	go build ./cmd/api ./cmd/worker ./cmd/tools/perft ./cmd/tools/enginebench

frontend-check:
	npm run typecheck:web
	npm run test:web
	npm run lint:web
	npm run build:web

