default: build

build:
	@go build ./...

xbuild:
	@gox -osarch="linux/386 linux/amd64 darwin/386 darwin/amd64" -output="build/{{.Dir}}_{{.OS}}_{{.Arch}}" -verbose
