package main

import (
	"fmt"
	pn "github.com/FLAGlab/DistributedPetriNets/petrinet"
	"os"
)

func main() {
	interf := os.Args[1]
	nombre := os.Args[2]
	fmt.Printf("init Vehivle net.... %v\n", interf)
	p := pn.InitPN(0)
	p.AddTransition(1, 1) //send
	p.AddTransition(2, 1) //commit
	p.AddPlace(1, "car", nombre)
	p.AddPlace(2, "vsadb", nombre)
	p.AddInArc(1, 2, 1)
	p.AddOutArc(2, 2, 1)
	p.AddInArc(2, 1, 1)
	p.AddRemoteOutArc(1, 1, "rsu")
	p.InitService(interf)
}
