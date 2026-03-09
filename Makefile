BIN := gtc
INSTALL_DIR := $(HOME)/.local/bin
VERSION := $(shell cat VERSION)
LDFLAGS := -ldflags="-s -w -X main.version=$(VERSION)"

.PHONY: build install fmt vet build-all

build:
	go build $(LDFLAGS) -o $(BIN) main.go

install: build
	install -m 755 $(BIN) $(INSTALL_DIR)/$(BIN)

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

build-all:
	GOOS=linux  GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=$(VERSION)" -o dist/$(BIN)-linux-amd64 main.go
	GOOS=linux  GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=$(VERSION)" -o dist/$(BIN)-linux-arm64 main.go
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=$(VERSION)" -o dist/$(BIN)-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=$(VERSION)" -o dist/$(BIN)-darwin-arm64 main.go
