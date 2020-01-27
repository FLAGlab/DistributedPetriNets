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
	p.AddTransition(1, 1) //hospital
	p.AddPlace(1, "ambulance", name)
	p.AddInArc(1, 1, 1)
	p.InitService(interf)
}
