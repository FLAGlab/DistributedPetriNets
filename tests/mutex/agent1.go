package main

import (
	"fmt"
	pn "github.com/FLAGlab/DistributedPetriNets/petrinet"
)

func main() {
	fmt.Println("init Agent1 net....")
	p := pn.InitPN(0)
	// Places
	p.AddPlace(1, 0, "grant1")
	p.AddPlace(2, 0, "exec1")
	p.AddPlace(3, 1, "rel1")
	// Transitions
	p.AddTransition(1, 1)
	p.AddTransition(2, 1)
	p.AddTransition(3, 1)
	
	// Arcs with e
	p.AddInArc(1, 1, 1)
	p.AddOutArc(1, 2, 1)
	// Arcs with x
	p.AddInArc(2, 2, 1)
	p.AddRemoteOutArc(2, 1, "exit")
	p.AddOutArc(2, 3, 1)
	//Arcs with r
	p.AddInArc(3, 3, 1)
	p.AddRemoteOutArc(3, 1, "req1")

	p.InitService()
}