FROM golang:1.12.1

WORKDIR /go/src/github.com/allanvieira/microservico-go/api
ADD . .

RUN go get
RUN go build -o main .
CMD ["go", "run", "main.go"]