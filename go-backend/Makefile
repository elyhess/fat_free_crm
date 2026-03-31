.PHONY: test test-verbose build

# Run tests sequentially across packages (-p 1) because multiple packages
# share the same PostgreSQL test database and would corrupt each other's
# data if run in parallel.
test:
	go test -p 1 ./internal/...

test-verbose:
	go test -p 1 -v ./internal/...

build:
	go build -o bin/server ./cmd/server
