BUILD_NUMBER=$(shell git rev-parse --short HEAD)

build:
	go build -v -o smoker ./cmd/smoker

clean:
	rm -f smoker
	rm -f bin/smoker*
	rm -f dist/smoker*

lint:
	go vet ./...
	staticcheck ./...

test:
	go test -race -v -covermode=atomic ./...

coverage:
	go test -race -v -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -html=coverage.txt

release: clean
	GOOS=darwin GOARCH=amd64 go build -o "bin/smoker_darwin_amd64" ./cmd/smoker
	GOOS=darwin GOARCH=386   go build -o "bin/smoker_darwin_386" ./cmd/smoker
	GOOS=linux  GOARCH=amd64 go build -o "bin/smoker_linux_amd64" ./cmd/smoker
	GOOS=linux  GOARCH=386   go build -o "bin/smoker_linux_386" ./cmd/smoker
	tar -zvcf dist/smoker-$(BUILD_NUMBER).tar.gz bin/smoker*
