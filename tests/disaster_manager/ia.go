package main

import (
	"fmt"
	pn "github.com/FLAGlab/DistributedPetriNets/petrinet"
	"os"
)

func main() {
	interf := os.Args[1]
	name := os.Args[2]
	fmt.Printf("init SENSOR net.... %v\n", interf)
	p := pn.InitPN(0)
	p.AddTransition(1, 1) //assign
	p.AddPlace(1, "patients", name)
	p.Places[1].AddMarks([]pn.Token{pn.Token{2}, pn.Token{1}, pn.Token{3}, pn.Token{4}, pn.Token{5}})
	p.AddInArc(1, 1, 1)
	p.AddRemoteOutArc(1, 1, "police")
	p.AddRemoteOutArc(1, 1, "fire")
	p.InitService(interf)
}
