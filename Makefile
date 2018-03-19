VERSION=$(shell git describe --tags --always --dirty)

all: build

.PHONY: all

build:
	go build \
		-ldflags "-X github.com/kinvolk/lerobot/cli/cmd.version=$(VERSION)" \
		-o bin/lerobot cli/main.go

.PHONY: build
