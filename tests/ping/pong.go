package main

import (
	pn "github.com/FLAGlab/DistributedPetriNets/petrinet"
	"fmt"
	"os"
)

func main() {
	interf := os.Args[1]
	fmt.Printf("init Ping net.... %v\n", interf)
	p := pn.InitPN(0)
	p.AddTransition(1, 1)
	p.AddPlace(1, "pong")
	p.AddInArc(1, 1, 1)
	p.AddRemoteOutArc(1, 1, "ping")
	p.InitService(interf)
}
