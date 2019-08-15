# Distributted Petri Nets

Distributed Petri nets consists of an implementation of Petri nets in which arcs may cross the boundaries of a node --that is, places and transitions of the net may reside in different nodes.
The objective of this implementation is to use the Petri nets formalism to analize ad hoc systems, and prove distribution properties (e.g., deadlocks, liveliness, reachabillity) about them.


## Getting Started

## Go

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

You will need to have Go on your machine. To get started with Go [click here](https://golang.org/doc/install). File [go.mod](https://github.com/FLAGlab/DCoPN/blob/master/go.mod) contains all the necessary libraries and requires to have the environment variable GO111MODULE set to "on".

### Installing and running

Clone or download the repo into `%GOPATH%/src/github.com/FLAGlab/DCoPN`.

Now build using

```
go build
```
That should create an executable named `DCoPN`. Now you can run it using
```
DCoPN [-ctx contextName] [-p port] [-h host] [-cdr contextDependencyRelationsJsonPath] [-l leader] [peerAddress]
```
<!-- TODO: Describe what each parameter is and how the contextDependencyRelationsJson works -->

### Running the tests

All tests can be run using `go test ./...`


## Elixir
