FROM golang:latest

WORKDIR /go
ADD . /go

RUN go get -u github.com/gin-gonic/gin

CMD ["go", "run", "main.go"]

