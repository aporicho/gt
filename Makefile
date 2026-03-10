BIN := gt
INSTALL_DIR := $(HOME)/.local/bin
VERSION := $(shell cat VERSION)
LDFLAGS := -ldflags="-s -w -X main.version=$(VERSION)"

.PHONY: build install fmt vet build-all clean

build:
	go build $(LDFLAGS) -o $(BIN) .

install: build
	install -m 755 $(BIN) $(INSTALL_DIR)/$(BIN)
	ln -sf $(BIN) $(INSTALL_DIR)/gtc

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

clean:
	rm -f $(BIN) gtc dist/*

build-all:
	GOOS=linux  GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=$(VERSION)" -o dist/$(BIN)-linux-amd64 .
	GOOS=linux  GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=$(VERSION)" -o dist/$(BIN)-linux-arm64 .
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=$(VERSION)" -o dist/$(BIN)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=$(VERSION)" -o dist/$(BIN)-darwin-arm64 .
