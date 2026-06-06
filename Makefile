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
	go test -v ./... $(ARGS)