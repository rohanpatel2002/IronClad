.PHONY: build run test clean fmt lint

build:
	go build -o bin/gate-go main.go

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

all: fmt lint test build
