.PHONY: default build test clean install xbuild lint help

default: build

help:
	@echo "lmk - Makefile targets"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build     Build lmk for current platform (default)"
	@echo "  test      Run all tests with race detection and coverage"
	@echo "  clean     Remove build artifacts and coverage files"
	@echo "  install   Install lmk to GOPATH/bin"
	@echo "  xbuild    Cross-compile for all platforms (Linux/macOS/Windows × amd64/arm64)"
	@echo "  lint      Run golangci-lint (requires golangci-lint installed)"
	@echo "  help      Show this help message"

build:
	go build -v -ldflags="-s -w" -o lmk .

test:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

clean:
	rm -rf lmk build/ dist/ coverage.txt

install:
	go install -v -ldflags="-s -w" .

# Cross-compilation for multiple architectures
xbuild:
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

lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed" && exit 1)
	golangci-lint run ./...
