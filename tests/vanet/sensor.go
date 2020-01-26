package main

import (
	"fmt"
	pn "github.com/FLAGlab/DistributedPetriNets/petrinet"
	"os"
)

func main() {
	interf := os.Args[1]
	fmt.Printf("init SENSOR net.... %v\n", interf)
	p := pn.InitPN(0)
	p.AddTransition(1, 1) //generator
	p.AddTransition(2, 1) //flush
	p.AddPlace(1, "ism")
	p.AddPlace(2, "generator")
	p.AddInArc(2, 1, 1)
	p.AddOutArc(1, 1, 1)
	p.AddInArc(1, 2, 1)
	p.Places[2].AddMarks([]pn.Token{pn.Token{2}, pn.Token{1}, pn.Token{3}, pn.Token{4}, pn.Token{5}})
	p.AddRemoteOutArc(2, 1, "car")
	p.InitService(interf)
}
