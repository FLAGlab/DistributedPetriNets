package main

import (
	"fmt"
	"os"
	pn "github.com/FLAGlab/DistributedPetriNets/petrinet"
)

func main() {
	interf := os.Args[1]
	fmt.Println("init Ping net....")
	p := pn.InitPN(0)
	p.AddTransition(1, 1)
	p.AddPlace(1, 1, "ping")
	p.AddInArc(1, 1, 1)
	p.AddRemoteOutArc(1, 1, "pong")
	p.InitService(interf)
}
