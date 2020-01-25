package petrinet

import (
	"errors"
	"fmt"
)

// Transition of a PetriNet
type Transition struct {
	ID            int
	Priority      int
	InArcs        []Arc
	OutArcs       []Arc
	RemoteOutArcs []RemoteArc
}



func (t Transition) String() string {
	arcListString := func(list []Arc) string {
		ans := "["
		for _, item := range list {
			if ans != "[" {
				ans += ", "
			}
			ans += item.String()
		}
		ans += "]"
		return ans
	}
	return fmt.Sprintf(
		"{ID: %v, priority: %v, inArcs: %v, outArcs: %v}",
		t.ID, t.Priority, arcListString(t.InArcs), arcListString(t.OutArcs))
}

//CanFire checks if the transition can fire
func (t *Transition) CanFire() bool {
	ans := true
	for _, currArc := range t.InArcs {
		ans = ans && currArc.Place.GetNumMarks() >= currArc.Weight
	}
	for _, remArc := range t.RemoteOutArcs {
		ans = ans && remArc.canFire()
	}
	return ans
}

//Fire fires the transition
func (t *Transition) Fire() error {
	if !t.CanFire() {
		return errors.New("Trying to fire transition that can't be fired")
	}
	marks := []Token{}
	for _, currArc := range t.InArcs {
		//fmt.Printf("Fire 1 = %v \n", currArc.Place.Marks)
		marks = append(marks,currArc.Place.GetMark(currArc.Weight)...)
		//fmt.Printf("Fire 1!= %v \n", currArc.Place.Marks)
	}

	for _, currArc := range t.OutArcs {
		currArc.Place.AddMarks(marks[0:currArc.Weight-1])
	}
	
	for _, remArc := range t.RemoteOutArcs {
		//fmt.Printf("voy a disparar, remoto \n")
		remArc.fire(marks)
		//fmt.Printf("Disparo remoto, o no? \n")
	}
	return nil
}

//AddInArc adds an arc to \cdot t
func (t *Transition) AddInArc(_arc Arc) {
	t.InArcs = append(t.InArcs, _arc)
}

//AddOutArc internode arcs
func (t *Transition) AddOutArc(_arc Arc) {
	t.OutArcs = append(t.OutArcs, _arc)
}

//AddRemoteOutArc arcs crossing nodes, alwasys from transition to place
func (t *Transition) AddRemoteOutArc(_rarc RemoteArc) {
	_rarc.Init()
	t.RemoteOutArcs = append(t.RemoteOutArcs, _rarc)
}
