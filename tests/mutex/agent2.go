package main

import (
	"fmt"
	pn "github.com/FLAGlab/DistributedPetriNets/petrinet"
)

func main() {
	fmt.Println("init Agent2 net....")
	p := pn.InitPN(0)
	// Places
	p.AddPlace(1, 0, "grant2")
	p.AddPlace(2, 0, "exec2")
	p.AddPlace(3, 1, "rel2")
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
	p.AddRemoteOutArc(3, 1, "req2")

	p.InitService()
}