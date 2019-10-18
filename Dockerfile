FROM golang:1.10
LABEL maintainer="Juan Sosa <juansesosaajedrez3@gmail.com>"


ENV GOPATH /go
ENV PATH $GOPATH/bin:$PATH

RUN apt-get update && apt-get install -y libzmq3-dev

WORKDIR $GOPATH