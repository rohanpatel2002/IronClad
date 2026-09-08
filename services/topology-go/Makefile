.PHONY: build run test clean fmt lint

build:
	go build -o bin/topology-go main.go

run:
	go run main.go

test:
	go test -v ./...

clean:
	rm -rf bin/

fmt:
	go fmt ./...

lint:
	go vet ./...

bench:
	go test -bench=. -benchmem ./graph/...

all: fmt lint test build

