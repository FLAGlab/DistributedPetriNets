package main

import (
	"fmt"
	pn "github.com/FLAGlab/DistributedPetriNets/petrinet"
	"os"
)

func main() {
	interf := os.Args[1]
	name := os.Args[2]
	fmt.Printf("init Ping net.... %v\n", interf)
	p := pn.InitPN(0)
	p.AddTransition(1, 1)
	p.AddPlace(1, "ping", name)
	p.Places[1].AddMarks([]pn.Token{pn.Token{1}})
	p.AddInArc(1, 1, 1)
	p.AddRemoteOutArc(1, 1, "pong")
	p.InitService(interf)
}
