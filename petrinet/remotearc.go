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
	return fmt.Sprintf("{arc to service: %v, weight: %v}", rt.ServiceName, rt.Weight)
}

//@TODO
func (rt *RemoteArc) canFire() bool {
	return true
}

//@TODO
func (rt *RemoteArc) fire() {

}
