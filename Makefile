.PHONY: build test clean

VERSION=$(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-s -w -X main.version=${VERSION}"

build:
	go build ${LDFLAGS} -o gopac main.go

test:
	go test ./...

clean:
	rm -f gopac
