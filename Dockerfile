# Development/CI build image with full Go toolchain
FROM docker.io/library/golang:1.25-alpine

# Install build dependencies
RUN apk add --no-cache git make curl ca-certificates

# Install golangci-lint
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /usr/local/bin

# Set working directory
WORKDIR /build

# Default command
CMD ["/bin/sh"]
