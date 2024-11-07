.PHONY: build test clean install release

VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT ?= $(shell git rev-parse --short HEAD)
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS = -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}

build:
	go build -ldflags "${LDFLAGS}" -o bin/wex ./cmd/wex

test:
	go test -v ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/ dist/

install: build
	install -d ${DESTDIR}/usr/local/bin/
	install -m 755 bin/wex ${DESTDIR}/usr/local/bin/

snapshot:
	goreleaser release --snapshot --rm-dist

release:
	goreleaser release --rm-dist

# Development helpers
dev: build
	./bin/wex

update-deps:
	go get -u
	go mod tidy

docker:
	docker build -t wex .