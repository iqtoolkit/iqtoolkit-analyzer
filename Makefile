BINARY  := iqtoolkit-analyzer
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    ?= $(shell date -u +%Y-%m-%d)
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

.PHONY: build test lint vet cover clean install

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

test:
	go test -race ./...

cover:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1

vet:
	go vet ./...

lint:
	golangci-lint run

install:
	go install -ldflags "$(LDFLAGS)" .

clean:
	rm -f $(BINARY) coverage.out
