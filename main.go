package main

import (
	"fmt"

	pn "github.com/FLAGlab/DistributedPetriNets/petrinet"
)

func main() {
	fmt.Println("init....")
	pla := &pn.Place{
		ID:    1,
		Marks: 2,
		Label: "este",
	}
	go pla.InitService()
	for {
	}
}
