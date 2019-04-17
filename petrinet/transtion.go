package petrinet

import (
	"fmt"
	"errors"
)

// Transition of a PetriNet
type Transition struct {
	ID int
	Priority int
	inArcs []arc
	outArcs []arc
	inhibitorArcs []arc
}

func (t Transition) String() string {
	arcListString := func(list []arc) string {
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
		"{ID: %v, priority: %v, inArcs: %v, outArcs: %v, inhibitorArcs: %v}",
		t.ID, t.Priority, arcListString(t.inArcs), arcListString(t.outArcs), arcListString(t.inhibitorArcs))
}


func (t *Transition) canFire() bool {
	ans := true
  for _, currArc := range t.inArcs {
		ans = ans && currArc.place.marks >= currArc.weight
  }
  for _, value := range t.inhibitorArcs {
    ans = ans && value.place.marks < value.weight
  }
  return ans
}

func (t *Transition) fire() error {
	if !t.canFire() {
		return errors.New("Trying to fire transition that can't be fired")
	}
	for _, currArc := range t.inArcs {
    currArc.place.marks -= currArc.weight
  }
  for _, currArc := range t.outArcs {
    currArc.place.marks += currArc.weight
  }
	return nil
}

func (t *Transition) addInArc(_arc arc) {
	t.inArcs = append(t.inArcs, _arc)
}

func (t *Transition) addOutArc(_arc arc) {
	t.outArcs = append(t.outArcs, _arc)
}

func (t *Transition) addInhibitorArc(_arc arc) {
	t.inhibitorArcs = append(t.inhibitorArcs, _arc)
}
