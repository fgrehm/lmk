default: build

build:
	go build ./...

xbuild:
	@mkdir -p build
	GOOS=linux GOARCH=amd64 go build -o build/linux_amd64 .
	GOOS=darwin GOARCH=amd64 go build -o build/darwin_amd64 .
	@echo 'DONE'
