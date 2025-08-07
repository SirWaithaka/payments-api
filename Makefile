-include .env
export

migrate: uri = "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DATABASE}?sslmode=disable"

# install the stringer binary only if not present
install-tools:
ifeq ($(shell which stringer 2>/dev/null),)
	go install golang.org/x/tools/cmd/stringer@v0.34.0
endif

install-deps: install-tools
	export GOPRIVATE=github.com/SirWaithaka && \
	go mod download

generate: install-tools
	go generate ./...

test: generate
	LOG_MODE=SILENT go test ./...

test.verbose: generate
	go test ./... -v

test.cover: generate
	go test ./... -v -coverprofile=coverage.out

migrate:
	@echo "Running migrations..."
	migrate -database ${uri} -path migrations up

build:
	mkdir -p bin
	go build -o bin/main cmd/main.go

run:
	go run cmd/main.go

run.prod:
	./main
