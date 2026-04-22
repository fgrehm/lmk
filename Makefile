.PHONY: help build test xbuild install clean lint fmt

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the lmk binary
	@echo "Building lmk..."
	@mkdir -p dist
	@go build -ldflags="-s -w" -o dist/lmk .
	@echo "Built to dist/lmk"

test: ## Run tests with race detection and coverage
	@go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

xbuild: ## Cross-compile for all platforms
	@mkdir -p dist
	@echo "Building for Linux amd64..."
	@GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/lmk-linux-amd64 .
	@echo "Building for Linux arm64..."
	@GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o dist/lmk-linux-arm64 .
	@echo "Building for macOS amd64..."
	@GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/lmk-darwin-amd64 .
	@echo "Building for macOS arm64..."
	@GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/lmk-darwin-arm64 .
	@echo "Building for Windows amd64..."
	@GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/lmk-windows-amd64.exe .
	@echo "Building for Windows arm64..."
	@GOOS=windows GOARCH=arm64 go build -ldflags="-s -w" -o dist/lmk-windows-arm64.exe .
	@echo "Done"

install: build ## Install lmk to ~/.local/bin (symlink)
	@mkdir -p "$(HOME)/.local/bin"
	@ln -sf "$(CURDIR)/dist/lmk" "$(HOME)/.local/bin/lmk"
	@echo "Installed to ~/.local/bin/lmk"

clean: ## Remove build artifacts
	@rm -rf dist/ coverage.txt
	@echo "Cleaned"

lint: ## Run golangci-lint
	@go tool golangci-lint run ./...

fmt: ## Format code with golangci-lint
	@go tool golangci-lint fmt ./...
