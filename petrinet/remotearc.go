package petrinet

import (
	"fmt"
)

// RemoteArc for arcs crossing nodes
type RemoteArc struct {
	PlaceID int
	Address string
	Weight  int
	Marks   int
}

func (rt RemoteArc) String() string {
	return fmt.Sprintf("{placeID: %v, address: %v, weight: %v, marks: %v}", a.placeID, a.address, a.weight, a.marks)
}

//@TODO
func (rt *RemoteArc) canFire() bool {
	return true
}

//@TODO
func (rt *RemoteArc) fire() {

}
