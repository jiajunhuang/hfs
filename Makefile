.PHONY: all

all: protogen fmt vet test build
	echo "success"

fmt:
	go fmt ./...

vet:
	go vet -v ./...

test:
	go test -race -cover ./...

build:
	go build -o bin/chunkserver cmd/chunkserver/main.go
	go build -o bin/hfsclient cmd/hfsclient/main.go

protogen:
	protoc -I pb pb/service.proto --go_out=plugins=grpc:pb
