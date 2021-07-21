FROM golang:latest

RUN mkdir -p /go/src/postgresApi

WORKDIR /go/src/postgresApi

COPY . /go/src/postgresApi

RUN go install postgresApi

CMD /go/bin/postgresApi

EXPOSE 3001