default: build

build:
	@go build ./...

xbuild:
	@gox -osarch="linux/386 linux/amd64 darwin/386 darwin/amd64" -output="build/{{.Dir}}_{{.OS}}_{{.Arch}}" -verbose

release: xbuild
	@test -z '$(version)' && echo 'version parameter not provided to `make`!' && exit 1 || return 0
	@gh release create -d -a build/ $(version)
	@git tag $(version)
	@git push && git push --tags
