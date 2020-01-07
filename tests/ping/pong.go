package ping

import (
	pn "github.com/FLAGlab/DistributedPetriNets/petrinet"
)

func pongMain() {
	pla := &pn.Place{
		ID:    2,
		Marks: 0,
		Label: "Pong",
	}
	arc := pn.Arc{
		Place:  pla,
		Weight: 1,
	}
	tpong := pn.Transition{
		ID:       2,
		Priority: 1,
		InArcs:   make([]pn.Arc, 0),
		OutArcs:  make([]pn.Arc, 0),
	}
	tpong.AddInArc(arc)
	go pla.InitService()
}
