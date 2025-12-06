-include .env
export

# build postgres uri from env variables
POSTGRES_URI = "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DATABASE}?sslmode=disable"
# add postgres uri as variable to commands
migrate: uri = $(POSTGRES_URI)
seed: uri = $(POSTGRES_URI)
seed.down: uri = $(POSTGRES_URI)

# install the stringer binary only if not present
install-tools:
ifeq ($(shell which stringer 2>/dev/null),)
	go install golang.org/x/tools/cmd/stringer@v0.34.0
endif

install-deps: install-tools
	export GOPRIVATE=github.com/SirWaithaka && \
	go mod download

install-linters:
	GO111MODULE=on go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

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

seed: ## Execute all *.up.sql files to seed the database
	@echo "Seeding database with up migrations..."
	@for file in $$(find seeds -name "*.up.sql" | sort); do \
		echo "Executing: $$file"; \
		psql "${uri}" -f "$$file" || exit 1; \
	done
	@echo "Database seeding completed successfully!"

seed.down: ## Execute all *.down.sql files to tear down seeded data
	@echo "Executing down migrations..."
	@for file in $$(find seeds -name "*.down.sql" | sort -r); do \
		echo "Executing: $$file"; \
		psql "${uri}" -f "$$file" || exit 1; \
	done
	@echo "Database teardown completed successfully!"

build:
	mkdir -p bin
	go build -o bin/payments main.go

run:
	go run cmd/main.go

run.prod:
	./main
