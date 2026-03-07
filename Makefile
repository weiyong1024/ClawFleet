BINARY     = clawsandbox
BUILD_DIR  = ./bin
MODULE     = github.com/weiyong1024/clawsandbox
IMAGE      = clawsandbox/openclaw:latest

.PHONY: build build-all docker-build install clean tidy

build:
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/clawsandbox

build-all:
	GOOS=darwin  GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY)-darwin-arm64  ./cmd/clawsandbox
	GOOS=darwin  GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY)-darwin-amd64  ./cmd/clawsandbox
	GOOS=linux   GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY)-linux-amd64   ./cmd/clawsandbox
	GOOS=linux   GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY)-linux-arm64   ./cmd/clawsandbox

docker-build:
	docker build -t $(IMAGE) -f internal/assets/docker/Dockerfile internal/assets/docker/

install: build
	cp $(BUILD_DIR)/$(BINARY) /usr/local/bin/$(BINARY)

tidy:
	go mod tidy

clean:
	rm -rf $(BUILD_DIR)
