VERSION=$(shell git describe --tags --always --dirty)

.PHONY: all
all: build

.PHONY: build
build:
	go build \
		-tags osusergo,netgo -ldflags "-linkmode external -extldflags '-static' -s -w -X github.com/kinvolk/lerobot/cli/cmd.version=$(VERSION)" \
		-o bin/lerobot cli/main.go

.PHONY: test
test:
	go test -v ./...
