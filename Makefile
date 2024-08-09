build:
	go build -o build/pocket-exporter

test:
	go clean -testcache
	go test -v -cover -race ./...

lint:
	golangci-lint run ./...

.PHONY: build test lint
