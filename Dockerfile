FROM golang:latest as builder
WORKDIR /go/src
ADD . /go/src
RUN CGO_ENABLED=0 go build . && ls -l /go/src

FROM alpine:latest
WORKDIR /go
COPY --from=builder /go/src/web-api-v2 .
CMD ["./web-api-v2"]

