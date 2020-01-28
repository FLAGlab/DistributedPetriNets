package main

import (
	"fmt"
	"os"
	pn "github.com/FLAGlab/DistributedPetriNets/petrinet"
)

func main() {
	interf := os.Args[1]
	name := os.Args[2]
	fmt.Println("init Agent2 net....")
	p := pn.InitPN(0)
	// Places
	p.AddPlace(1,"grant2", name)
	p.AddPlace(2, "exec2", name)
	p.AddPlace(3,"rel2", name)
	p.Places[3].AddMarks([]pn.Token{pn.Token{1}})

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

	p.InitService(interf)
}