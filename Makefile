.PHONY: build test lint install-lint clean

BINARY := fauxfile
LINT_VERSION := v1.61.0

build:
	go build -o $(BINARY) ./cmd/fauxfile

test:
	go test ./...

lint: install-lint
	golangci-lint run ./...

install-lint:
	@command -v golangci-lint >/dev/null 2>&1 || (curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(LINT_VERSION))

clean:
	rm -f $(BINARY)
