.PHONY: default build test clean install xbuild lint help docker-build docker-image local-build local-test local-xbuild local-lint

# Docker configuration
DOCKER_IMAGE := lmk-builder
DOCKER_RUN := docker run --rm -v $(PWD):/build -w /build $(DOCKER_IMAGE)

default: build

help:
	@echo "lmk - Makefile targets"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Docker-based targets (default):"
	@echo "  build        Build lmk for current platform using Docker (default)"
	@echo "  test         Run all tests with race detection and coverage in Docker"
	@echo "  xbuild       Cross-compile for all platforms in Docker"
	@echo "  lint         Run golangci-lint in Docker"
	@echo "  install      Build and symlink lmk to ~/.local/bin"
	@echo ""
	@echo "Docker management:"
	@echo "  docker-image Build the Docker image for compilation"
	@echo "  docker-build Build lmk binary inside Docker and copy to host"
	@echo ""
	@echo "Local targets (no Docker):"
	@echo "  local-build   Build lmk for current platform locally"
	@echo "  local-test    Run tests locally"
	@echo "  local-xbuild  Cross-compile locally"
	@echo "  local-lint    Run golangci-lint locally"
	@echo "  local-install Install to GOPATH/bin using go install"
	@echo ""
	@echo "Other:"
	@echo "  clean        Remove build artifacts and coverage files"
	@echo "  help         Show this help message"

# Docker image builder
docker-image:
	docker build -t $(DOCKER_IMAGE) .

# Main build target using Docker
build: docker-image
	$(DOCKER_RUN) go build -v -ldflags="-s -w" -o lmk .

# Build Docker image with binary
docker-build:
	docker build -t lmk:latest .

# Local build (without Docker)
local-build:
	go build -v -ldflags="-s -w" -o lmk .

# Test using Docker
test: docker-image
	$(DOCKER_RUN) go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

# Local test (without Docker)
local-test:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

clean:
	rm -rf lmk build/ dist/ coverage.txt

# Install using Docker - builds binary and symlinks to ~/.local/bin
install: build
	@mkdir -p ~/.local/bin
	@ln -sf $(PWD)/lmk ~/.local/bin/lmk
	@echo "Symlinked $(PWD)/lmk -> ~/.local/bin/lmk"
	@echo "Make sure ~/.local/bin is in your PATH"

# Local install (without Docker)
local-install:
	go install -v -ldflags="-s -w" .

# Cross-compilation for multiple architectures using Docker
xbuild: docker-image
	@mkdir -p build
	@echo "Building for Linux amd64..."
	$(DOCKER_RUN) sh -c "GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o build/lmk-linux-amd64 ."
	@echo "Building for Linux arm64..."
	$(DOCKER_RUN) sh -c "GOOS=linux GOARCH=arm64 go build -ldflags='-s -w' -o build/lmk-linux-arm64 ."
	@echo "Building for macOS amd64..."
	$(DOCKER_RUN) sh -c "GOOS=darwin GOARCH=amd64 go build -ldflags='-s -w' -o build/lmk-darwin-amd64 ."
	@echo "Building for macOS arm64..."
	$(DOCKER_RUN) sh -c "GOOS=darwin GOARCH=arm64 go build -ldflags='-s -w' -o build/lmk-darwin-arm64 ."
	@echo "Building for Windows amd64..."
	$(DOCKER_RUN) sh -c "GOOS=windows GOARCH=amd64 go build -ldflags='-s -w' -o build/lmk-windows-amd64.exe ."
	@echo "Building for Windows arm64..."
	$(DOCKER_RUN) sh -c "GOOS=windows GOARCH=arm64 go build -ldflags='-s -w' -o build/lmk-windows-arm64.exe ."
	@echo 'DONE'

# Local cross-compilation (without Docker)
local-xbuild:
	@mkdir -p build
	@echo "Building for Linux amd64..."
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o build/lmk-linux-amd64 .
	@echo "Building for Linux arm64..."
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o build/lmk-linux-arm64 .
	@echo "Building for macOS amd64..."
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o build/lmk-darwin-amd64 .
	@echo "Building for macOS arm64..."
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o build/lmk-darwin-arm64 .
	@echo "Building for Windows amd64..."
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o build/lmk-windows-amd64.exe .
	@echo "Building for Windows arm64..."
	GOOS=windows GOARCH=arm64 go build -ldflags="-s -w" -o build/lmk-windows-arm64.exe .
	@echo 'DONE'

# Lint using Docker
lint: docker-image
	$(DOCKER_RUN) sh -c "which golangci-lint > /dev/null || (echo 'golangci-lint not installed in Docker image' && exit 1); golangci-lint run ./..."

# Local lint (without Docker)
local-lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed" && exit 1)
	golangci-lint run ./...
