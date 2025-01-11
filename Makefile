.PHONY: all build run clean test

# 构建目标
APISERVER_BINARY=i18n-apiserver

all: build

build: 
	go build -o bin/$(APISERVER_BINARY) cmd/i18n-apiserver/apiserver.go

run-api:
	go run cmd/i18n-apiserver/apiserver.go

test:
	go test -v ./...

clean:
	rm -f bin/$(APISERVER_BINARY)