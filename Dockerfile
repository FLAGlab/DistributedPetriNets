FROM golang:1.17
LABEL maintainer="Juan Sosa <juansesosaajedrez3@gmail.com>"


ENV GOPATH /go
ENV PATH $GOPATH/bin:$PATH

RUN apt-get update && apt-get install -y libzmq3-dev net-tools vim
RUN go install github.com/FLAGlab/DistributedPetriNets/dpn@v1.0.4

WORKDIR /data

COPY ./configs .

ENTRYPOINT ["dpn", "run"]