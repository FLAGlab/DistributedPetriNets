package main

import (
	"fmt"
	pn "github.com/FLAGlab/DistributedPetriNets/petrinet"
	"os"
)

func main() {
	interf := os.Args[1]
	nombre := os.Args[2]
	fmt.Printf("init RSU net.... %v\n", interf)
	p := pn.InitPN(0)
	p.AddTransition(1, 1) //push
	p.AddTransition(2, 1) //commit
	p.AddPlace(1, "rsu", nombre)
	p.AddPlace(2, "sadb", nombre)
	//p.Places[1].AddMarks([]pn.Token{pn.Token{1}})
	p.AddInArc(1, 2, 1)
	p.AddOutArc(2, 2, 1)
	p.AddInArc(2, 1, 1)
	p.AddRemoteOutArc(1, 1, "car")
	p.InitService(interf)
}
