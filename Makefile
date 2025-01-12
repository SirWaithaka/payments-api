-include .env
export

install-tools:
	go install golang.org/x/tools/cmd/stringer@v0.24.0

install-deps: install-tools
	export GOPRIVATE=github.com/fingoafrica && \
	go mod download

generate: install-tools
	go generate ./...

test: generate
	LOG_MODE=SILENT go test ./...

test.verbose: generate
	go test ./... -v

test.cover: generate
	go test ./... -v -coverprofile=coverage.out
