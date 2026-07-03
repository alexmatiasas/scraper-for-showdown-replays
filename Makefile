GO ?= go
GOLANGCI_LINT ?= golangci-lint

.PHONY: fmt lint vet build test clean

fmt:
	$(GOLANGCI_LINT) fmt
	$(GO) fmt ./...

lint:
	$(GOLANGCI_LINT) run ./...

vet:
	$(GO) vet ./...

build:
	$(GO) build -o scraper ./cmd/scraper/

test:
	$(GO) test -v -race -count=1 ./...

clean:
	rm -f scraper
	rm -rf bin/
	rm -f coverage.out
	rm -f *.db
