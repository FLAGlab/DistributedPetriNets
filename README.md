# DCoPN

<!-- TODO: One Paragraph of project description goes here -->

## Getting Started

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

## Running the tests

All tests can be run using `go test ./...`
