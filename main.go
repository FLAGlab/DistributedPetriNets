package main

import (
	"fmt"

	comm "github.com/FLAGlab/DistributedPetriNets/communication"
)

func main() {
	fmt.Println("init....")
	serv := &comm.ServiceNode{
		ServiceName: "echo",
	}
	go serv.RunService()
	for {
	}
}
