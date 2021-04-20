FROM golang:1.16-alpine3.13 AS builder

WORKDIR /go/src/github.com/banviktor/image-clone-controller
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /go/bin/ ./cmd/image-clone-controller

FROM alpine:3.13
COPY --from=builder /go/bin/image-clone-controller /
USER nobody
ENTRYPOINT ["/image-clone-controller"]
CMD []
