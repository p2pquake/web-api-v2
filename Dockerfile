FROM golang:latest

WORKDIR /go

RUN go get -u github.com/gin-gonic/gin && \
    go get -u gopkg.in/olahol/melody.v1 && \
    go get -u go.mongodb.org/mongo-driver/mongo

ADD . /go

CMD ["go", "run", "main.go"]

