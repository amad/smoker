FROM golang:1.13-alpine

LABEL maintainer="Ahmad Samiei"

ENV GOPATH /go
ENV CGO_ENABLED 0
ENV GO111MODULE on

RUN  \
    apk add --no-cache git && \
    git clone https://github.com/amad/smoker && cd smoker && \
    go install -v -ldflags "-s -w -X github.com/amad/smoker/version.version=$(git describe --abbrev=0)" ./cmd/smoker

FROM alpine:3.10

COPY --from=0 /go/bin/smoker /usr/bin/smoker

RUN \
    chmod +x /usr/bin/smoker

VOLUME ["/data"]

ENTRYPOINT ["/usr/bin/smoker"]

CMD ["/usr/bin/smoker"]
