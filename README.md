# Distributed Petri Nets

Distributed Petri nets consists of an implementation of Petri nets in which arcs may cross the boundaries of a node --that is, places and transitions of the net may reside in different nodes.
The objective of this implementation is to use the Petri nets formalism to analize ad hoc systems, and prove distribution properties (e.g., deadlocks, liveliness, reachabillity) about them.


## Getting Started

## Go

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

You will need to have Go on your machine. To get started with Go [click here](https://golang.org/doc/install). File [go.mod](https://github.com/FLAGlab/DCoPN/blob/master/go.mod) contains all the necessary libraries and requires to have the environment variable GO111MODULE set to "on".

### Installing and running

The code is contained in a docker container

`docker build -t dpn .`

`docker run -i -t --network host -v "$(pwd)":/go/src/github.com/FLAGlab/DistributedPetriNets dpn`

You can run the code by going to the `go/src/github.com/FLAGlab/DistributedPetriNets` and then running the main using the command `go run main.go`

### Running the tests

All tests can be run using `go test ./...`
