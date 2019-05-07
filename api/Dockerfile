FROM golang:1.12.1

RUN go get -u github.com/lib/pq

RUN mkdir /api

ADD . /api/

WORKDIR /api

RUN go build -o main .
CMD ["/api/main"]