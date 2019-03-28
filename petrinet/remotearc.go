package petrinet

import (
  "fmt"
)

type RemoteArc struct {
  PlaceID int
  Context string
	Address string
	Weight int
  Marks int
}

func (a RemoteArc) String() string {
	return fmt.Sprintf("{placeID: %v, address: %v, weight: %v, marks: %v}", a.PlaceID, a.Address, a.Weight, a.Marks)
}
