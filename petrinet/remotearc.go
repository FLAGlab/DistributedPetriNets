package petrinet

import (
	"fmt"
)

type RemoteArc struct {
	placeID int
	address string
	weight  int
	marks   int
}

func (a RemoteArc) String() string {
	return fmt.Sprintf("{placeID: %v, address: %v, weight: %v, marks: %v}", a.placeID, a.address, a.weight, a.marks)
}

//@TODO
func (t *RemoteArc) canFire() bool {
	return true
}

//@TODO
func (t *RemoteArc) fire() {

}
