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
	p.AddPlace(1, "police", name)
	p.AddInArc(1, 1, 1)
	p.AddRemoteOutArc(1, 1, "wait")
	p.InitService(interf)
}
