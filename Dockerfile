FROM golang:1.23.12-alpine
RUN apk add git tzdata

WORKDIR /go/src/docker-volume-monitor
COPY . .

RUN go mod vendor
RUN go build ./...
RUN go install -v ./...

ENTRYPOINT ["/go/bin/docker-volume-monitor"]
CMD ["-prune-unused", "-interval", "10"]