.PHONY: all clean build-mac release-tar test

VERSION ?= 0.1.0

all: build-mac

build-mac: build-mac-amd64 build-mac-arm64

build-mac-amd64:
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X xzap/cmd.Version=$(VERSION)" -o bin/xzap_darwin_amd64 main.go

build-mac-arm64:
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X xzap/cmd.Version=$(VERSION)" -o bin/xzap_darwin_arm64 main.go

test:
	go test -v ./...

clean:
	rm -rf bin/* dist/*

release-tar:
	mkdir -p dist
	tar -czvf dist/xzap_darwin_amd64.tar.gz -C bin xzap_darwin_amd64
	tar -czvf dist/xzap_darwin_arm64.tar.gz -C bin xzap_darwin_arm64

goreleaser:
	goreleaser release --clean --skip-validate --skip-publish