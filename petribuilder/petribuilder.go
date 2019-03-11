package petribuilder

import (
  "github.com/FLAGlab/DCoPN/petrinet"
)

// BuildPetriNet builds a test petri net
func BuildPetriNet() *petrinet.PetriNet {
	p := petrinet.Init(1)
	p.AddPlace(1, 1, "")
	p.AddPlace(2, 1, "")
	p.AddPlace(3, 2, "")
	p.AddPlace(4, 1, "")
	p.AddTransition(1,0)
	p.AddTransition(2,0)
	p.AddInArc(1,1,1)
	p.AddInArc(2,2,1)
	p.AddInArc(3,2,1)
	p.AddOutArc(1,2,1)
	p.AddOutArc(1,3,1)
	p.AddOutArc(2,4,1)
  //p.AddInhibitorArc(4,2,1)
	// fmt.Printf("%v\n", p)
	return p
}
