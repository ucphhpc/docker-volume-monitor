FROM golang:1.26-alpine
RUN apk add git tzdata

WORKDIR /go/src/docker-volume-monitor

COPY src src
COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod vendor
RUN go build ./...
RUN go install -v ./...

ENTRYPOINT ["/go/bin/docker-volume-monitor"]
CMD ["-prune-unused", "-interval", "10"]