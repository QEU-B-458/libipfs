.PHONY: default build fmt lint run run_race test clean vet docker_build docker_run docker_clean

LIB_NAME := libipfs
default: build

clean:
		rm -rf bin/ && mkdir bin/

build_linux_x64:
		CGO_ENABLED=1 \
		GOOS=linux \
		GOARCH=amd64 \
		go build -buildmode=c-archive -o ./bin/${LIB_NAME}-linux.a ./src/${LIB_NAME}.go

build_windows_x64:
		CC=x86_64-w64-mingw32-gcc \
		CGO_ENABLED=1 \
		GOOS=windows \
		GOARCH=amd64 \
		go build -buildmode=c-archive -o ./bin/${LIB_NAME}-windows.a ./src/${LIB_NAME}.go

build_darwin_arm64:
		CGO_ENABLED=1 \
		GOOS=darwin \
		GOARCH=arm64 \
		go build -buildmode=c-archive -o ./bin/${LIB_NAME}-darwin-arm64.a ./src/${LIB_NAME}.go

build-all: build_linux_x64  build_windows_x64  build_darwin_arm64 

build: clean \
	build-all
