package main

import (
	"fmt"
	pn "github.com/FLAGlab/DistributedPetriNets/petrinet"
)

func main() {
	fmt.Println("init Mutex net....")
	p := pn.InitPN(0)
	// Places
	p.AddPlace(1, 1, "exit")
	p.AddPlace(2, 0, "req1")
	p.AddPlace(3, 0, "req2")
	// Transitions
	p.AddTransition(1, 1)
	p.AddTransition(2, 1)

	// Arcs from exit
	p.AddInArc(1, 1, 1)
	p.AddInArc(1, 2, 1)
	// Arcs from req1
	p.AddInArc(2, 1, 1)
	// Arcs from req2
	p.AddInArc(3, 2, 1)

	//Remote Arcs
	p.AddRemoteOutArc(1, 1, "grant1")
	p.AddRemoteOutArc(2, 1, "grant2")
	
	p.InitService()
}