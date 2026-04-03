.PHONY: test test-verbose build build-frontend migrate migrate-status migrate-create migrate-down migrate-mark-baseline

# Run tests sequentially across packages (-p 1) because multiple packages
# share the same PostgreSQL test database and would corrupt each other's
# data if run in parallel.
test:
	go test -p 1 ./internal/...

test-verbose:
	go test -p 1 -v ./internal/...

build: build-frontend
	go build -o bin/server ./cmd/server
	go build -o bin/migrate ./cmd/migrate

# Build React frontend and copy into go:embed directory
build-frontend:
	cd ../react-frontend && npm run build
	rm -rf internal/frontend/dist
	cp -r ../react-frontend/dist internal/frontend/dist

# Database migrations (via goose)
migrate:
	go run ./cmd/migrate up

migrate-status:
	go run ./cmd/migrate status

migrate-down:
	go run ./cmd/migrate down

migrate-create:
	@test -n "$(NAME)" || (echo "Usage: make migrate-create NAME=add_foo_table" && exit 1)
	go run ./cmd/migrate create $(NAME)

migrate-mark-baseline:
	go run ./cmd/migrate mark-applied
