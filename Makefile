-include .env
export

# install the stringer binary only if not present
install-tools:
ifeq ($(shell which stringer 2>/dev/null),)
	go install golang.org/x/tools/cmd/stringer@v0.24.0
endif

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
