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
	return fmt.Sprintf("{placeID: %v, address: %v, context: %v, weight: %v, marks: %v}", a.PlaceID, a.Address, a.Context, a.Weight, a.Marks)
}
