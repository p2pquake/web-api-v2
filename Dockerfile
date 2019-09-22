FROM golang:latest

WORKDIR /go

RUN go get -u github.com/gin-gonic/gin && \
    go get -u github.com/gin-contrib/cors && \
    go get -u gopkg.in/olahol/melody.v1 && \
    go get -u go.mongodb.org/mongo-driver/mongo && \
    go get -u gopkg.in/go-playground/validator.v8

ADD . /go
CMD ["go", "run", "main.go"]

