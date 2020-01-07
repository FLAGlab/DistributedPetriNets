package ping

import (
	"fmt"
	pn "github.com/FLAGlab/DistributedPetriNets/petrinet"
)

func pingMain() {
	fmt.Println("init Ping net....")
	pla := &pn.Place{
		ID:    1,
		Marks: 2,
		Label: "Ping",
	}
	arc := pn.Arc{
		Place:  pla,
		Weight: 1,
	}
	tping := pn.Transition{
		ID:       1,
		Priority: 1,
		InArcs:   make([]pn.Arc, 0),
		OutArcs:  make([]pn.Arc, 0),
	}
	tping.AddInArc(arc)
	go pla.InitService()
}
