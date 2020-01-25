package main

import (
	"fmt"
	"os"
	pn "github.com/FLAGlab/DistributedPetriNets/petrinet"
)

func main() {
	interf := os.Args[1]
	fmt.Printf("init Ping net.... %v\n", interf)
	p := pn.InitPN(0)
	p.AddTransition(1, 1)
	p.AddPlace(1, "ping")
	p.Places[1].AddMarks([]pn.Token{pn.Token{1}})
	p.AddInArc(1, 1, 1)
	p.AddRemoteOutArc(1, 1, "pong")
	p.InitService(interf)
}
