.PHONY: build run test lint clean release-snapshot

BINARY := bin/brr

build:
	go build -o $(BINARY) ./cmd/brr

run: build
	./$(BINARY)

test:
	go test ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/ dist/

release-snapshot:
	goreleaser build --snapshot --clean
