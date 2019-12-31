FROM golang:1.10
LABEL maintainer="Juan Sosa <juansesosaajedrez3@gmail.com>"


ENV GOPATH /go
ENV PATH $GOPATH/bin:$PATH

RUN apt-get update && apt-get install -y libzmq3-dev
RUN go get -u github.com/ursiform/sleuth
RUN mkdir /go/src/github.com/FLAGlab
ADD . /go/src/github.com/FLAGlab/DistributedPetriNets

WORKDIR $GOPATH