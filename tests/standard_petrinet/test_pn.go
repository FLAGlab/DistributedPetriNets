package main

import (
	"fmt"
	pn "github.com/FLAGlab/DistributedPetriNets/petrinet"
)

func main() {
	fmt.Println("init Test net....")
	p := pn.InitPN(0)
	p.AddPlace(1, 1, "ping")
	p.AddPlace(2, 0, "pong")
	p.AddTransition(1, 1)
	p.AddTransition(2, 1)
	p.AddInArc(1, 1, 1)
	p.AddOutArc(1, 2, 1)
	p.AddInArc(2, 2, 1)
	p.AddOutArc(2, 1, 1)
	p.InitService()
}