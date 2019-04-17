package petribuilder

import (
  "github.com/FLAGlab/DCoPN/petrinet"
)

// BuildPetriNet builds a test petri net
func BuildPetriNet1() *petrinet.PetriNet {
	p := petrinet.Init(1, "ctx0")
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

// BuildPetriNet builds a test petri net
func BuildPetriNet2() *petrinet.PetriNet {
	p := petrinet.Init(2, "ctx1")
	p.AddPlace(1, 4, "")
	p.AddPlace(2, 0, "")
	p.AddTransition(1,0)
	p.AddInArc(1,1,1)
	p.AddOutArc(1,2,1)
  p.AddRemoteTransition(1)
  p.AddRemoteInArc(1, 1, 1, "ctx2")
  p.AddRemoteOutArc(1, 2, 1, "ctx2")
	return p
}

// BuildPetriNet builds a test petri net
func BuildPetriNet3() *petrinet.PetriNet {
	p := petrinet.Init(3, "ctx2")

	p.AddPlace(1, 4, "")
	p.AddPlace(2, 0, "")
	p.AddTransition(1, 0)
	p.AddInArc(1,1,1)
	p.AddOutArc(1,2,1)
  p.AddRemoteTransition(1)
  p.AddRemoteInArc(1, 1, 1, "ctx1")
  p.AddRemoteOutArc(1, 2, 1, "ctx1")
  return p
}

// BuildExperiment1 builds experiment 1 petri net
func BuildExperiment1(id int) *petrinet.PetriNet {
  p := petrinet.Init(id, "exp1")

  p.AddPlace(1,4,"")
  p.AddPlace(2,0,"")
  p.AddPlace(3,0,"")
  p.AddTransition(1, 1)
  p.AddTransition(3, 1)
  p.AddTransition(2, 0)
  p.AddTransition(4, 0)
  p.AddRemoteTransition(2)
  p.AddRemoteInhibitorArc(2, 2, 1, "exp1")
  p.AddInArc(1, 2, 1)
  p.AddInArc(3, 4, 1)
  p.AddInArc(2, 4, 1)
  p.AddOutArc(1, 1, 1)
  p.AddOutArc(2, 2, 1)
  p.AddOutArc(3, 3, 1)
  return p
}
