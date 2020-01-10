package petrinet

import (
	"fmt"
)

// RemoteArc for arcs crossing nodes
type RemoteArc struct {
	ServiceName string
	Weight  int
}

func (rt RemoteArc) String() string {
	return fmt.Sprintf("{placeID: %v, address: %v, weight: %v, marks: %v}", rt.PlaceID, rt.Address, rt.Weight, rt.Marks)
}

//@TODO
func (rt *RemoteArc) canFire() bool {
	return true
}

//@TODO
func (rt *RemoteArc) fire() {

}
