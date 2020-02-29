BUILD_NUMBER=$(shell git rev-parse --short HEAD)

setup:
	go mod download
	docker pull golangci/golangci-lint:latest-alpine
.PHONY: setup

build:
	go build -v -o smoker ./cmd/smoker
.PHONY: build

clean:
	rm -f smoker
	rm -f bin/smoker*
	rm -f dist/smoker*
.PHONY: clean

lint:
	go vet ./...
	staticcheck ./...
	docker run --rm -v ${PWD}:/app -w /app golangci/golangci-lint:latest-alpine golangci-lint run
.PHONY: lint

test:
	go test -race -v -covermode=atomic ./...
.PHONY: test

coverage:
	go test -race -v -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -html=coverage.txt
.PHONY: coverage

release: clean
	GOOS=darwin GOARCH=amd64 go build -o "bin/smoker_darwin_amd64" ./cmd/smoker
	GOOS=darwin GOARCH=386   go build -o "bin/smoker_darwin_386" ./cmd/smoker
	GOOS=linux  GOARCH=amd64 go build -o "bin/smoker_linux_amd64" ./cmd/smoker
	GOOS=linux  GOARCH=386   go build -o "bin/smoker_linux_386" ./cmd/smoker
	tar -zvcf dist/smoker-$(BUILD_NUMBER).tar.gz bin/smoker*
.PHONY: release

docker:
	docker build --tag smoker:(git describe --abbrev=0) .
.PHONY: docker
