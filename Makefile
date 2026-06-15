ARGS ?=

run:
	go run ./cmd/app
dev:
	air
lint:
	golangci-lint run
lint-fix:
	golangci-lint run --fix
test:
	go test -tags=integration -v ./... $(ARGS)
test-race:
	go test -tags=integration -v -race ./...
