package petrinet

import (
	"errors"
	"fmt"
)

// Transition of a PetriNet
type Transition struct {
	ID            int
	priority      int
	inArcs        []arc
	outArcs       []arc
	remoteOutArcs []RemoteArc
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
		"{ID: %v, priority: %v, inArcs: %v, outArcs: %v}",
		t.ID, t.priority, arcListString(t.inArcs), arcListString(t.outArcs))
}

func (t *Transition) canFire() bool {
	ans := true
	for _, currArc := range t.inArcs {
		ans = ans && currArc.place.Marks >= currArc.weight
	}
	for _, remArc := range t.remoteOutArcs {
		ans = ans && remArc.canFire()
	}
	return ans
}

func (t *Transition) fire() error {
	if !t.canFire() {
		return errors.New("Trying to fire transition that can't be fired")
	}
	for _, currArc := range t.inArcs {
		currArc.place.Marks -= currArc.weight
	}
	for _, currArc := range t.outArcs {
		currArc.place.Marks += currArc.weight
	}
	for _, remArc := range t.remoteOutArcs {
		remArc.fire()
	}
	return nil
}

func (t *Transition) addInArc(_arc arc) {
	t.inArcs = append(t.inArcs, _arc)
}

func (t *Transition) addOutArc(_arc arc) {
	t.outArcs = append(t.outArcs, _arc)
}

func (t *Transition) addRemoteOutArc(_rarc RemoteArc) {
	t.remoteOutArcs = append(t.remoteOutArcs, _rarc)
}
