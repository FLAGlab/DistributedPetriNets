package main

import (
	"fmt"
	"os"
	pn "github.com/FLAGlab/DistributedPetriNets/petrinet"
)

func main() {
	interf := os.Args[1]
	name := os.Args[2]
	fmt.Println("init Mutex net....")
	p := pn.InitPN(0)
	// Places
	p.AddPlace(1,"exit", name)
	p.AddPlace(2,"req1", name)
	p.AddPlace(3,"req2",name)
	p.Places[1].AddMarks([]pn.Token{pn.Token{1}})

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
	
	p.InitService(interf)
}