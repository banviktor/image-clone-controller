FROM golang:1.16-alpine3.13 AS builder

RUN apk add --no-cache make
WORKDIR /go/src/image-clone-controller
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make

FROM alpine:3.13
COPY --from=builder /go/src/image-clone-controller/build/image-clone-controller /
USER nobody
ENTRYPOINT ["/image-clone-controller"]
CMD []
