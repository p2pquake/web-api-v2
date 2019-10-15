FROM golang:latest

WORKDIR /go

RUN ["mkdir", "/go/app"]
ADD . /go/app

WORKDIR /go/app
CMD ["go", "run", "main.go"]

