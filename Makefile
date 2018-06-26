.PHONY: all

all: fmt vet test build
	echo "success"

fmt:
	go fmt ./...

vet:
	go vet -v ./...

test:
	go test -race -cover ./...

build:
	go build
