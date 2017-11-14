.PHONY: build

PACKAGES = $(shell go list ./... | grep -v /vendor/)

ifneq ($(shell uname), Darwin)
	EXTLDFLAGS = -extldflags "-static" $(null)
else
	EXTLDFLAGS =
endif

BUILD_NUMBER=$(shell git rev-parse --short HEAD)

all: build

test:
	go test -cover $(PACKAGES)

build: build_static 

build_static:
	mkdir -p make/release
	go build -o  make/release/kubepipe -ldflags '${EXTLDFLAGS}-X github.com/arlert/kubepipe/version.VersionDev=build.$(BUILD_NUMBER)' github.com/arlert/kubepipe

build_cross:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '-X github.com/arlert/kubepipe/version.VersionDev=build.$(BUILD_NUMBER)' -o make/release/linux/amd64/kubepipe   github.com/arlert/kubepipe
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '-X github.com/arlert/kubepipe/version.VersionDev=build.$(BUILD_NUMBER)' -o make/release/darwin/amd64/kubepipe   github.com/arlert/kubepipe

build_tar: build_cross
	tar -cvzf make/release/linux/amd64/kubepipe.tar.gz   -C make/release/linux/amd64/kubepipe
	tar -cvzf make/release/darwin/amd64/kubepipe.tar.gz  -C make/release/darwin/amd64/kubepipe
